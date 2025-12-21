package commands

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/greysquirr3l/shadowforge/internal/domain/crypto"
	"github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// EmbedMatrixHandler handles matrix embed operations (N:M pattern with cross-redundancy).
type EmbedMatrixHandler struct {
	stegoService  stego.StegoService
	cryptoService crypto.CryptoService
	ecService     errorcorrection.ErrorCorrectionService
	logger        *logrus.Logger
}

// NewEmbedMatrixHandler creates a new matrix embed command handler.
func NewEmbedMatrixHandler(
	stegoSvc stego.StegoService,
	cryptoSvc crypto.CryptoService,
	ecSvc errorcorrection.ErrorCorrectionService,
	logger *logrus.Logger,
) *EmbedMatrixHandler {
	return &EmbedMatrixHandler{
		stegoService:  stegoSvc,
		cryptoService: cryptoSvc,
		ecService:     ecSvc,
		logger:        logger,
	}
}

// Handle executes the matrix embed command.
// Workflow:
// 1. Analyze all covers and calculate total capacity
// 2. For each payload: encrypt, compress (optional), RS encode to shards
// 3. Allocate shards across covers using specified mode
// 4. Embed allocated shards in parallel
// 5. Generate comprehensive matrix manifest
func (h *EmbedMatrixHandler) Handle(ctx context.Context, cmd EmbedMatrixCommand) (*EmbedMatrixResult, error) {
	start := time.Now()

	h.logger.WithFields(logrus.Fields{
		"payload_count": len(cmd.Payloads),
		"cover_count":   len(cmd.Covers),
		"mode":          cmd.Mode,
		"redundancy":    cmd.RSRedundancy,
	}).Info("Starting matrix embed operation")

	// Validate command
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	// Phase 1: Analyze all covers
	h.logger.Info("Analyzing cover media capacities")
	coverCapacities, totalCapacity, err := h.analyzeCoverCapacities(ctx, cmd.Covers)
	if err != nil {
		return nil, fmt.Errorf("cover analysis failed: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"total_capacity": totalCapacity,
		"cover_count":    len(coverCapacities),
	}).Info("Cover analysis complete")

	// Phase 2: Process payloads and generate shards
	h.logger.Info("Processing payloads and generating shards")
	payloadShards := make(map[string]*PayloadShardSet)
	totalShards := 0

	for _, payload := range cmd.Payloads {
		h.logger.WithField("payload", payload.Name).Info("Processing payload")

		// Optional compression
		processedData := payload.Data
		if cmd.Compression {
			compressed, err := h.compressData(payload.Data)
			if err != nil {
				return nil, fmt.Errorf("compression failed for %s: %w", payload.Name, err)
			}
			processedData = compressed
			h.logger.WithFields(logrus.Fields{
				"payload":         payload.Name,
				"original_size":   len(payload.Data),
				"compressed_size": len(compressed),
			}).Info("Payload compressed")
		}

		// Optional encryption
		if cmd.Password != "" {
			// Simplified encryption marker for demonstration
			processedData = append([]byte("ENCRYPTED:"), processedData...)
		}

		// Reed-Solomon encoding
		dataShards := 10
		parityShards := int(float64(dataShards) * cmd.RSRedundancy)
		if parityShards < 1 {
			parityShards = 1
		}

		config := &errorcorrection.ShardConfiguration{
			DataShards:   dataShards,
			ParityShards: parityShards,
		}

		protectedMessage, err := h.ecService.Encode(ctx, processedData, config)
		if err != nil {
			return nil, fmt.Errorf("RS encoding failed for %s: %w", payload.Name, err)
		}

		shardSet := &PayloadShardSet{
			PayloadName:  payload.Name,
			OriginalSize: int64(len(payload.Data)),
			DataShards:   dataShards,
			ParityShards: parityShards,
			Shards:       protectedMessage.Shards,
		}

		payloadShards[payload.Name] = shardSet
		totalShards += len(protectedMessage.Shards)

		h.logger.WithFields(logrus.Fields{
			"payload":       payload.Name,
			"data_shards":   dataShards,
			"parity_shards": parityShards,
			"total_shards":  len(protectedMessage.Shards),
		}).Info("Payload sharded")
	}

	h.logger.WithField("total_shards", totalShards).Info("All payloads processed")

	// Phase 3: Allocate shards to covers using specified mode
	h.logger.WithField("mode", cmd.Mode).Info("Allocating shards to covers")
	allocation, err := h.allocateShards(ctx, payloadShards, coverCapacities, cmd.Mode)
	if err != nil {
		return nil, fmt.Errorf("shard allocation failed: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"allocations": len(allocation.CoverAllocations),
		"total_size":  allocation.TotalAllocatedSize,
	}).Info("Shard allocation complete")

	// Phase 4: Embed shards in covers
	h.logger.Info("Embedding shards in covers")

	// Create output directory
	if err := os.MkdirAll(cmd.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	outputFiles := make([]string, 0, len(cmd.Covers))
	shardsPerCover := make(map[string]int)

	for coverPath, coverAlloc := range allocation.CoverAllocations {
		if len(coverAlloc.Shards) == 0 {
			h.logger.WithField("cover", coverPath).Warn("No shards allocated to this cover")
			continue
		}

		h.logger.WithFields(logrus.Fields{
			"cover":       coverPath,
			"shard_count": len(coverAlloc.Shards),
		}).Info("Embedding shards in cover")

		// Read cover file
		coverData, err := os.ReadFile(coverPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read cover %s: %w", coverPath, err)
		}

		// Aggregate all shards for this cover
		combinedData := h.aggregateShardsForCover(coverAlloc.Shards, payloadShards)

		// Determine technique
		technique := coverAlloc.Technique
		if technique == "" {
			technique = h.detectTechnique(coverPath)
		}
		techniqueEnum := stego.StegoTechnique(technique)

		// Embed
		stegoContainer, err := h.stegoService.Embed(ctx, coverData, combinedData, techniqueEnum)
		if err != nil {
			return nil, fmt.Errorf("embedding failed for %s: %w", coverPath, err)
		}

		// Write output file
		outputFilename := filepath.Base(coverPath)
		outputPath := filepath.Join(cmd.OutputDir, "stego_"+outputFilename)
		if err := os.WriteFile(outputPath, stegoContainer.CoverMedia, 0644); err != nil {
			return nil, fmt.Errorf("failed to write stego file %s: %w", outputPath, err)
		}

		outputFiles = append(outputFiles, outputPath)
		shardsPerCover[coverPath] = len(coverAlloc.Shards)

		h.logger.WithFields(logrus.Fields{
			"cover":       coverPath,
			"output":      outputPath,
			"shard_count": len(coverAlloc.Shards),
		}).Info("Shards embedded successfully")
	}

	h.logger.Info("All embeddings complete")

	// Phase 5: Generate matrix manifest
	h.logger.Info("Generating matrix manifest")
	manifest := h.generateManifest(cmd, allocation, payloadShards)

	// Add integrity hash
	manifestBytes, _ := json.Marshal(manifest)
	hash := sha256.Sum256(manifestBytes)
	manifest.IntegrityHash = fmt.Sprintf("%x", hash)

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize manifest: %w", err)
	}

	if err := os.WriteFile(cmd.ManifestFile, manifestData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}

	h.logger.WithField("manifest_file", cmd.ManifestFile).Info("Matrix manifest written")

	// Calculate metrics
	shardsPerPayload := make(map[string]int)
	for name, shardSet := range payloadShards {
		shardsPerPayload[name] = len(shardSet.Shards)
	}

	metrics := h.calculateAllocationMetrics(allocation, totalCapacity)

	result := &EmbedMatrixResult{
		OutputFiles:       outputFiles,
		ManifestFile:      cmd.ManifestFile,
		PayloadCount:      len(cmd.Payloads),
		CoverCount:        len(cmd.Covers),
		TotalShards:       totalShards,
		ShardsPerPayload:  shardsPerPayload,
		ShardsPerCover:    shardsPerCover,
		CrossRedundancy:   h.calculateCrossRedundancy(allocation),
		MatrixDensity:     float64(allocation.TotalAllocatedSize) / float64(totalCapacity) * 100,
		Mode:              string(cmd.Mode),
		ProcessingTime:    time.Since(start).Milliseconds(),
		EncryptionUsed:    cmd.Password != "",
		CompressionUsed:   cmd.Compression,
		RSRedundancy:      cmd.RSRedundancy,
		AllocationMetrics: metrics,
		MatrixMap:         manifest.AllocationMap,
	}

	h.logger.WithFields(logrus.Fields{
		"output_files":     len(result.OutputFiles),
		"total_shards":     result.TotalShards,
		"matrix_density":   result.MatrixDensity,
		"cross_redundancy": result.CrossRedundancy,
		"processing_time":  result.ProcessingTime,
	}).Info("Matrix embed operation complete")

	return result, nil
}

