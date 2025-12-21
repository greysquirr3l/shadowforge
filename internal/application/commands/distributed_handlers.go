package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/greysquirr3l/shadowforge/internal/domain/crypto"
	"github.com/greysquirr3l/shadowforge/internal/domain/distribution"
	"github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// EmbedDistributedHandler handles distributed embed command operations (1:N pattern).
type EmbedDistributedHandler struct {
	stegoService        stego.StegoService
	cryptoService       crypto.CryptoService
	ecService           errorcorrection.ErrorCorrectionService
	mediaService        media.Service
	distributionService distribution.Service
	manifestSerializer  *distribution.ManifestSerializer
	logger              *logrus.Logger
}

// NewEmbedDistributedHandler creates a new distributed embed command handler.
func NewEmbedDistributedHandler(
	stegoSvc stego.StegoService,
	cryptoSvc crypto.CryptoService,
	ecSvc errorcorrection.ErrorCorrectionService,
	mediaSvc media.Service,
	distSvc distribution.Service,
	manifestSer *distribution.ManifestSerializer,
	logger *logrus.Logger,
) *EmbedDistributedHandler {
	return &EmbedDistributedHandler{
		stegoService:        stegoSvc,
		cryptoService:       cryptoSvc,
		ecService:           ecSvc,
		mediaService:        mediaSvc,
		distributionService: distSvc,
		manifestSerializer:  manifestSer,
		logger:              logger,
	}
}

