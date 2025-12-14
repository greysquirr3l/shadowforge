// Package distribution provides infrastructure implementations for distribution services.
package distribution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/greysquirr3l/shadowforge/internal/domain/distribution"
	"github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// DistributionService implements the distribution domain service.
type DistributionService struct {
	ecService    errorcorrection.ErrorCorrectionService
	mediaService media.Service
	stegoService stego.StegoService
	repository   distribution.Repository
	logger       *logrus.Logger
}

// NewDistributionService creates a new distribution service.
func NewDistributionService(
	ecService errorcorrection.ErrorCorrectionService,
	mediaService media.Service,
	stegoService stego.StegoService,
	repository distribution.Repository,
	logger *logrus.Logger,
) *DistributionService {
	if logger == nil {
		logger = logrus.New()
	}

	return &DistributionService{
		ecService:    ecService,
		mediaService: mediaService,
		stegoService: stegoService,
		repository:   repository,
		logger:       logger,
	}
}

// CreateStrategy creates a new distribution strategy with validation.
func (s *DistributionService) CreateStrategy(
	ctx context.Context,
	pattern distribution.PatternType,
	dataShards, parityShards int,
	targetMediaIDs []string,
) (*distribution.DistributionStrategy, error) {
	s.logger.WithFields(logrus.Fields{
		"pattern":       pattern,
		"data_shards":   dataShards,
		"parity_shards": parityShards,
		"target_count":  len(targetMediaIDs),
	}).Info("Creating distribution strategy")

	// Create the strategy entity
	strategy, err := distribution.NewDistributionStrategy(
		pattern,
		dataShards,
		parityShards,
		targetMediaIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create strategy: %w", err)
	}

	// Validate the strategy
	if err := s.ValidateStrategy(ctx, strategy); err != nil {
		return nil, fmt.Errorf("strategy validation failed: %w", err)
	}

	// Persist the strategy
	if s.repository != nil {
		if err := s.repository.SaveStrategy(ctx, strategy); err != nil {
			return nil, fmt.Errorf("failed to save strategy: %w", err)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"strategy_id": strategy.ID.String(),
		"threshold":   strategy.Threshold,
		"redundancy":  fmt.Sprintf("%.1f%%", strategy.RedundancyLevel()*100),
	}).Info("Distribution strategy created successfully")

	return strategy, nil
}

// DistributeShards distributes shards according to strategy.
func (s *DistributionService) DistributeShards(
	ctx context.Context,
	strategyID distribution.StrategyID,
	shardData [][]byte,
) (*distribution.ShardManifest, error) {
	s.logger.WithFields(logrus.Fields{
		"strategy_id": strategyID.String(),
		"shard_count": len(shardData),
	}).Info("Distributing shards")

	// Retrieve the strategy
	var strategy *distribution.DistributionStrategy
	var err error
	if s.repository != nil {
		strategy, err = s.repository.GetStrategy(ctx, strategyID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve strategy: %w", err)
		}
	} else {
		return nil, fmt.Errorf("repository not available")
	}

	// Validate shard count
	if len(shardData) != strategy.TotalShards {
		return nil, fmt.Errorf("shard count mismatch: got %d, expected %d",
			len(shardData), strategy.TotalShards)
	}

	// Create shard metadata
	shardMetadata := make([]distribution.ShardMetadata, len(shardData))
	for i, shard := range shardData {
		// Calculate checksum
		hash := sha256.Sum256(shard)
		checksum := hex.EncodeToString(hash[:])

		// Get target media ID
		var mediaID string
		if i < len(strategy.TargetMediaIDs) {
			mediaID = strategy.TargetMediaIDs[i]
		} else {
			mediaID = fmt.Sprintf("media-%d", i)
		}

		metadata, err := distribution.NewShardMetadata(
			i,
			int64(len(shard)),
			checksum,
			mediaID,
			"", // RecipientID can be set later if needed
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create shard metadata for shard %d: %w", i, err)
		}
		shardMetadata[i] = metadata
	}

	// Create manifest
	manifest, err := distribution.NewShardManifest(
		strategyID,
		shardMetadata,
		strategy.Threshold,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest: %w", err)
	}

	// Update strategy with manifest
	if err := strategy.Complete(manifest); err != nil {
		return nil, fmt.Errorf("failed to complete strategy: %w", err)
	}

	// Persist updated strategy
	if s.repository != nil {
		if err := s.repository.SaveStrategy(ctx, strategy); err != nil {
			s.logger.WithError(err).Warn("Failed to persist strategy update")
		}
		if err := s.repository.SaveManifest(ctx, manifest); err != nil {
			s.logger.WithError(err).Warn("Failed to persist manifest")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"manifest_id":     manifest.ID.String(),
		"total_shards":    manifest.TotalShards,
		"required_shards": manifest.RequiredShards,
	}).Info("Shards distributed successfully")

	return manifest, nil
}