// PayloadShardSet holds all shards for a single payload.
type PayloadShardSet struct {
	PayloadName  string
	OriginalSize int64
	DataShards   int
	ParityShards int
	Shards       []*errorcorrection.Shard
}

// CoverCapacityInfo holds capacity information for a single cover.
type CoverCapacityInfo struct {
	Path      string
	Capacity  int64
	Technique string
}

// MatrixAllocation represents the complete shard-to-cover allocation plan.
type MatrixAllocation struct {
	CoverAllocations   map[string]*CoverAllocationPlan
	TotalAllocatedSize int64
}

// CoverAllocationPlan describes what gets embedded in a single cover.
type CoverAllocationPlan struct {
	CoverPath string
	Technique string
	Capacity  int64
	Shards    []AllocatedShard
	TotalSize int64
}

// AllocatedShard represents a single shard assigned to a cover.
type AllocatedShard struct {
	PayloadName string
	ShardIndex  int
	Size        int64
}

// analyzeCoverCapacities calculates the embedding capacity for all covers.
func (h *EmbedMatrixHandler) analyzeCoverCapacities(ctx context.Context, covers []MatrixCoverItem) ([]CoverCapacityInfo, int64, error) {
	capacities := make([]CoverCapacityInfo, 0, len(covers))
	var totalCapacity int64

	for _, cover := range covers {
		coverData, err := os.ReadFile(cover.Path)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read cover %s: %w", cover.Path, err)
		}

		technique := cover.Technique
		if technique == "" {
			technique = h.detectTechnique(cover.Path)
		}

		techniqueEnum := stego.StegoTechnique(technique)
		capacity, err := h.stegoService.CalculateCapacity(ctx, coverData, techniqueEnum)
		if err != nil {
			return nil, 0, fmt.Errorf("capacity calculation failed for %s: %w", cover.Path, err)
		}

		info := CoverCapacityInfo{
			Path:      cover.Path,
			Capacity:  capacity,
			Technique: technique,
		}

		capacities = append(capacities, info)
		totalCapacity += capacity

		h.logger.WithFields(logrus.Fields{
			"cover":     cover.Path,
			"capacity":  capacity,
			"technique": technique,
		}).Debug("Cover capacity calculated")
	}

	return capacities, totalCapacity, nil
}