// Handle processes the distributed embed command.
func (h *EmbedDistributedHandler) Handle(ctx context.Context, cmd EmbedDistributedCommand) (*EmbedDistributedResult, error) {
	start := time.Now()

	h.logger.WithFields(logrus.Fields{
		"input_file":       cmd.InputFile,
		"cover_files":      len(cmd.CoverFiles),
		"output_directory": cmd.OutputDirectory,
		"manifest_path":    cmd.ManifestPath,
		"technique":        cmd.Technique,
		"pattern":          cmd.Pattern,
	}).Info("Starting distributed embed operation")

	// Validate command
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid distributed embed command: %w", err)
	}

	// Read input payload
	payload, err := os.ReadFile(cmd.InputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file %s: %w", cmd.InputFile, err)
	}

	h.logger.WithField("payload_size", len(payload)).Info("Payload loaded")

	// Encrypt payload if password provided
	processedPayload := payload
	var encryptionUsed bool
	if cmd.Password != "" {
		encryptionKey, err := h.cryptoService.DeriveKey(ctx, cmd.Password, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to derive encryption key: %w", err)
		}

		encryptedPayload, err := h.cryptoService.Encrypt(ctx, payload, encryptionKey, crypto.Kyber1024)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt payload: %w", err)
		}

		processedPayload = encryptedPayload.Data
		encryptionUsed = true
		h.logger.Info("Payload encrypted successfully")
	}

	// Calculate shard configuration
	totalShards := len(cmd.CoverFiles)
	var dataShards, parityShards int

	if cmd.DataShards > 0 && cmd.ParityShards > 0 {
		// User-specified configuration
		dataShards = cmd.DataShards
		parityShards = cmd.ParityShards
	} else if cmd.RedundancyLevel > 0 {
		// Calculate from redundancy level
		dataShards = int(float64(totalShards) / (1.0 + cmd.RedundancyLevel))
		parityShards = totalShards - dataShards
	} else {
		// Use distribution service to calculate optimal configuration
		dataShards, parityShards, err = h.distributionService.CalculateOptimalSharding(ctx, int64(len(processedPayload)), totalShards)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate optimal sharding: %w", err)
		}
	}

	requiredShards := cmd.RequiredShards
	if requiredShards == 0 {
		requiredShards = dataShards
	}

	h.logger.WithFields(logrus.Fields{
		"total_shards":    totalShards,
		"data_shards":     dataShards,
		"parity_shards":   parityShards,
		"required_shards": requiredShards,
	}).Info("Shard configuration calculated")

	// Create distribution strategy using service
	pattern := cmd.Pattern
	if pattern == "" {
		pattern = distribution.PatternOneToMany
	}

	strategy, err := h.distributionService.CreateStrategy(
		ctx,
		pattern,
		dataShards,
		parityShards,
		cmd.CoverFiles,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create distribution strategy: %w", err)
	}

	strategyID := strategy.ID
	h.logger.WithField("strategy_id", strategyID).Info("Distribution strategy created")

	// Apply Reed-Solomon error correction
	shardConfig := &errorcorrection.ShardConfiguration{
		DataShards:   dataShards,
		ParityShards: parityShards,
	}

	protectedMessage, err := h.ecService.Encode(ctx, processedPayload, shardConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to encode payload with error correction: %w", err)
	}

	h.logger.WithField("total_shards", len(protectedMessage.Shards)).Info("Reed-Solomon encoding complete")

	// Read all cover media
	coverMedia := make([][]byte, len(cmd.CoverFiles))
	for i, coverFile := range cmd.CoverFiles {
		coverData, err := os.ReadFile(coverFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read cover file %s: %w", coverFile, err)
		}
		coverMedia[i] = coverData
	}

	// Auto-detect technique if not specified (use first cover file as reference)
	technique := cmd.Technique
	if technique == "" && len(coverMedia) > 0 {
		technique, err = h.stegoService.OptimizeTechnique(ctx, coverMedia[0], int64(len(payload)/totalShards))
		if err != nil {
			return nil, fmt.Errorf("failed to auto-detect technique: %w", err)
		}
		h.logger.WithField("auto_detected_technique", technique).Info("Auto-detected steganography technique")
	}

	// Distribute shards across cover media (simple round-robin allocation)
	// Each shard goes to a corresponding cover file
	if len(protectedMessage.Shards) != len(cmd.CoverFiles) {
		return nil, fmt.Errorf("shard count (%d) must match cover file count (%d)",
			len(protectedMessage.Shards), len(cmd.CoverFiles))
	}

	h.logger.WithField("shard_count", len(protectedMessage.Shards)).Info("Beginning parallel shard embedding")

	// Ensure output directory exists
	if err := os.MkdirAll(cmd.OutputDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Parallel embedding with goroutine pool
	maxWorkers := cmd.MaxWorkers
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	type embeddingTask struct {
		index     int
		shardData []byte
		coverData []byte
		coverFile string
	}

	type embeddingResult struct {
		index      int
		stegoData  []byte
		outputPath string
		capacity   float64
		err        error
	}

	tasks := make(chan embeddingTask, len(protectedMessage.Shards))
	results := make(chan embeddingResult, len(protectedMessage.Shards))
	var wg sync.WaitGroup

	// Start worker goroutines
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range tasks {
				h.logger.WithFields(logrus.Fields{
					"worker": workerID,
					"shard":  task.index,
				}).Debug("Processing shard embedding")

				// Calculate capacity
				capacity, capErr := h.stegoService.CalculateCapacity(ctx, task.coverData, technique)
				if capErr != nil {
					results <- embeddingResult{index: task.index, err: fmt.Errorf("capacity calculation failed: %w", capErr)}
					continue
				}

				if int64(len(task.shardData)) > capacity {
					results <- embeddingResult{index: task.index, err: fmt.Errorf("shard too large: %d bytes, max capacity: %d bytes", len(task.shardData), capacity)}
					continue
				}

				// Embed shard
				stegoContainer, embedErr := h.stegoService.Embed(ctx, task.coverData, task.shardData, technique)
				if embedErr != nil {
					results <- embeddingResult{index: task.index, err: fmt.Errorf("embedding failed: %w", embedErr)}
					continue
				}

				// Generate output filename
				baseFilename := filepath.Base(task.coverFile)
				ext := filepath.Ext(baseFilename)
				nameWithoutExt := baseFilename[:len(baseFilename)-len(ext)]
				outputFilename := fmt.Sprintf("%s_shard_%03d%s", nameWithoutExt, task.index, ext)
				outputPath := filepath.Join(cmd.OutputDirectory, outputFilename)

				results <- embeddingResult{
					index:      task.index,
					stegoData:  stegoContainer.CoverMedia,
					outputPath: outputPath,
					capacity:   float64(len(task.shardData)) / float64(capacity) * 100.0,
					err:        nil,
				}
			}
		}(w)
	}

	// Queue embedding tasks (1:1 mapping: shard i -> cover file i)
	for i := range protectedMessage.Shards {
		tasks <- embeddingTask{
			index:     i,
			shardData: protectedMessage.Shards[i].Data,
			coverData: coverMedia[i],
			coverFile: cmd.CoverFiles[i],
		}
	}
	close(tasks)

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	outputFiles := make([]string, len(protectedMessage.Shards))
	var capacitySum float64
	var failedShards []int

	for result := range results {
		if result.err != nil {
			h.logger.WithFields(logrus.Fields{
				"shard": result.index,
				"error": result.err,
			}).Error("Shard embedding failed")
			failedShards = append(failedShards, result.index)
			continue
		}

		// Write stego file
		if err := os.WriteFile(result.outputPath, result.stegoData, 0644); err != nil {
			h.logger.WithFields(logrus.Fields{
				"shard":       result.index,
				"output_file": result.outputPath,
				"error":       err,
			}).Error("Failed to write stego file")
			failedShards = append(failedShards, result.index)
			continue
		}

		outputFiles[result.index] = result.outputPath
		capacitySum += result.capacity
		h.logger.WithField("shard", result.index).Debug("Shard embedded successfully")
	}

	// Check if we have enough successful shards
	successfulShards := len(protectedMessage.Shards) - len(failedShards)
	if successfulShards < requiredShards {
		return nil, fmt.Errorf("insufficient successful embeddings: %d/%d (required: %d)",
			successfulShards, len(protectedMessage.Shards), requiredShards)
	}

	h.logger.WithFields(logrus.Fields{
		"successful_shards": successfulShards,
		"failed_shards":     len(failedShards),
		"total_shards":      len(protectedMessage.Shards),
	}).Info("Parallel embedding complete")

	// Generate manifest
	manifestID, err := distribution.NewManifestID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate manifest ID: %w", err)
	}

	shardMetadata := make([]distribution.ShardMetadata, 0, len(protectedMessage.Shards))

	for i := range protectedMessage.Shards {
		// Check if this shard failed
		failed := false
		for _, failedIdx := range failedShards {
			if failedIdx == i {
				failed = true
				break
			}
		}
		if failed {
			continue
		}

		// Get file info
		fileInfo, _ := os.Stat(outputFiles[i])
		var fileSize int64
		if fileInfo != nil {
			fileSize = fileInfo.Size()
		}

		metadata := distribution.ShardMetadata{
			Index:    i,
			Size:     fileSize,
			Checksum: fmt.Sprintf("%x", protectedMessage.Shards[i].Data[:16]), // Simple checksum
			MediaID:  cmd.CoverFiles[i],                                       // Use cover file path as media ID
		}
		shardMetadata = append(shardMetadata, metadata)
	}

	manifest := &distribution.ShardManifest{
		ID:             manifestID,
		StrategyID:     strategyID,
		ShardMetadata:  shardMetadata,
		TotalShards:    totalShards,
		RequiredShards: requiredShards,
		CreatedAt:      time.Now(),
	}

	// Serialize and save manifest
	manifestData, err := h.manifestSerializer.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize manifest: %w", err)
	}

	if err := os.WriteFile(cmd.ManifestPath, manifestData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write manifest file: %w", err)
	}

	h.logger.WithField("manifest_path", cmd.ManifestPath).Info("Manifest generated and saved")

	// Build result
	result := &EmbedDistributedResult{
		OutputFiles:     outputFiles,
		ManifestPath:    cmd.ManifestPath,
		Technique:       technique,
		Pattern:         pattern,
		PayloadSize:     int64(len(payload)),
		TotalShards:     totalShards,
		DataShards:      dataShards,
		ParityShards:    parityShards,
		RequiredShards:  requiredShards,
		AverageCapacity: capacitySum / float64(successfulShards),
		ProcessingTime:  time.Since(start).Milliseconds(),
		ParallelWorkers: maxWorkers,
		EncryptionUsed:  encryptionUsed,
		ManifestSigned:  true, // HMAC signature always applied
		FailedShards:    failedShards,
	}

	h.logger.WithFields(logrus.Fields{
		"output_files":    len(result.OutputFiles),
		"manifest_path":   result.ManifestPath,
		"processing_time": result.ProcessingTime,
		"failed_shards":   len(failedShards),
	}).Info("Distributed embed operation complete")

	return result, nil
}

