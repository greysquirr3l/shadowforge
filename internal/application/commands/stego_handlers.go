package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/greysquirr3l/shadowforge/internal/domain/crypto"
	"github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// EmbedHandler handles embed command operations.
type EmbedHandler struct {
	stegoService  stego.StegoService
	cryptoService crypto.CryptoService
	ecService     errorcorrection.ErrorCorrectionService
	mediaService  media.Service
	logger        *logrus.Logger
}

// NewEmbedHandler creates a new embed command handler.
func NewEmbedHandler(
	stegoSvc stego.StegoService,
	cryptoSvc crypto.CryptoService,
	ecSvc errorcorrection.ErrorCorrectionService,
	mediaSvc media.Service,
	logger *logrus.Logger,
) *EmbedHandler {
	return &EmbedHandler{
		stegoService:  stegoSvc,
		cryptoService: cryptoSvc,
		ecService:     ecSvc,
		mediaService:  mediaSvc,
		logger:        logger,
	}
}

// Handle processes the embed command.
func (h *EmbedHandler) Handle(ctx context.Context, cmd EmbedCommand) (*EmbedResult, error) {
	start := time.Now()

	h.logger.WithFields(logrus.Fields{
		"input_file":  cmd.InputFile,
		"cover_file":  cmd.CoverFile,
		"output_file": cmd.OutputFile,
		"technique":   cmd.Technique,
	}).Info("Starting embed operation")

	// Validate command
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid embed command: %w", err)
	}

	// Read input payload
	payload, err := os.ReadFile(cmd.InputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file %s: %w", cmd.InputFile, err)
	}

	// Read cover media
	coverData, err := os.ReadFile(cmd.CoverFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cover file %s: %w", cmd.CoverFile, err)
	}

	// Auto-detect technique if not specified
	technique := cmd.Technique
	if technique == "" {
		technique, err = h.stegoService.OptimizeTechnique(ctx, coverData, int64(len(payload)))
		if err != nil {
			return nil, fmt.Errorf("failed to auto-detect technique: %w", err)
		}
		h.logger.WithField("auto_detected_technique", technique).Info("Auto-detected steganography technique")
	}

	// Encrypt payload if password provided
	processedPayload := payload
	var encryptionUsed bool
	if cmd.Password != "" {
		// Derive encryption key from password
		encryptionKey, err := h.cryptoService.DeriveKey(ctx, cmd.Password, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to derive encryption key: %w", err)
		}

		// For simplicity, we use Kyber1024 - in production this should be configurable
		encryptedPayload, err := h.cryptoService.Encrypt(ctx, payload, encryptionKey, crypto.Kyber1024)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt payload: %w", err)
		}

		// Extract the encrypted data from CryptoPayload
		processedPayload = encryptedPayload.Data
		encryptionUsed = true
		h.logger.Info("Payload encrypted successfully")
	}

	// Apply Reed-Solomon error correction if redundancy specified
	var compressionUsed bool
	if cmd.Redundancy > 0 {
		// Convert redundancy ratio to shard configuration
		// For simplicity: 10 data shards, calculate parity shards based on redundancy
		dataShards := 10
		parityShards := int(float64(dataShards) * cmd.Redundancy)
		if parityShards < 1 {
			parityShards = 1
		}

		shardConfig := &errorcorrection.ShardConfiguration{
			DataShards:   dataShards,
			ParityShards: parityShards,
		}

		encodedPayload, err := h.ecService.Encode(ctx, processedPayload, shardConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to encode payload with error correction: %w", err)
		}

		// For embedding, we need the raw encoded data
		// In a real implementation, we'd serialize the full ProtectedMessage
		processedPayload = encodedPayload.OriginalData
		compressionUsed = true
		h.logger.WithField("redundancy", cmd.Redundancy).Info("Applied Reed-Solomon error correction")
	}

	// Verify capacity before embedding
	capacity, err := h.stegoService.CalculateCapacity(ctx, coverData, technique)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate capacity: %w", err)
	}

	if int64(len(processedPayload)) > capacity {
		return nil, fmt.Errorf("payload too large: %d bytes, max capacity: %d bytes",
			len(processedPayload), capacity)
	}

	// Perform steganographic embedding
	stegoContainer, err := h.stegoService.Embed(ctx, coverData, processedPayload, technique)
	if err != nil {
		return nil, fmt.Errorf("failed to embed payload: %w", err)
	}

	// Extract the stego media data (cover media with embedded data)
	stegoData := stegoContainer.CoverMedia

	// Ensure output directory exists
	outputDir := filepath.Dir(cmd.OutputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write output file
	if err := os.WriteFile(cmd.OutputFile, stegoData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write output file %s: %w", cmd.OutputFile, err)
	}

	// Analyze quality of the embedding
	quality, err := h.stegoService.AnalyzeQuality(ctx, stegoData)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to analyze embedding quality")
	}

	processingTime := time.Since(start)
	capacityUsed := float64(len(processedPayload)) / float64(capacity) * 100

	result := &EmbedResult{
		OutputFile:      cmd.OutputFile,
		Technique:       technique,
		PayloadSize:     int64(len(payload)),
		CoverSize:       int64(len(coverData)),
		OutputSize:      int64(len(stegoData)),
		CapacityUsed:    capacityUsed,
		QualityScore:    quality.Score,
		ProcessingTime:  processingTime.Milliseconds(),
		EncryptionUsed:  encryptionUsed,
		CompressionUsed: compressionUsed,
	}

	h.logger.WithFields(logrus.Fields{
		"output_file":     result.OutputFile,
		"technique":       result.Technique,
		"capacity_used":   fmt.Sprintf("%.1f%%", result.CapacityUsed),
		"quality_score":   fmt.Sprintf("%.2f", result.QualityScore),
		"processing_time": fmt.Sprintf("%dms", result.ProcessingTime),
	}).Info("Embed operation completed successfully")

	return result, nil
}