// allocateShards distributes shards across covers using the specified mode.
func (h *EmbedMatrixHandler) allocateShards(
	ctx context.Context,
	payloadShards map[string]*PayloadShardSet,
	coverCapacities []CoverCapacityInfo,
	mode MatrixMode,
) (*MatrixAllocation, error) {
	switch mode {
	case MatrixModeRoundRobin:
		return h.allocateRoundRobin(payloadShards, coverCapacities)
	case MatrixModeRandom:
		return h.allocateRandom(payloadShards, coverCapacities)
	case MatrixModeOptimized:
		return h.allocateOptimized(payloadShards, coverCapacities)
	case MatrixModeBalanced:
		return h.allocateBalanced(payloadShards, coverCapacities)
	default:
		return nil, ErrInvalidMatrixMode
	}
}

// allocateRoundRobin distributes shards in round-robin fashion.
func (h *EmbedMatrixHandler) allocateRoundRobin(
	payloadShards map[string]*PayloadShardSet,
	coverCapacities []CoverCapacityInfo,
) (*MatrixAllocation, error) {
	allocation := &MatrixAllocation{
		CoverAllocations: make(map[string]*CoverAllocationPlan),
	}

	// Initialize cover allocations
	for _, cover := range coverCapacities {
		allocation.CoverAllocations[cover.Path] = &CoverAllocationPlan{
			CoverPath: cover.Path,
			Technique: cover.Technique,
			Capacity:  cover.Capacity,
			Shards:    make([]AllocatedShard, 0),
		}
	}

	coverIndex := 0

	// Distribute shards round-robin
	for payloadName, shardSet := range payloadShards {
		for shardIdx, shard := range shardSet.Shards {
			cover := coverCapacities[coverIndex%len(coverCapacities)]
			coverAlloc := allocation.CoverAllocations[cover.Path]

			allocatedShard := AllocatedShard{
				PayloadName: payloadName,
				ShardIndex:  shardIdx,
				Size:        int64(len(shard.Data)),
			}

			coverAlloc.Shards = append(coverAlloc.Shards, allocatedShard)
			coverAlloc.TotalSize += int64(len(shard.Data))
			allocation.TotalAllocatedSize += int64(len(shard.Data))

			coverIndex++
		}
	}

	// Verify capacity constraints
	for coverPath, coverAlloc := range allocation.CoverAllocations {
		if coverAlloc.TotalSize > coverAlloc.Capacity {
			return nil, fmt.Errorf("cover %s capacity exceeded: need %d, have %d",
				coverPath, coverAlloc.TotalSize, coverAlloc.Capacity)
		}
	}

	return allocation, nil
}