// ValidateStrategy validates a distribution strategy configuration.
func (s *DistributionService) ValidateStrategy(
	ctx context.Context,
	strategy *distribution.DistributionStrategy,
) error {
	// Basic validation (already done in entity, but double-check)
	if strategy.DataShards < 1 {
		return distribution.ErrInvalidDataShards
	}
	if strategy.ParityShards < 0 {
		return distribution.ErrInvalidParityShards
	}

	totalShards := strategy.DataShards + strategy.ParityShards
	if totalShards > len(strategy.TargetMediaIDs) {
		return fmt.Errorf("insufficient media: need %d, have %d",
			totalShards, len(strategy.TargetMediaIDs))
	}

	// Validate redundancy level is reasonable
	redundancy := strategy.RedundancyLevel()
	if redundancy > 2.0 { // More than 200% redundancy is excessive
		s.logger.WithField("redundancy", redundancy).Warn(
			"Unusually high redundancy level detected")
	}

	// Pattern-specific validation
	switch strategy.Pattern {
	case distribution.PatternOneToOne:
		if totalShards != 1 {
			return fmt.Errorf("one-to-one pattern requires exactly 1 shard, got %d", totalShards)
		}
	case distribution.PatternOneToMany:
		if totalShards < 2 {
			return fmt.Errorf("one-to-many pattern requires at least 2 shards, got %d", totalShards)
		}
	case distribution.PatternManyToOne:
		// Multiple payloads in one media - validate later with actual data
	case distribution.PatternManyToMany:
		// Complex validation - ensure we have enough media for all payload/shard combinations
	}

	return nil
}

// CalculateOptimalSharding calculates optimal data/parity shard configuration.
func (s *DistributionService) CalculateOptimalSharding(
	ctx context.Context,
	dataSize int64,
	targetCount int,
) (dataShards, parityShards int, err error) {
	s.logger.WithFields(logrus.Fields{
		"data_size":    dataSize,
		"target_count": targetCount,
	}).Debug("Calculating optimal sharding")

	if dataSize <= 0 {
		return 0, 0, fmt.Errorf("invalid data size: %d", dataSize)
	}
	if targetCount < 1 {
		return 0, 0, fmt.Errorf("invalid target count: %d", targetCount)
	}

	// For one-to-one, no sharding needed
	if targetCount == 1 {
		return 1, 0, nil
	}

	// Strategy: Use 60-70% for data shards, 30-40% for parity shards
	// This provides good redundancy while not wasting too much capacity
	// Start with a baseline: use 2/3 of targets for data, 1/3 for parity
	dataShards = (targetCount * 2) / 3
	if dataShards < 1 {
		dataShards = 1
	}
	parityShards = targetCount - dataShards

	// Ensure we have at least 1 parity shard for redundancy
	if parityShards < 1 && targetCount > 1 {
		dataShards = targetCount - 1
		parityShards = 1
	}

	// Optimize based on data size
	// For very small data, we can afford more redundancy
	if dataSize < 1024 { // Less than 1KB
		// Use 50/50 split for small data
		dataShards = targetCount / 2
		parityShards = targetCount - dataShards
	}

	// For very large data, prefer less redundancy to save space
	if dataSize > 100*1024*1024 { // More than 100MB
		// Use 75/25 split for large data
		dataShards = (targetCount * 3) / 4
		if dataShards < 1 {
			dataShards = 1
		}
		parityShards = targetCount - dataShards
	}

	s.logger.WithFields(logrus.Fields{
		"data_shards":   dataShards,
		"parity_shards": parityShards,
		"redundancy":    fmt.Sprintf("%.1f%%", float64(parityShards)/float64(dataShards)*100),
		"threshold":     fmt.Sprintf("%d of %d", dataShards, targetCount),
	}).Info("Optimal sharding calculated")

	return dataShards, parityShards, nil
}