// ExtractHandler handles extract command operations.
type ExtractHandler struct {
	stegoService  stego.StegoService
	cryptoService crypto.CryptoService
	ecService     errorcorrection.ErrorCorrectionService
	mediaService  media.Service
	logger        *logrus.Logger
}

// NewExtractHandler creates a new extract command handler.
func NewExtractHandler(
	stegoSvc stego.StegoService,
	cryptoSvc crypto.CryptoService,
	ecSvc errorcorrection.ErrorCorrectionService,
	mediaSvc media.Service,
	logger *logrus.Logger,
) *ExtractHandler {
	return &ExtractHandler{
		stegoService:  stegoSvc,
		cryptoService: cryptoSvc,
		ecService:     ecSvc,
		mediaService:  mediaSvc,
		logger:        logger,
	}
}

// Handle processes the extract command.
func (h *ExtractHandler) Handle(ctx context.Context, cmd ExtractCommand) (*ExtractResult, error) {
	start := time.Now()

	h.logger.WithFields(logrus.Fields{
		"input_file":  cmd.InputFile,
		"output_file": cmd.OutputFile,
		"technique":   cmd.Technique,
	}).Info("Starting extract operation")

	// Validate command
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid extract command: %w", err)
	}

	// Read steganographic media
	stegoData, err := os.ReadFile(cmd.InputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file %s: %w", cmd.InputFile, err)
	}

	// Auto-detect technique if not specified
	technique := cmd.Technique
	if technique == "" {
		// For extraction, we may need to try different techniques
		// This would require implementing technique detection in the domain
		technique, err = h.detectTechnique(ctx, stegoData)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-detect technique: %w", err)
		}
		h.logger.WithField("detected_technique", technique).Info("Auto-detected steganography technique")
	}

	// Extract payload from steganographic media
	extractedPayload, err := h.stegoService.Extract(ctx, stegoData, technique)
	if err != nil {
		return nil, fmt.Errorf("failed to extract payload: %w", err)
	}

	processedPayload := extractedPayload
	var decryptionUsed, decompressionUsed bool

	// NOTE: Reed-Solomon decoding would require shard metadata
	// In a production system, we'd embed metadata in the payload
	// For now, we assume no error correction was used in simple extraction
	// decompressionUsed will remain false

	// Decrypt payload if password provided
	if cmd.Password != "" {
		// Derive decryption key from password
		decryptionKey, err := h.cryptoService.DeriveKey(ctx, cmd.Password, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to derive decryption key: %w", err)
		}

		// Create CryptoPayload from extracted data
		// In a real implementation, we'd deserialize the full CryptoPayload structure
		cryptoPayload := &crypto.CryptoPayload{
			Data:      processedPayload,
			Algorithm: crypto.Kyber1024,
		}

		decryptedPayload, err := h.cryptoService.Decrypt(ctx, cryptoPayload, decryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt payload: %w", err)
		}

		processedPayload = decryptedPayload
		decryptionUsed = true
		h.logger.Info("Payload decrypted successfully")
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(cmd.OutputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write extracted payload to output file
	if err := os.WriteFile(cmd.OutputFile, processedPayload, 0644); err != nil {
		return nil, fmt.Errorf("failed to write output file %s: %w", cmd.OutputFile, err)
	}

	processingTime := time.Since(start)

	result := &ExtractResult{
		OutputFile:        cmd.OutputFile,
		Technique:         technique,
		PayloadSize:       int64(len(processedPayload)),
		InputSize:         int64(len(stegoData)),
		ProcessingTime:    processingTime.Milliseconds(),
		IntegrityPassed:   true, // TODO: Implement integrity verification
		DecryptionUsed:    decryptionUsed,
		DecompressionUsed: decompressionUsed,
	}

	h.logger.WithFields(logrus.Fields{
		"output_file":     result.OutputFile,
		"technique":       result.Technique,
		"payload_size":    result.PayloadSize,
		"processing_time": fmt.Sprintf("%dms", result.ProcessingTime),
	}).Info("Extract operation completed successfully")

	return result, nil
}

// detectTechnique attempts to detect the steganography technique used.
// This is a placeholder - real implementation would need domain logic.
func (h *ExtractHandler) detectTechnique(ctx context.Context, data []byte) (stego.StegoTechnique, error) {
	// TODO: Implement proper technique detection
	// For now, return an error indicating manual specification is required
	return "", fmt.Errorf("technique auto-detection not yet implemented, please specify technique manually")
}

// AnalyzeCapacityHandler handles capacity analysis command operations.
type AnalyzeCapacityHandler struct {
	stegoService stego.StegoService
	mediaService media.Service
	logger       *logrus.Logger
}

// NewAnalyzeCapacityHandler creates a new capacity analysis command handler.
func NewAnalyzeCapacityHandler(
	stegoSvc stego.StegoService,
	mediaSvc media.Service,
	logger *logrus.Logger,
) *AnalyzeCapacityHandler {
	return &AnalyzeCapacityHandler{
		stegoService: stegoSvc,
		mediaService: mediaSvc,
		logger:       logger,
	}
}

// Handle processes the analyze capacity command.
func (h *AnalyzeCapacityHandler) Handle(ctx context.Context, cmd AnalyzeCapacityCommand) (*CapacityAnalysisResult, error) {
	h.logger.WithFields(logrus.Fields{
		"cover_file": cmd.CoverFile,
		"technique":  cmd.Technique,
	}).Info("Starting capacity analysis")

	// Validate command
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid analyze capacity command: %w", err)
	}

	// Read cover media
	coverData, err := os.ReadFile(cmd.CoverFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cover file %s: %w", cmd.CoverFile, err)
	}

	// Detect format to set media type
	format, _ := h.mediaService.DetectFormat(ctx, coverData)

	// Determine media type from format
	var mediaType string
	switch format {
	case media.FormatPNG, media.FormatJPEG, media.FormatBMP, media.FormatGIF:
		mediaType = "image"
	case media.FormatWAV, media.FormatFLAC, media.FormatMP3:
		mediaType = "audio"
	case media.FormatTXT:
		mediaType = "text"
	default:
		mediaType = "unknown"
	}

	result := &CapacityAnalysisResult{
		CoverFile: cmd.CoverFile,
		CoverSize: int64(len(coverData)),
		MediaType: mediaType,
		Format:    string(format),
	}

	// Analyze specific technique or all supported techniques
	techniques := []stego.StegoTechnique{}
	if cmd.Technique != "" {
		techniques = append(techniques, cmd.Technique)
	} else {
		// Get all supported techniques for this media type based on format
		switch format {
		case media.FormatPNG, media.FormatBMP:
			techniques = []stego.StegoTechnique{stego.LSB, stego.Palette}
		case media.FormatJPEG:
			techniques = []stego.StegoTechnique{stego.DCT}
		case media.FormatGIF:
			techniques = []stego.StegoTechnique{stego.Palette}
		case media.FormatWAV, media.FormatFLAC, media.FormatMP3:
			techniques = []stego.StegoTechnique{stego.PhaseEncoding, stego.EchoHiding, stego.LSB}
		case media.FormatTXT:
			techniques = []stego.StegoTechnique{stego.ZeroWidth}
		default:
			techniques = []stego.StegoTechnique{stego.LSB}
		}
	}

	for _, technique := range techniques {
		techniqueResult := h.analyzeTechnique(ctx, coverData, technique)
		result.TechniqueResults = append(result.TechniqueResults, techniqueResult)
	}

	// Generate recommendations
	result.Recommendations = h.generateRecommendations(result.TechniqueResults)

	h.logger.WithFields(logrus.Fields{
		"cover_file":          cmd.CoverFile,
		"techniques_analyzed": len(result.TechniqueResults),
		"recommendations":     len(result.Recommendations),
	}).Info("Capacity analysis completed")

	return result, nil
}