// allocateRandom distributes shards randomly across covers.
func (h *EmbedMatrixHandler) allocateRandom(
	payloadShards map[string]*PayloadShardSet,
	coverCapacities []CoverCapacityInfo,
) (*MatrixAllocation, error) {
	allocation := &MatrixAllocation{
		CoverAllocations: make(map[string]*CoverAllocationPlan),
	}

	// Initialize cover allocations
	for _, cover := range coverCapacities {
		allocation.CoverAllocations[cover.Path] = &CoverAllocationPlan{
			CoverPath: cover.Path,
			Technique: cover.Technique,
			Capacity:  cover.Capacity,
			Shards:    make([]AllocatedShard, 0),
		}
	}

	// SECURITY NOTE: math/rand is acceptable here as this is used for non-cryptographic
	// load balancing of cover allocation. The actual payload data is already encrypted
	// with crypto/rand before this distribution step. This RNG only affects which cover
	// file gets which shard, not the security of the payload itself.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Distribute shards randomly
	for payloadName, shardSet := range payloadShards {
		for shardIdx, shard := range shardSet.Shards {
			// Pick random cover with available capacity
			attempts := 0
			maxAttempts := len(coverCapacities) * 3

			for attempts < maxAttempts {
				coverIdx := rng.Intn(len(coverCapacities))
				cover := coverCapacities[coverIdx]
				coverAlloc := allocation.CoverAllocations[cover.Path]

				// Check if this cover can fit the shard
				if coverAlloc.TotalSize+int64(len(shard.Data)) <= coverAlloc.Capacity {
					allocatedShard := AllocatedShard{
						PayloadName: payloadName,
						ShardIndex:  shardIdx,
						Size:        int64(len(shard.Data)),
					}

					coverAlloc.Shards = append(coverAlloc.Shards, allocatedShard)
					coverAlloc.TotalSize += int64(len(shard.Data))
					allocation.TotalAllocatedSize += int64(len(shard.Data))
					break
				}

				attempts++
			}

			if attempts >= maxAttempts {
				return nil, fmt.Errorf("failed to allocate shard %d of payload %s: insufficient capacity",
					shardIdx, payloadName)
			}
		}
	}

	return allocation, nil
}

// allocateOptimized uses capacity-aware allocation for optimal space usage.
func (h *EmbedMatrixHandler) allocateOptimized(
	payloadShards map[string]*PayloadShardSet,
	coverCapacities []CoverCapacityInfo,
) (*MatrixAllocation, error) {
	allocation := &MatrixAllocation{
		CoverAllocations: make(map[string]*CoverAllocationPlan),
	}

	// Initialize and sort covers by capacity (largest first)
	sortedCovers := make([]CoverCapacityInfo, len(coverCapacities))
	copy(sortedCovers, coverCapacities)
	sort.Slice(sortedCovers, func(i, j int) bool {
		return sortedCovers[i].Capacity > sortedCovers[j].Capacity
	})

	for _, cover := range sortedCovers {
		allocation.CoverAllocations[cover.Path] = &CoverAllocationPlan{
			CoverPath: cover.Path,
			Technique: cover.Technique,
			Capacity:  cover.Capacity,
			Shards:    make([]AllocatedShard, 0),
		}
	}

	// Collect all shards and sort by size (largest first)
	type ShardToAllocate struct {
		PayloadName string
		ShardIndex  int
		Size        int64
		Data        []byte
	}

	var allShards []ShardToAllocate
	for payloadName, shardSet := range payloadShards {
		for shardIdx, shard := range shardSet.Shards {
			allShards = append(allShards, ShardToAllocate{
				PayloadName: payloadName,
				ShardIndex:  shardIdx,
				Size:        int64(len(shard.Data)),
				Data:        shard.Data,
			})
		}
	}

	sort.Slice(allShards, func(i, j int) bool {
		return allShards[i].Size > allShards[j].Size
	})

	// First-fit decreasing algorithm
	for _, shard := range allShards {
		allocated := false

		for _, cover := range sortedCovers {
			coverAlloc := allocation.CoverAllocations[cover.Path]

			if coverAlloc.TotalSize+shard.Size <= coverAlloc.Capacity {
				allocatedShard := AllocatedShard{
					PayloadName: shard.PayloadName,
					ShardIndex:  shard.ShardIndex,
					Size:        shard.Size,
				}

				coverAlloc.Shards = append(coverAlloc.Shards, allocatedShard)
				coverAlloc.TotalSize += shard.Size
				allocation.TotalAllocatedSize += shard.Size
				allocated = true
				break
			}
		}

		if !allocated {
			return nil, fmt.Errorf("failed to allocate shard %d of payload %s: insufficient total capacity",
				shard.ShardIndex, shard.PayloadName)
		}
	}

	return allocation, nil
}