// ExtractDistributedHandler handles distributed extraction command operations (1:N pattern).
type ExtractDistributedHandler struct {
	stegoService        stego.StegoService
	cryptoService       crypto.CryptoService
	ecService           errorcorrection.ErrorCorrectionService
	distributionService distribution.Service
	manifestSerializer  *distribution.ManifestSerializer
	logger              *logrus.Logger
}

// NewExtractDistributedHandler creates a new distributed extraction command handler.
func NewExtractDistributedHandler(
	stegoSvc stego.StegoService,
	cryptoSvc crypto.CryptoService,
	ecSvc errorcorrection.ErrorCorrectionService,
	distSvc distribution.Service,
	manifestSer *distribution.ManifestSerializer,
	logger *logrus.Logger,
) *ExtractDistributedHandler {
	return &ExtractDistributedHandler{
		stegoService:        stegoSvc,
		cryptoService:       cryptoSvc,
		ecService:           ecSvc,
		distributionService: distSvc,
		manifestSerializer:  manifestSer,
		logger:              logger,
	}
}

// Handle executes the distributed extraction command.
// Workflow:
// 1. Read and unmarshal manifest
// 2. Validate HMAC signature
// 3. Check recoverability (K-of-N threshold)
// 4. Read available stego files
// 5. Parallel shard extraction (worker pool)
// 6. Reed-Solomon reconstruction
// 7. Optional decryption
// 8. Write reconstructed payload
func (h *ExtractDistributedHandler) Handle(ctx context.Context, cmd ExtractDistributedCommand) (*ExtractDistributedResult, error) {
	start := time.Now()

	h.logger.WithFields(logrus.Fields{
		"manifest_path":   cmd.ManifestPath,
		"available_files": len(cmd.StegoFiles),
		"decryption":      cmd.Password != "",
	}).Info("Starting distributed extraction")

	// Validate command
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	// Phase 1: Read and deserialize manifest
	h.logger.Info("Reading manifest file")
	manifestData, err := os.ReadFile(cmd.ManifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	manifest, err := h.manifestSerializer.Unmarshal(manifestData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize manifest (HMAC validation failed): %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"total_shards":    manifest.TotalShards,
		"required_shards": manifest.RequiredShards,
		"manifest_id":     manifest.ID.String(),
	}).Info("Manifest loaded and validated")

	// Phase 2: Check recoverability
	availableCount := len(cmd.StegoFiles)
	if availableCount < manifest.RequiredShards {
		return nil, fmt.Errorf("insufficient shards: have %d, need %d (K-of-N threshold not met)",
			availableCount, manifest.RequiredShards)
	}

	h.logger.WithFields(logrus.Fields{
		"available": availableCount,
		"required":  manifest.RequiredShards,
		"total":     manifest.TotalShards,
	}).Info("Recoverability check passed")

	// Phase 3: Setup worker pool for parallel extraction
	maxWorkers := runtime.NumCPU() // Use all CPUs for extraction

	h.logger.WithField("workers", maxWorkers).Info("Initializing extraction worker pool")

	type extractionTask struct {
		index     int
		stegoFile string
		metadata  distribution.ShardMetadata
	}

	type extractionResult struct {
		index     int
		shardData []byte
		checksum  string
		err       error
	}

	tasks := make(chan extractionTask, len(cmd.StegoFiles))
	results := make(chan extractionResult, len(cmd.StegoFiles))

	// Phase 4: Start extraction workers
	var wg sync.WaitGroup
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for task := range tasks {
				h.logger.WithFields(logrus.Fields{
					"worker":     workerID,
					"shard":      task.index,
					"stego_file": task.stegoFile,
				}).Debug("Processing extraction task")

				// Read stego file
				stegoData, err := os.ReadFile(task.stegoFile)
				if err != nil {
					results <- extractionResult{
						index: task.index,
						err:   fmt.Errorf("failed to read stego file: %w", err),
					}
					continue
				}

				// Detect technique from file (could enhance this with manifest metadata)
				techniqueStr := h.detectTechniqueFromFile(task.stegoFile)
				technique := stego.StegoTechnique(techniqueStr)

				// Extract shard using steganography service
				extractedData, err := h.stegoService.Extract(ctx, stegoData, technique)
				if err != nil {
					results <- extractionResult{
						index: task.index,
						err:   fmt.Errorf("extraction failed: %w", err),
					}
					continue
				}

				// Calculate checksum for verification
				checksum := fmt.Sprintf("%x", extractedData[:min(16, len(extractedData))])

				results <- extractionResult{
					index:     task.index,
					shardData: extractedData,
					checksum:  checksum,
					err:       nil,
				}

				h.logger.WithFields(logrus.Fields{
					"worker": workerID,
					"shard":  task.index,
					"size":   len(extractedData),
				}).Debug("Extraction task complete")
			}
		}(w)
	}

	// Phase 5: Queue extraction tasks
	h.logger.Info("Queuing extraction tasks")
	for i, stegoFile := range cmd.StegoFiles {
		if i >= len(manifest.ShardMetadata) {
			break // Don't process more files than shards
		}

		tasks <- extractionTask{
			index:     i,
			stegoFile: stegoFile,
			metadata:  manifest.ShardMetadata[i],
		}
	}
	close(tasks)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Phase 6: Collect extraction results
	h.logger.Info("Collecting extraction results")
	shardData := make([][]byte, manifest.TotalShards)
	successfulExtractions := 0
	failedExtractions := 0
	var extractionErrors []string

	for result := range results {
		if result.err != nil {
			h.logger.WithFields(logrus.Fields{
				"shard": result.index,
				"error": result.err,
			}).Warn("Shard extraction failed")
			failedExtractions++
			extractionErrors = append(extractionErrors, fmt.Sprintf("shard %d: %v", result.index, result.err))
			continue
		}

		// Verify checksum if available
		expectedChecksum := manifest.ShardMetadata[result.index].Checksum
		if expectedChecksum != "" && result.checksum != expectedChecksum {
			h.logger.WithFields(logrus.Fields{
				"shard":    result.index,
				"expected": expectedChecksum,
				"actual":   result.checksum,
			}).Warn("Checksum mismatch - shard may be corrupted")
			// Continue anyway - Reed-Solomon can handle some corruption
		}

		shardData[result.index] = result.shardData
		successfulExtractions++

		h.logger.WithFields(logrus.Fields{
			"shard": result.index,
			"size":  len(result.shardData),
		}).Debug("Shard extracted successfully")
	}

	h.logger.WithFields(logrus.Fields{
		"successful": successfulExtractions,
		"failed":     failedExtractions,
		"required":   manifest.RequiredShards,
	}).Info("Extraction phase complete")

	// Verify we have enough shards for reconstruction
	if successfulExtractions < manifest.RequiredShards {
		return nil, fmt.Errorf("insufficient successful extractions: got %d, need %d (some shards corrupted/failed)",
			successfulExtractions, manifest.RequiredShards)
	}

	// Phase 7: Reed-Solomon reconstruction
	h.logger.Info("Starting Reed-Solomon reconstruction")

	config := &errorcorrection.ShardConfiguration{
		DataShards:   manifest.TotalShards - (manifest.TotalShards - manifest.RequiredShards), // Calculate from manifest
		ParityShards: manifest.TotalShards - manifest.RequiredShards,
	}

	// Convert [][]byte to []*Shard
	shards := make([]*errorcorrection.Shard, len(shardData))
	for i, data := range shardData {
		if data != nil {
			shards[i] = &errorcorrection.Shard{
				Index: i,
				Data:  data,
			}
		}
	}

	reconstructedData, err := h.ecService.Decode(ctx, shards, config)
	if err != nil {
		return nil, fmt.Errorf("Reed-Solomon reconstruction failed: %w", err)
	}

	h.logger.WithField("size", len(reconstructedData)).Info("Reconstruction successful")

	// Phase 8: Optional decryption
	processedData := reconstructedData
	decryptionUsed := false

	if cmd.Password != "" {
		h.logger.Info("Decrypting payload")

		// Derive key from password
		key, err := h.cryptoService.DeriveKey(ctx, cmd.Password, nil)
		if err != nil {
			return nil, fmt.Errorf("key derivation failed: %w", err)
		}

		// Create CryptoPayload wrapper for decryption
		payload := &crypto.CryptoPayload{
			Data: reconstructedData,
			// Algorithm and other fields would come from manifest in production
		}

		// Decrypt
		decryptedData, err := h.cryptoService.Decrypt(ctx, payload, key)
		if err != nil {
			return nil, fmt.Errorf("decryption failed: %w", err)
		}

		processedData = decryptedData
		decryptionUsed = true

		h.logger.WithField("size", len(decryptedData)).Info("Decryption successful")
	}

	// Phase 9: Write reconstructed payload to output file
	h.logger.WithField("output_file", cmd.OutputFile).Info("Writing reconstructed payload")

	if err := os.WriteFile(cmd.OutputFile, processedData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write output file: %w", err)
	}

	// Build result
	result := &ExtractDistributedResult{
		OutputFile:      cmd.OutputFile,
		PayloadSize:     int64(len(processedData)),
		ShardsUsed:      successfulExtractions,
		ShardsRequired:  manifest.RequiredShards,
		ShardsTotal:     manifest.TotalShards,
		ProcessingTime:  time.Since(start).Milliseconds(),
		IntegrityPassed: true, // HMAC validation passed
		DecryptionUsed:  decryptionUsed,
		RecoveryMode:    h.determineRecoveryMode(successfulExtractions, manifest.RequiredShards, manifest.TotalShards),
		CorruptedShards: h.extractCorruptedIndices(extractionErrors),
	}

	h.logger.WithFields(logrus.Fields{
		"output_file":        result.OutputFile,
		"payload_size":       result.PayloadSize,
		"processing_time":    result.ProcessingTime,
		"failed_extractions": failedExtractions,
	}).Info("Distributed extraction operation complete")

	return result, nil
}

// determineRecoveryMode determines the recovery mode based on shard availability.
func (h *ExtractDistributedHandler) determineRecoveryMode(successful, required, total int) string {
	if successful >= total {
		return "full"
	} else if successful > required {
		return "partial"
	}
	return "minimal"
}

// extractCorruptedIndices extracts shard indices from error messages.
func (h *ExtractDistributedHandler) extractCorruptedIndices(errors []string) []int {
	// Simple implementation - in production would parse error messages
	return []int{}
}

// detectTechniqueFromFile determines the steganography technique based on file extension.
func (h *ExtractDistributedHandler) detectTechniqueFromFile(filename string) string {
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
		return "lsb" // Default fallback
	}
}