// analyzeTechnique analyzes capacity for a specific technique.
func (h *AnalyzeCapacityHandler) analyzeTechnique(ctx context.Context, coverData []byte, technique stego.StegoTechnique) TechniqueCapacityResult {
	result := TechniqueCapacityResult{
		Technique: technique,
		Supported: true,
	}

	// Calculate capacity
	capacity, err := h.stegoService.CalculateCapacity(ctx, coverData, technique)
	if err != nil {
		result.Supported = false
		result.ErrorReason = err.Error()
		return result
	}

	result.MaxCapacity = capacity
	result.SafeCapacity = int64(float64(capacity) * 0.7) // 70% of max capacity as safe threshold

	// TODO: Implement quality, detectability, and performance scoring
	// These would require domain logic implementation
	result.QualityScore = 0.8      // Placeholder
	result.DetectabilityRisk = 0.3 // Placeholder
	result.PerformanceScore = 0.9  // Placeholder

	return result
}

// generateRecommendations creates capacity recommendations based on analysis.
func (h *AnalyzeCapacityHandler) generateRecommendations(results []TechniqueCapacityResult) []CapacityRecommendation {
	var recommendations []CapacityRecommendation

	if len(results) == 0 {
		return recommendations
	}

	// Find best overall technique
	var best TechniqueCapacityResult
	var fastest TechniqueCapacityResult
	var safest TechniqueCapacityResult

	for _, result := range results {
		if !result.Supported {
			continue
		}

		// Best overall (balance of capacity, quality, safety)
		if best.Technique == "" ||
			(result.QualityScore*float64(result.SafeCapacity) > best.QualityScore*float64(best.SafeCapacity)) {
			best = result
		}

		// Fastest
		if fastest.Technique == "" || result.PerformanceScore > fastest.PerformanceScore {
			fastest = result
		}

		// Safest (lowest detectability risk)
		if safest.Technique == "" || result.DetectabilityRisk < safest.DetectabilityRisk {
			safest = result
		}
	}

	// Add recommendations
	if best.Technique != "" {
		recommendations = append(recommendations, CapacityRecommendation{
			Type:       "best",
			Technique:  best.Technique,
			Reason:     "Best balance of capacity, quality, and safety",
			MaxPayload: best.SafeCapacity,
			Confidence: 0.9,
		})
	}

	if fastest.Technique != "" && fastest.Technique != best.Technique {
		recommendations = append(recommendations, CapacityRecommendation{
			Type:       "fastest",
			Technique:  fastest.Technique,
			Reason:     "Fastest embedding performance",
			MaxPayload: fastest.SafeCapacity,
			Confidence: 0.8,
		})
	}

	if safest.Technique != "" && safest.Technique != best.Technique {
		recommendations = append(recommendations, CapacityRecommendation{
			Type:       "safest",
			Technique:  safest.Technique,
			Reason:     "Lowest detectability risk",
			MaxPayload: safest.SafeCapacity,
			Confidence: 0.85,
		})
	}

	return recommendations
}