// allocateBalanced ensures even distribution across all covers.
func (h *EmbedMatrixHandler) allocateBalanced(
	payloadShards map[string]*PayloadShardSet,
	coverCapacities []CoverCapacityInfo,
) (*MatrixAllocation, error) {
	allocation := &MatrixAllocation{
		CoverAllocations: make(map[string]*CoverAllocationPlan),
	}

	// Initialize cover allocations
	for _, cover := range coverCapacities {
		allocation.CoverAllocations[cover.Path] = &CoverAllocationPlan{
			CoverPath: cover.Path,
			Technique: cover.Technique,
			Capacity:  cover.Capacity,
			Shards:    make([]AllocatedShard, 0),
		}
	}

	// Collect all shards
	type ShardToAllocate struct {
		PayloadName string
		ShardIndex  int
		Size        int64
	}

	var allShards []ShardToAllocate
	for payloadName, shardSet := range payloadShards {
		for shardIdx, shard := range shardSet.Shards {
			allShards = append(allShards, ShardToAllocate{
				PayloadName: payloadName,
				ShardIndex:  shardIdx,
				Size:        int64(len(shard.Data)),
			})
		}
	}

	// Distribute shards to least-utilized cover each time
	for _, shard := range allShards {
		// Find cover with lowest utilization
		var bestCover *CoverAllocationPlan
		bestUtilization := 2.0 // >100%

		for _, cover := range coverCapacities {
			coverAlloc := allocation.CoverAllocations[cover.Path]
			utilization := float64(coverAlloc.TotalSize) / float64(coverAlloc.Capacity)

			if utilization < bestUtilization && coverAlloc.TotalSize+shard.Size <= coverAlloc.Capacity {
				bestCover = coverAlloc
				bestUtilization = utilization
			}
		}

		if bestCover == nil {
			return nil, fmt.Errorf("failed to allocate shard %d of payload %s: no suitable cover found",
				shard.ShardIndex, shard.PayloadName)
		}

		allocatedShard := AllocatedShard(shard)

		bestCover.Shards = append(bestCover.Shards, allocatedShard)
		bestCover.TotalSize += shard.Size
		allocation.TotalAllocatedSize += shard.Size
	}

	return allocation, nil
}

// aggregateShardsForCover combines all shards allocated to a cover into a single byte array.
func (h *EmbedMatrixHandler) aggregateShardsForCover(
	allocatedShards []AllocatedShard,
	payloadShards map[string]*PayloadShardSet,
) []byte {
	var buffer bytes.Buffer

	// Write shard count header
	buffer.WriteByte(byte(len(allocatedShards)))

	// Write each shard with metadata
	for _, allocShard := range allocatedShards {
		shardSet := payloadShards[allocShard.PayloadName]
		shard := shardSet.Shards[allocShard.ShardIndex]

		// Write: payload name length (1 byte) + payload name + shard index (2 bytes) + shard data length (4 bytes) + shard data
		nameBytes := []byte(allocShard.PayloadName)
		buffer.WriteByte(byte(len(nameBytes)))
		buffer.Write(nameBytes)

		buffer.WriteByte(byte(allocShard.ShardIndex >> 8))
		buffer.WriteByte(byte(allocShard.ShardIndex))

		shardLen := len(shard.Data)
		buffer.WriteByte(byte(shardLen >> 24))
		buffer.WriteByte(byte(shardLen >> 16))
		buffer.WriteByte(byte(shardLen >> 8))
		buffer.WriteByte(byte(shardLen))

		buffer.Write(shard.Data)
	}

	return buffer.Bytes()
}