// CheckRecoverability checks if data can be recovered with available shards.
func (s *DistributionService) CheckRecoverability(
	ctx context.Context,
	manifestID distribution.ManifestID,
	availableIndices []int,
) (bool, error) {
	if s.repository == nil {
		return false, fmt.Errorf("repository not available")
	}

	// Retrieve manifest
	manifest, err := s.repository.GetManifest(ctx, manifestID)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve manifest: %w", err)
	}

	// Check if manifest has expired
	if manifest.IsExpired() {
		s.logger.WithField("manifest_id", manifestID.String()).Warn("Manifest has expired")
		return false, fmt.Errorf("manifest has expired")
	}

	// Check if we have enough shards
	canRecover := len(availableIndices) >= manifest.RequiredShards

	s.logger.WithFields(logrus.Fields{
		"manifest_id":      manifestID.String(),
		"available_shards": len(availableIndices),
		"required_shards":  manifest.RequiredShards,
		"can_recover":      canRecover,
	}).Info("Recoverability check completed")

	return canRecover, nil
}

// RevokeDistribution revokes a distribution strategy.
func (s *DistributionService) RevokeDistribution(
	ctx context.Context,
	strategyID distribution.StrategyID,
) error {
	if s.repository == nil {
		return fmt.Errorf("repository not available")
	}

	// Retrieve strategy
	strategy, err := s.repository.GetStrategy(ctx, strategyID)
	if err != nil {
		return fmt.Errorf("failed to retrieve strategy: %w", err)
	}

	// Update status to revoked
	strategy.Status = distribution.StatusRevoked

	// Persist
	if err := s.repository.SaveStrategy(ctx, strategy); err != nil {
		return fmt.Errorf("failed to save revoked strategy: %w", err)
	}

	s.logger.WithField("strategy_id", strategyID.String()).Info("Distribution strategy revoked")

	return nil
}

// GetDistributionStatus retrieves current distribution status.
func (s *DistributionService) GetDistributionStatus(
	ctx context.Context,
	strategyID distribution.StrategyID,
) (distribution.DistributionStatus, error) {
	if s.repository == nil {
		return "", fmt.Errorf("repository not available")
	}

	strategy, err := s.repository.GetStrategy(ctx, strategyID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve strategy: %w", err)
	}

	return strategy.Status, nil
}

