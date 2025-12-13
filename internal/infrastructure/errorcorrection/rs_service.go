// Package errorcorrection provides infrastructure implementations for the Error Correction domain.
package errorcorrection

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/klauspost/reedsolomon"

	"github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
)

// RSService implements ErrorCorrectionService using the klauspost/reedsolomon library.
type RSService struct {
	logger *logrus.Logger
}

// NewRSService creates a new Reed-Solomon service instance.
func NewRSService(logger *logrus.Logger) *RSService {
	return &RSService{
		logger: logger,
	}
}

// Encode encodes data into shards using Reed-Solomon encoding.
func (s *RSService) Encode(ctx context.Context, data []byte, config *errorcorrection.ShardConfiguration) (*errorcorrection.ProtectedMessage, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if len(data) == 0 {
		return nil, errorcorrection.ErrEmptyData
	}

	if config == nil {
		return nil, errorcorrection.ErrInvalidShardCount
	}

	s.logger.WithFields(logrus.Fields{
		"data_size":     len(data),
		"data_shards":   config.DataShards,
		"parity_shards": config.ParityShards,
		"redundancy":    config.Redundancy.String(),
	}).Info("Encoding data with Reed-Solomon")

	// Prepend the original data size (4 bytes, big-endian) for recovery
	originalSize := len(data)
	sizeBytes := make([]byte, 4)
	sizeBytes[0] = byte(originalSize >> 24)
	sizeBytes[1] = byte(originalSize >> 16)
	sizeBytes[2] = byte(originalSize >> 8)
	sizeBytes[3] = byte(originalSize)

	// Combine size prefix with data
	dataWithSize := append(sizeBytes, data...)

	// Create Reed-Solomon encoder
	enc, err := reedsolomon.New(config.DataShards, config.ParityShards)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to create Reed-Solomon encoder")
		return nil, fmt.Errorf("%w: %v", errorcorrection.ErrEncodingFailed, err)
	}

	// Split data into shards
	shardData, err := enc.Split(dataWithSize)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to split data into shards")
		return nil, fmt.Errorf("%w: failed to split data: %v", errorcorrection.ErrEncodingFailed, err)
	}

	// Encode parity shards
	if err := enc.Encode(shardData); err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to encode parity shards")
		return nil, fmt.Errorf("%w: failed to encode: %v", errorcorrection.ErrEncodingFailed, err)
	}

	// Create ProtectedMessage
	messageID := errorcorrection.GenerateMessageID()
	message, err := errorcorrection.NewProtectedMessage(
		messageID,
		data,
		config.DataShards,
		config.ParityShards,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create protected message: %w", err)
	}

	// Create Shard entities with checksums
	for i, shardBytes := range shardData {
		shardID := errorcorrection.GenerateShardID()
		isParity := i >= config.DataShards

		shard, err := errorcorrection.NewShard(
			shardID,
			messageID,
			i,
			shardBytes,
			isParity,
		)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"index": i,
				"error": err.Error(),
			}).Error("Failed to create shard")
			return nil, fmt.Errorf("failed to create shard %d: %w", i, err)
		}

		// Calculate and set checksum for integrity verification
		checksum := s.calculateChecksum(shardBytes)
		shard.SetChecksum(checksum)

		if err := message.AddShard(shard); err != nil {
			s.logger.WithFields(logrus.Fields{
				"index": i,
				"error": err.Error(),
			}).Error("Failed to add shard to message")
			return nil, fmt.Errorf("failed to add shard %d: %w", i, err)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"message_id":    messageID.String(),
		"total_shards":  len(shardData),
		"data_shards":   config.DataShards,
		"parity_shards": config.ParityShards,
	}).Info("Successfully encoded data")

	return message, nil
}