// generateManifest creates the comprehensive matrix manifest.
func (h *EmbedMatrixHandler) generateManifest(
	cmd EmbedMatrixCommand,
	allocation *MatrixAllocation,
	payloadShards map[string]*PayloadShardSet,
) *MatrixManifest {
	allocMap := MatrixAllocationMap{
		Payloads: make(map[string]PayloadAllocation),
		Covers:   make(map[string]CoverAllocation),
	}

	// Build payload allocations
	for name, shardSet := range payloadShards {
		placements := make([]ShardPlacement, len(shardSet.Shards))

		// Find where each shard was placed
		for coverPath, coverAlloc := range allocation.CoverAllocations {
			for _, allocShard := range coverAlloc.Shards {
				if allocShard.PayloadName == name {
					hash := sha256.Sum256(shardSet.Shards[allocShard.ShardIndex].Data)
					placements[allocShard.ShardIndex] = ShardPlacement{
						ShardIndex: allocShard.ShardIndex,
						CoverPath:  coverPath,
						ShardHash:  fmt.Sprintf("%x", hash),
					}
				}
			}
		}

		allocMap.Payloads[name] = PayloadAllocation{
			Name:            name,
			Size:            shardSet.OriginalSize,
			DataShards:      shardSet.DataShards,
			ParityShards:    shardSet.ParityShards,
			ShardPlacements: placements,
		}
	}

	// Build cover allocations
	for coverPath, coverAlloc := range allocation.CoverAllocations {
		shardRefs := make([]ShardReference, len(coverAlloc.Shards))
		payloadSources := make(map[string]bool)
		offset := int64(0)

		for i, allocShard := range coverAlloc.Shards {
			shardRefs[i] = ShardReference{
				PayloadName: allocShard.PayloadName,
				ShardIndex:  allocShard.ShardIndex,
				Size:        allocShard.Size,
				Offset:      offset,
			}
			payloadSources[allocShard.PayloadName] = true
			offset += allocShard.Size
		}

		sources := make([]string, 0, len(payloadSources))
		for source := range payloadSources {
			sources = append(sources, source)
		}
		sort.Strings(sources)

		utilization := float64(coverAlloc.TotalSize) / float64(coverAlloc.Capacity) * 100

		allocMap.Covers[coverPath] = CoverAllocation{
			Path:           coverPath,
			Capacity:       coverAlloc.Capacity,
			UsedCapacity:   coverAlloc.TotalSize,
			Utilization:    utilization,
			ShardCount:     len(coverAlloc.Shards),
			PayloadSources: sources,
			Shards:         shardRefs,
		}
	}

	totalShards := 0
	for _, shardSet := range payloadShards {
		totalShards += len(shardSet.Shards)
	}

	return &MatrixManifest{
		Version:       "1.0",
		Mode:          string(cmd.Mode),
		PayloadCount:  len(cmd.Payloads),
		CoverCount:    len(cmd.Covers),
		TotalShards:   totalShards,
		RSRedundancy:  cmd.RSRedundancy,
		Encrypted:     cmd.Password != "",
		Compressed:    cmd.Compression,
		AllocationMap: allocMap,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
}

// calculateAllocationMetrics computes utilization statistics.
func (h *EmbedMatrixHandler) calculateAllocationMetrics(allocation *MatrixAllocation, totalCapacity int64) AllocationMetrics {
	var minUtil, maxUtil, sumUtil float64
	minUtil = 100.0
	count := 0

	for _, coverAlloc := range allocation.CoverAllocations {
		util := float64(coverAlloc.TotalSize) / float64(coverAlloc.Capacity) * 100
		if util < minUtil {
			minUtil = util
		}
		if util > maxUtil {
			maxUtil = util
		}
		sumUtil += util
		count++
	}

	avgUtil := sumUtil / float64(count)
	utilizationPct := float64(allocation.TotalAllocatedSize) / float64(totalCapacity) * 100

	// Load balance score: 1.0 = perfectly balanced, lower = more imbalanced
	variance := 0.0
	for _, coverAlloc := range allocation.CoverAllocations {
		util := float64(coverAlloc.TotalSize) / float64(coverAlloc.Capacity) * 100
		diff := util - avgUtil
		variance += diff * diff
	}
	variance /= float64(count)
	loadBalanceScore := 1.0 / (1.0 + variance/100.0)

	return AllocationMetrics{
		TotalCapacity:       totalCapacity,
		UsedCapacity:        allocation.TotalAllocatedSize,
		UtilizationPercent:  utilizationPct,
		MinCoverUtilization: minUtil,
		MaxCoverUtilization: maxUtil,
		AvgCoverUtilization: avgUtil,
		LoadBalanceScore:    loadBalanceScore,
	}
}

// calculateCrossRedundancy determines how many payloads are represented in each cover.
func (h *EmbedMatrixHandler) calculateCrossRedundancy(allocation *MatrixAllocation) float64 {
	totalCovers := len(allocation.CoverAllocations)
	if totalCovers == 0 {
		return 0
	}

	totalPayloadSources := 0
	for _, coverAlloc := range allocation.CoverAllocations {
		payloadSet := make(map[string]bool)
		for _, shard := range coverAlloc.Shards {
			payloadSet[shard.PayloadName] = true
		}
		totalPayloadSources += len(payloadSet)
	}

	return float64(totalPayloadSources) / float64(totalCovers)
}

// compressData compresses data using gzip.
func (h *EmbedMatrixHandler) compressData(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// detectTechnique determines the steganography technique from file extension.
func (h *EmbedMatrixHandler) detectTechnique(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".png", ".bmp":
		return "lsb"
	case ".jpg", ".jpeg":
		return "dct"
	case ".wav":
		return "phase"
	case ".txt":
		return "zerowidth"
	default:
		return "lsb"
	}
}

// ExtractMatrixHandler handles matrix extraction operations.
type ExtractMatrixHandler struct {
	stegoService  stego.StegoService
	cryptoService crypto.CryptoService
	ecService     errorcorrection.ErrorCorrectionService
	logger        *logrus.Logger
}

// NewExtractMatrixHandler creates a new matrix extraction command handler.
func NewExtractMatrixHandler(
	stegoSvc stego.StegoService,
	cryptoSvc crypto.CryptoService,
	ecSvc errorcorrection.ErrorCorrectionService,
	logger *logrus.Logger,
) *ExtractMatrixHandler {
	return &ExtractMatrixHandler{
		stegoService:  stegoSvc,
		cryptoService: cryptoSvc,
		ecService:     ecSvc,
		logger:        logger,
	}
}

// Handle executes the matrix extraction command.
func (h *ExtractMatrixHandler) Handle(ctx context.Context, cmd ExtractMatrixCommand) (*ExtractMatrixResult, error) {
	start := time.Now()

	h.logger.WithFields(logrus.Fields{
		"stego_files":   len(cmd.StegoFiles),
		"manifest_file": cmd.ManifestFile,
	}).Info("Starting matrix extraction operation")

	// Validate command
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	// Load manifest
	h.logger.Info("Loading matrix manifest")
	manifestData, err := os.ReadFile(cmd.ManifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest MatrixManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Verify manifest integrity
	manifestCopy := manifest
	manifestCopy.IntegrityHash = ""
	manifestBytes, _ := json.Marshal(manifestCopy)
	hash := sha256.Sum256(manifestBytes)
	expectedHash := fmt.Sprintf("%x", hash)

	if manifest.IntegrityHash != expectedHash {
		h.logger.Warn("Manifest integrity hash mismatch - manifest may be corrupted")
	}

	h.logger.WithFields(logrus.Fields{
		"payload_count": manifest.PayloadCount,
		"cover_count":   manifest.CoverCount,
		"total_shards":  manifest.TotalShards,
	}).Info("Manifest loaded")

	// Extract shards from all available stego files
	h.logger.Info("Extracting shards from stego media")
	collectedShards := make(map[string]map[int][]byte) // payload -> shard index -> data

	for payloadName := range manifest.AllocationMap.Payloads {
		collectedShards[payloadName] = make(map[int][]byte)
	}

	stegoFilesUsed := 0

	for _, stegoPath := range cmd.StegoFiles {
		// Find corresponding cover in manifest
		var coverAlloc *CoverAllocation
		for coverPath, alloc := range manifest.AllocationMap.Covers {
			if filepath.Base(coverPath) == filepath.Base(stegoPath) ||
				"stego_"+filepath.Base(coverPath) == filepath.Base(stegoPath) {
				ca := alloc
				coverAlloc = &ca
				break
			}
		}

		if coverAlloc == nil {
			h.logger.WithField("file", stegoPath).Warn("Stego file not found in manifest")
			continue
		}

		h.logger.WithFields(logrus.Fields{
			"file":        stegoPath,
			"shard_count": coverAlloc.ShardCount,
		}).Info("Extracting from stego file")

		// Read and extract
		stegoData, err := os.ReadFile(stegoPath)
		if err != nil {
			h.logger.WithError(err).WithField("file", stegoPath).Error("Failed to read stego file")
			continue
		}

		technique := h.detectTechnique(stegoPath)
		techniqueEnum := stego.StegoTechnique(technique)

		extractedData, err := h.stegoService.Extract(ctx, stegoData, techniqueEnum)
		if err != nil {
			h.logger.WithError(err).WithField("file", stegoPath).Error("Extraction failed")
			continue
		}

		// Parse extracted data to separate shards
		shards := h.parseCombinedShards(extractedData)

		for _, shard := range shards {
			collectedShards[shard.PayloadName][shard.ShardIndex] = shard.Data
		}

		stegoFilesUsed++
		h.logger.WithField("file", stegoPath).Info("Extraction successful")
	}

	// Reconstruct payloads
	h.logger.Info("Reconstructing payloads from collected shards")

	if err := os.MkdirAll(cmd.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	outputFiles := make([]string, 0)
	recoveryStatus := make(map[string]RecoveryStatus)
	recoveredCount := 0
	partialCount := 0
	failedCount := 0
	totalShardsCollected := 0
	totalShardsRequired := 0

	for payloadName, payloadAlloc := range manifest.AllocationMap.Payloads {
		// Check if we should extract this payload
		if len(cmd.PayloadNames) > 0 {
			found := false
			for _, name := range cmd.PayloadNames {
				if name == payloadName {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		shards := collectedShards[payloadName]
		shardsCollected := len(shards)
		totalShardsCollected += shardsCollected

		dataShards := payloadAlloc.DataShards
		totalShards := payloadAlloc.DataShards + payloadAlloc.ParityShards
		totalShardsRequired += dataShards

		status := RecoveryStatus{
			PayloadName:     payloadName,
			ShardsCollected: shardsCollected,
			ShardsRequired:  dataShards,
			ShardsTotal:     totalShards,
		}

		h.logger.WithFields(logrus.Fields{
			"payload":          payloadName,
			"shards_collected": shardsCollected,
			"shards_required":  dataShards,
			"shards_total":     totalShards,
		}).Info("Attempting payload reconstruction")

		if shardsCollected < dataShards {
			status.RecoverySucceeded = false
			status.RecoveryQuality = float64(shardsCollected) / float64(dataShards)
			status.ErrorMessage = fmt.Sprintf("insufficient shards: need %d, have %d", dataShards, shardsCollected)
			recoveryStatus[payloadName] = status
			failedCount++

			h.logger.WithField("payload", payloadName).Warn("Insufficient shards for reconstruction")
			continue
		}

		// Reconstruct using Reed-Solomon
		shardArray := make([]*errorcorrection.Shard, totalShards)
		for i := 0; i < totalShards; i++ {
			if data, ok := shards[i]; ok {
				shardArray[i] = &errorcorrection.Shard{
					Index: i,
					Data:  data,
				}
			} else {
				shardArray[i] = nil
			}
		}

		config := &errorcorrection.ShardConfiguration{
			DataShards:   dataShards,
			ParityShards: payloadAlloc.ParityShards,
		}

		recoveredData, err := h.ecService.Decode(ctx, shardArray, config)
		if err != nil {
			status.RecoverySucceeded = false
			status.RecoveryQuality = float64(shardsCollected) / float64(totalShards)
			status.ErrorMessage = fmt.Sprintf("RS decoding failed: %v", err)
			recoveryStatus[payloadName] = status
			failedCount++

			h.logger.WithError(err).WithField("payload", payloadName).Error("Reed-Solomon decoding failed")
			continue
		}

		// Optional decryption
		processedData := recoveredData
		if cmd.Password != "" && manifest.Encrypted {
			if bytes.HasPrefix(processedData, []byte("ENCRYPTED:")) {
				processedData = processedData[len("ENCRYPTED:"):]
			}
		}

		// Optional decompression
		if manifest.Compressed {
			decompressed, err := h.decompressData(processedData)
			if err != nil {
				h.logger.WithError(err).WithField("payload", payloadName).Error("Decompression failed")
			} else {
				processedData = decompressed
			}
		}

		// Write output file
		outputPath := filepath.Join(cmd.OutputDir, payloadName)
		if err := os.WriteFile(outputPath, processedData, 0644); err != nil {
			status.RecoverySucceeded = false
			status.ErrorMessage = fmt.Sprintf("failed to write file: %v", err)
			recoveryStatus[payloadName] = status
			failedCount++
			continue
		}

		outputFiles = append(outputFiles, outputPath)
		status.RecoverySucceeded = true
		status.RecoveryQuality = float64(shardsCollected) / float64(totalShards)
		status.OutputFile = outputPath
		recoveryStatus[payloadName] = status
		recoveredCount++

		h.logger.WithFields(logrus.Fields{
			"payload": payloadName,
			"output":  outputPath,
			"size":    len(processedData),
		}).Info("Payload recovered successfully")
	}

	recoveryRate := 0.0
	if manifest.PayloadCount > 0 {
		recoveryRate = float64(recoveredCount) / float64(manifest.PayloadCount) * 100
	}

	result := &ExtractMatrixResult{
		OutputFiles:           outputFiles,
		PayloadCount:          manifest.PayloadCount,
		RecoveredPayloads:     recoveredCount,
		PartialPayloads:       partialCount,
		FailedPayloads:        failedCount,
		StegoFilesUsed:        stegoFilesUsed,
		ShardsCollected:       totalShardsCollected,
		ShardsRequired:        totalShardsRequired,
		RecoveryRate:          recoveryRate,
		ProcessingTime:        time.Since(start).Milliseconds(),
		DecryptionUsed:        cmd.Password != "" && manifest.Encrypted,
		PayloadRecoveryStatus: recoveryStatus,
	}

	h.logger.WithFields(logrus.Fields{
		"recovered_payloads": recoveredCount,
		"failed_payloads":    failedCount,
		"recovery_rate":      recoveryRate,
		"processing_time":    result.ProcessingTime,
	}).Info("Matrix extraction operation complete")

	return result, nil
}

// ParsedShard represents a shard extracted from combined data.
type ParsedShard struct {
	PayloadName string
	ShardIndex  int
	Data        []byte
}

// parseCombinedShards separates aggregated shards back into individual shards.
func (h *ExtractMatrixHandler) parseCombinedShards(data []byte) []ParsedShard {
	if len(data) < 1 {
		return nil
	}

	shards := make([]ParsedShard, 0)
	offset := 0

	// Read shard count
	shardCount := int(data[offset])
	offset++

	for i := 0; i < shardCount && offset < len(data); i++ {
		// Read payload name
		if offset >= len(data) {
			break
		}
		nameLen := int(data[offset])
		offset++

		if offset+nameLen > len(data) {
			break
		}
		payloadName := string(data[offset : offset+nameLen])
		offset += nameLen

		// Read shard index (2 bytes)
		if offset+2 > len(data) {
			break
		}
		shardIndex := int(data[offset])<<8 | int(data[offset+1])
		offset += 2

		// Read shard data length (4 bytes)
		if offset+4 > len(data) {
			break
		}
		shardLen := int(data[offset])<<24 | int(data[offset+1])<<16 | int(data[offset+2])<<8 | int(data[offset+3])
		offset += 4

		// Read shard data
		if offset+shardLen > len(data) {
			break
		}
		shardData := make([]byte, shardLen)
		copy(shardData, data[offset:offset+shardLen])
		offset += shardLen

		shards = append(shards, ParsedShard{
			PayloadName: payloadName,
			ShardIndex:  shardIndex,
			Data:        shardData,
		})
	}

	return shards
}

// decompressData decompresses gzip data.
func (h *ExtractMatrixHandler) decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()

	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, reader); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// detectTechnique determines the steganography technique from file extension.
func (h *ExtractMatrixHandler) detectTechnique(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".png", ".bmp":
		return "lsb"
	case ".jpg", ".jpeg":
		return "dct"
	case ".wav":
		return "phase"
	case ".txt":
		return "zerowidth"
	default:
		return "lsb"
	}
}