// CalculateCapacityAwareDistribution calculates shard allocation based on media capacities.
//
// This is the key algorithm for capacity-aware distribution. It analyzes each target
// media's available capacity and allocates shards accordingly, ensuring:
// 1. Shards fit within media capacity
// 2. Load is balanced across media when possible
// 3. Larger capacity media receive proportionally more data
func (s *DistributionService) CalculateCapacityAwareDistribution(
	ctx context.Context,
	targetMediaData [][]byte,
	techniques []stego.StegoTechnique,
	totalDataSize int64,
	dataShards int,
) (allocation []ShardAllocation, err error) {
	s.logger.WithFields(logrus.Fields{
		"target_count":    len(targetMediaData),
		"total_data_size": totalDataSize,
		"data_shards":     dataShards,
	}).Info("Calculating capacity-aware shard allocation")

	if len(targetMediaData) == 0 {
		return nil, fmt.Errorf("no target media provided")
	}
	if len(targetMediaData) != len(techniques) {
		return nil, fmt.Errorf("media count (%d) must match technique count (%d)",
			len(targetMediaData), len(techniques))
	}

	// Calculate capacity for each media
	capacities := make([]MediaCapacity, len(targetMediaData))
	totalCapacity := int64(0)
	for i, mediaData := range targetMediaData {
		capacity, err := s.stegoService.CalculateCapacity(ctx, mediaData, techniques[i])
		if err != nil {
			s.logger.WithError(err).WithField("media_index", i).Warn("Failed to calculate capacity")
			capacities[i] = MediaCapacity{
				Index:    i,
				Capacity: 0,
				Valid:    false,
			}
			continue
		}
		capacities[i] = MediaCapacity{
			Index:     i,
			Capacity:  capacity,
			Technique: techniques[i],
			Valid:     capacity > 0,
		}
		totalCapacity += capacity
	}

	// Check if total capacity is sufficient
	shardSize := totalDataSize / int64(dataShards)
	if totalCapacity < totalDataSize {
		return nil, fmt.Errorf("insufficient total capacity: need %d bytes, have %d bytes",
			totalDataSize, totalCapacity)
	}

	// Allocate shards proportionally based on capacity
	allocation = make([]ShardAllocation, len(capacities))
	remainingShards := dataShards
	for i, cap := range capacities {
		if !cap.Valid || cap.Capacity == 0 {
			allocation[i] = ShardAllocation{
				MediaIndex:  i,
				ShardCount:  0,
				Capacity:    0,
				Utilization: 0,
			}
			continue
		}

		// Calculate proportional shard count
		proportion := float64(cap.Capacity) / float64(totalCapacity)
		shardCount := int(float64(dataShards) * proportion)

		// Ensure at least one shard if capacity allows
		if shardCount == 0 && cap.Capacity >= shardSize && remainingShards > 0 {
			shardCount = 1
		}

		// Don't exceed remaining shards
		if shardCount > remainingShards {
			shardCount = remainingShards
		}

		allocation[i] = ShardAllocation{
			MediaIndex:  i,
			ShardCount:  shardCount,
			Capacity:    cap.Capacity,
			Utilization: float64(shardCount*int(shardSize)) / float64(cap.Capacity),
		}
		remainingShards -= shardCount
	}

	// Distribute any remaining shards to media with lowest utilization
	for remainingShards > 0 {
		bestIdx := -1
		lowestUtilization := 2.0 // > 100%
		for i, alloc := range allocation {
			if capacities[i].Valid && alloc.Utilization < lowestUtilization {
				// Check if this media can handle one more shard
				newUtilization := float64((alloc.ShardCount+1)*int(shardSize)) / float64(alloc.Capacity)
				if newUtilization <= 0.9 { // Don't exceed 90% utilization
					lowestUtilization = alloc.Utilization
					bestIdx = i
				}
			}
		}

		if bestIdx == -1 {
			// Can't allocate remaining shards without exceeding capacity
			break
		}

		allocation[bestIdx].ShardCount++
		allocation[bestIdx].Utilization = float64(allocation[bestIdx].ShardCount*int(shardSize)) /
			float64(allocation[bestIdx].Capacity)
		remainingShards--
	}

	if remainingShards > 0 {
		return nil, fmt.Errorf("could not allocate all shards: %d remaining", remainingShards)
	}

	s.logger.WithFields(logrus.Fields{
		"allocations":            len(allocation),
		"total_shards_allocated": dataShards,
	}).Info("Capacity-aware allocation completed")

	return allocation, nil
}

// MediaCapacity represents the capacity of a single media file.
type MediaCapacity struct {
	Index     int
	Capacity  int64
	Technique stego.StegoTechnique
	Valid     bool
}

// ShardAllocation represents shard allocation for a specific media.
type ShardAllocation struct {
	MediaIndex  int
	ShardCount  int
	Capacity    int64
	Utilization float64 // Percentage (0.0-1.0)
}