// Decode reconstructs the original data from available shards.
func (s *RSService) Decode(ctx context.Context, shards []*errorcorrection.Shard, config *errorcorrection.ShardConfiguration) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if len(shards) == 0 {
		return nil, errorcorrection.ErrInsufficientShards
	}

	if config == nil {
		return nil, errorcorrection.ErrInvalidShardCount
	}

	// Check if we have enough shards
	if len(shards) < config.DataShards {
		s.logger.WithFields(logrus.Fields{
			"available": len(shards),
			"required":  config.DataShards,
		}).Error("Insufficient shards for recovery")
		return nil, errorcorrection.ErrInsufficientShards
	}

	s.logger.WithFields(logrus.Fields{
		"available_shards": len(shards),
		"required_shards":  config.DataShards,
		"total_shards":     config.TotalShards(),
	}).Info("Decoding data from shards")

	// Create Reed-Solomon encoder (used for decoding too)
	enc, err := reedsolomon.New(config.DataShards, config.ParityShards)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to create Reed-Solomon encoder")
		return nil, fmt.Errorf("%w: %v", errorcorrection.ErrDecodingFailed, err)
	}

	// Create shard data array with nil placeholders for missing shards
	shardData := make([][]byte, config.TotalShards())
	presentShards := make(map[int]bool)

	for _, shard := range shards {
		if err := shard.Validate(); err != nil {
			s.logger.WithFields(logrus.Fields{
				"index": shard.Index,
				"error": err.Error(),
			}).Warn("Invalid shard detected")
			continue
		}

		// Verify checksum if present
		if len(shard.Checksum) > 0 {
			calculatedChecksum := s.calculateChecksum(shard.Data)
			if !shard.VerifyChecksum(calculatedChecksum) {
				s.logger.WithFields(logrus.Fields{
					"index":    shard.Index,
					"shard_id": shard.ID.String(),
				}).Warn("Shard checksum verification failed")
				return nil, fmt.Errorf("%w: shard %d", errorcorrection.ErrChecksumMismatch, shard.Index)
			}
		}

		if shard.Index >= 0 && shard.Index < config.TotalShards() {
			shardData[shard.Index] = shard.Data
			presentShards[shard.Index] = true
		}
	}

	// Log which shards are present and which are missing
	var missingIndices []int
	for i := 0; i < config.TotalShards(); i++ {
		if !presentShards[i] {
			missingIndices = append(missingIndices, i)
		}
	}

	if len(missingIndices) > 0 {
		s.logger.WithFields(logrus.Fields{
			"missing_count":   len(missingIndices),
			"missing_indices": missingIndices,
		}).Info("Reconstructing with missing shards")
	}

	// Verify we have enough shards
	if len(presentShards) < config.DataShards {
		s.logger.WithFields(logrus.Fields{
			"valid_shards": len(presentShards),
			"required":     config.DataShards,
		}).Error("Still insufficient shards after validation")
		return nil, errorcorrection.ErrInsufficientShards
	}

	// Reconstruct missing shards if necessary
	if len(missingIndices) > 0 {
		if err := enc.Reconstruct(shardData); err != nil {
			s.logger.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("Failed to reconstruct missing shards")
			return nil, fmt.Errorf("%w: reconstruction failed: %v", errorcorrection.ErrDecodingFailed, err)
		}

		s.logger.WithFields(logrus.Fields{
			"reconstructed": len(missingIndices),
		}).Info("Successfully reconstructed missing shards")
	}

	// Join data shards back together
	dataShards := shardData[:config.DataShards]

	// Use Join to reconstruct the data with proper padding handling
	var buf bytes.Buffer

	// Calculate total size of all data shards for Join
	totalShardSize := 0
	for _, shard := range dataShards {
		totalShardSize += len(shard)
	}

	err = enc.Join(&buf, dataShards, totalShardSize)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to join shards")
		return nil, fmt.Errorf("%w: join failed: %v", errorcorrection.ErrDecodingFailed, err)
	}

	recoveredData := buf.Bytes()

	// Extract the original size from the first 4 bytes
	if len(recoveredData) < 4 {
		return nil, fmt.Errorf("%w: insufficient data for size prefix", errorcorrection.ErrDecodingFailed)
	}

	originalSize := int(recoveredData[0])<<24 | int(recoveredData[1])<<16 |
		int(recoveredData[2])<<8 | int(recoveredData[3])

	// Extract the actual data (skip 4-byte size prefix and trim to original size)
	if len(recoveredData) < 4+originalSize {
		return nil, fmt.Errorf("%w: recovered data smaller than expected size", errorcorrection.ErrDecodingFailed)
	}

	actualData := recoveredData[4 : 4+originalSize]

	s.logger.WithFields(logrus.Fields{
		"recovered_size": len(actualData),
		"shards_used":    len(presentShards),
	}).Info("Successfully decoded data")

	return actualData, nil
}

// VerifyShards checks the integrity of all shards.
func (s *RSService) VerifyShards(ctx context.Context, shards []*errorcorrection.Shard) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if len(shards) == 0 {
		return errorcorrection.ErrInsufficientShards
	}

	s.logger.WithFields(logrus.Fields{
		"shard_count": len(shards),
	}).Info("Verifying shard integrity")

	corruptedCount := 0
	for _, shard := range shards {
		// Validate shard structure
		if err := shard.Validate(); err != nil {
			s.logger.WithFields(logrus.Fields{
				"index": shard.Index,
				"error": err.Error(),
			}).Warn("Shard validation failed")
			corruptedCount++
			continue
		}

		// Verify checksum if present
		if len(shard.Checksum) > 0 {
			calculatedChecksum := s.calculateChecksum(shard.Data)
			if !shard.VerifyChecksum(calculatedChecksum) {
				s.logger.WithFields(logrus.Fields{
					"index":    shard.Index,
					"shard_id": shard.ID.String(),
				}).Warn("Shard checksum mismatch")
				corruptedCount++
			}
		}
	}

	if corruptedCount > 0 {
		s.logger.WithFields(logrus.Fields{
			"corrupted_count": corruptedCount,
			"total_shards":    len(shards),
		}).Error("Shard verification completed with errors")
		return fmt.Errorf("%w: %d of %d shards corrupted", errorcorrection.ErrCorruptedShard, corruptedCount, len(shards))
	}

	s.logger.WithFields(logrus.Fields{
		"shard_count": len(shards),
	}).Info("All shards verified successfully")

	return nil
}

// CalculateCapacity determines the maximum data size for a given shard configuration.
func (s *RSService) CalculateCapacity(config *errorcorrection.ShardConfiguration, shardSize int) int {
	if config == nil || shardSize <= 0 {
		return 0
	}

	// Total capacity is data shards * shard size
	return config.DataShards * shardSize
}

// OptimizeConfiguration recommends optimal shard configuration for given data size.
func (s *RSService) OptimizeConfiguration(ctx context.Context, dataSize int, redundancyLevel errorcorrection.RedundancyLevel) (*errorcorrection.ShardConfiguration, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if dataSize <= 0 {
		return nil, errorcorrection.ErrEmptyData
	}

	s.logger.WithFields(logrus.Fields{
		"data_size":  dataSize,
		"redundancy": redundancyLevel.String(),
	}).Info("Optimizing shard configuration")

	// Start with default configuration
	dataShards := errorcorrection.DefaultDataShards
	parityShards := redundancyLevel.ParityShardCount(dataShards)

	// Adjust for very small data (< 10KB)
	if dataSize < 10*1024 {
		dataShards = 5
		parityShards = redundancyLevel.ParityShardCount(dataShards)
	}

	// Adjust for very large data (> 100MB)
	if dataSize > 100*1024*1024 {
		dataShards = 20
		parityShards = redundancyLevel.ParityShardCount(dataShards)
	}

	// Ensure we don't exceed maximum shards
	if dataShards+parityShards > errorcorrection.MaxTotalShards {
		dataShards = errorcorrection.MaxTotalShards - parityShards
		if dataShards <= 0 {
			return nil, errorcorrection.ErrTooManyShards
		}
	}

	config, err := errorcorrection.NewShardConfiguration(dataShards, parityShards)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to create optimized configuration")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"data_shards":   config.DataShards,
		"parity_shards": config.ParityShards,
		"redundancy":    config.Redundancy.String(),
		"max_failures":  config.MaximumFailures(),
	}).Info("Optimized configuration created")

	return config, nil
}

// calculateChecksum calculates SHA-256 checksum for shard data.
func (s *RSService) calculateChecksum(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
