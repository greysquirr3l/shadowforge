// Package stego provides steganography service implementation
package stego

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
	infraMedia "github.com/greysquirr3l/shadowforge/internal/infrastructure/media"
	infraStego "github.com/greysquirr3l/shadowforge/internal/infrastructure/stego"
)

// StegoService implements the domain stego.StegoService interface.
// It coordinates between media processing and steganographic technique implementations.
type StegoService struct {
	mediaService *infraMedia.MediaService
	logger       *logrus.Logger
}

// NewStegoService creates a new steganography service.
func NewStegoService(mediaService *infraMedia.MediaService, logger *logrus.Logger) *StegoService {
	return &StegoService{
		mediaService: mediaService,
		logger:       logger,
	}
}

// Embed embeds data into cover media using the specified technique.
func (s *StegoService) Embed(ctx context.Context, coverMedia, payload []byte, technique stego.StegoTechnique) (*stego.StegoContainer, error) {
	s.logger.WithFields(logrus.Fields{
		"technique":    string(technique),
		"payload_size": len(payload),
		"cover_size":   len(coverMedia),
	}).Info("Embedding payload")

	// Detect cover media format
	format, err := s.mediaService.DetectFormat(ctx, coverMedia)
	if err != nil {
		return nil, fmt.Errorf("failed to detect cover format: %w", err)
	}

	// Embed using appropriate technique
	var stegoData []byte
	switch technique {
	case stego.LSB:
		stegoData, err = s.embedLSB(coverMedia, payload, format)
	case stego.DCT:
		stegoData, err = s.embedDCT(coverMedia, payload)
	case stego.PhaseEncoding:
		stegoData, err = s.embedPhase(coverMedia, payload)
	case stego.EchoHiding:
		stegoData, err = s.embedEcho(coverMedia, payload)
	case stego.ZeroWidth:
		stegoData, err = s.embedZeroWidth(coverMedia, payload)
	case stego.Palette:
		stegoData, err = s.embedPalette(coverMedia, payload)
	default:
		return nil, fmt.Errorf("unsupported technique: %s", technique)
	}

	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	// Create stego container
	containerID := stego.GenerateContainerID()
	container, err := stego.NewStegoContainer(containerID, coverMedia, technique)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Calculate and set capacity
	capacity, err := s.CalculateCapacity(ctx, coverMedia, technique)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate capacity: %w", err)
	}
	if err := container.SetCapacity(capacity); err != nil {
		return nil, fmt.Errorf("failed to set capacity: %w", err)
	}

	// Mark as embedded with the ORIGINAL payload data
	// (stegoData is the complete output image, not what we store in EmbeddedData)
	if err := container.Embed(payload, nil); err != nil {
		return nil, fmt.Errorf("failed to mark as embedded: %w", err)
	}

	// Store the actual stego output in the container
	// (The domain model tracks the payload, but we return the full stego media)
	container.CoverMedia = stegoData

	s.logger.WithFields(logrus.Fields{
		"stego_size": len(stegoData),
	}).Info("Embedding completed")

	return container, nil
}

// Extract extracts hidden data from steganographic media.
func (s *StegoService) Extract(ctx context.Context, stegoMedia []byte, technique stego.StegoTechnique) ([]byte, error) {
	s.logger.WithFields(logrus.Fields{
		"technique":  string(technique),
		"stego_size": len(stegoMedia),
	}).Info("Extracting payload")

	var payload []byte
	var err error

	switch technique {
	case stego.LSB:
		payload, err = s.extractLSB(stegoMedia)
	case stego.DCT:
		payload, err = s.extractDCT(stegoMedia)
	case stego.PhaseEncoding:
		payload, err = s.extractPhase(stegoMedia)
	case stego.EchoHiding:
		payload, err = s.extractEcho(stegoMedia)
	case stego.ZeroWidth:
		payload, err = s.extractZeroWidth(stegoMedia)
	case stego.Palette:
		payload, err = s.extractPalette(stegoMedia)
	default:
		return nil, fmt.Errorf("unsupported technique: %s", technique)
	}

	if err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"payload_size": len(payload),
	}).Info("Extraction completed")

	return payload, nil
}

// CalculateCapacity calculates the maximum embedding capacity for given media.
func (s *StegoService) CalculateCapacity(ctx context.Context, coverMedia []byte, technique stego.StegoTechnique) (int64, error) {
	// Detect media format
	format, err := s.mediaService.DetectFormat(ctx, coverMedia)
	if err != nil {
		return 0, fmt.Errorf("failed to detect format: %w", err)
	}

	// Calculate capacity based on technique and format
	var capacity int64
	switch format {
	case media.FormatPNG, media.FormatBMP:
		capacity, err = s.calculateImageCapacity(coverMedia, technique)
	case media.FormatJPEG:
		if technique == stego.DCT {
			capacity, err = s.calculateDCTCapacity(coverMedia)
		} else {
			return 0, fmt.Errorf("JPEG only supports DCT technique")
		}
	case media.FormatWAV:
		capacity, err = s.calculateAudioCapacity(coverMedia, technique)
	case media.FormatTXT, media.FormatMarkdown:
		capacity, err = s.calculateTextCapacity(coverMedia, technique)
	default:
		return 0, fmt.Errorf("unsupported format: %s", format)
	}

	return capacity, err
}

// AnalyzeQuality analyzes the quality and detectability of an embedding.
func (s *StegoService) AnalyzeQuality(ctx context.Context, stegoMedia []byte) (*stego.Quality, error) {
	// For now, return a basic quality assessment
	// TODO: Implement proper statistical analysis
	quality, err := stego.NewQuality(0.85)
	if err != nil {
		return nil, fmt.Errorf("failed to create quality: %w", err)
	}

	// Set additional quality metrics
	_ = quality.SetDetectability(0.15) // 0.15 = low detectability
	_ = quality.SetFidelityScore(0.92) // 0.92 = high fidelity

	return &quality, nil
}

// ValidateContainer validates a steganographic container for correctness.
func (s *StegoService) ValidateContainer(ctx context.Context, container *stego.StegoContainer) error {
	if container == nil {
		return fmt.Errorf("nil container")
	}

	if len(container.CoverMedia) == 0 {
		return fmt.Errorf("empty cover media")
	}

	if container.Technique == "" {
		return fmt.Errorf("technique not specified")
	}

	return nil
}

// OptimizeTechnique selects the best technique for given media and requirements.
func (s *StegoService) OptimizeTechnique(ctx context.Context, coverMedia []byte, payloadSize int64) (stego.StegoTechnique, error) {
	format, err := s.mediaService.DetectFormat(ctx, coverMedia)
	if err != nil {
		return "", fmt.Errorf("failed to detect format: %w", err)
	}

	// Select technique based on format
	switch format {
	case media.FormatPNG, media.FormatBMP:
		return stego.LSB, nil
	case media.FormatJPEG:
		return stego.DCT, nil
	case media.FormatWAV:
		return stego.PhaseEncoding, nil
	case media.FormatTXT, media.FormatMarkdown:
		return stego.ZeroWidth, nil
	default:
		return stego.LSB, nil // Default fallback
	}
}

// Helper methods for each technique

func (s *StegoService) embedLSB(cover, payload []byte, format media.MediaFormat) ([]byte, error) {
	technique := infraStego.NewLSBTechnique()
	return technique.Embed(context.Background(), cover, payload)
}

func (s *StegoService) extractLSB(stego []byte) ([]byte, error) {
	technique := infraStego.NewLSBTechnique()
	return technique.Extract(context.Background(), stego)
}

func (s *StegoService) embedDCT(cover, payload []byte) ([]byte, error) {
	technique := infraStego.NewDCTTechnique()
	return technique.Embed(context.Background(), cover, payload)
}

func (s *StegoService) extractDCT(stego []byte) ([]byte, error) {
	technique := infraStego.NewDCTTechnique()
	return technique.Extract(context.Background(), stego)
}

func (s *StegoService) embedPhase(cover, payload []byte) ([]byte, error) {
	technique := infraStego.NewPhaseWithDefaults(s.logger)
	return technique.Embed(context.Background(), cover, payload)
}

func (s *StegoService) extractPhase(stego []byte) ([]byte, error) {
	technique := infraStego.NewPhaseWithDefaults(s.logger)
	return technique.Extract(context.Background(), stego)
}

func (s *StegoService) embedEcho(cover, payload []byte) ([]byte, error) {
	technique := infraStego.NewEchoWithDefaults(s.logger)
	return technique.Embed(context.Background(), cover, payload)
}

func (s *StegoService) extractEcho(stego []byte) ([]byte, error) {
	technique := infraStego.NewEchoWithDefaults(s.logger)
	return technique.Extract(context.Background(), stego)
}

func (s *StegoService) embedZeroWidth(cover, payload []byte) ([]byte, error) {
	technique := infraStego.NewTextWithDefaults(s.logger)
	return technique.Embed(context.Background(), cover, payload)
}

func (s *StegoService) extractZeroWidth(stego []byte) ([]byte, error) {
	technique := infraStego.NewTextWithDefaults(s.logger)
	return technique.Extract(context.Background(), stego)
}

func (s *StegoService) embedPalette(cover, payload []byte) ([]byte, error) {
	technique := infraStego.NewPaletteWithDefaults(s.logger)
	return technique.Embed(context.Background(), cover, payload)
}

func (s *StegoService) extractPalette(stego []byte) ([]byte, error) {
	technique := infraStego.NewPaletteWithDefaults(s.logger)
	return technique.Extract(context.Background(), stego)
}

func (s *StegoService) calculateImageCapacity(cover []byte, technique stego.StegoTechnique) (int64, error) {
	processor := infraMedia.NewImageProcessor()
	capacityInfo, err := processor.CalculateCapacity(cover, string(technique))
	if err != nil {
		return 0, err
	}
	return capacityInfo.TotalCapacity, nil
}

func (s *StegoService) calculateDCTCapacity(cover []byte) (int64, error) {
	processor := infraMedia.NewImageProcessor()
	capacityInfo, err := processor.CalculateCapacity(cover, "dct")
	if err != nil {
		return 0, err
	}
	return capacityInfo.TotalCapacity, nil
}

func (s *StegoService) calculateAudioCapacity(cover []byte, technique stego.StegoTechnique) (int64, error) {
	processor := infraMedia.NewAudioProcessor()
	// Convert domain technique to infrastructure technique name
	infraTechnique := s.techniqueToInfraTechnique(technique)
	capacityInfo, err := processor.CalculateCapacity(cover, infraTechnique)
	if err != nil {
		return 0, err
	}
	return capacityInfo.TotalCapacity, nil
}

func (s *StegoService) calculateTextCapacity(cover []byte, technique stego.StegoTechnique) (int64, error) {
	processor := infraMedia.NewTextProcessor()
	// Convert domain technique to infrastructure technique name
	infraTechnique := s.techniqueToInfraTechnique(technique)
	capacityInfo, err := processor.CalculateCapacity(cover, infraTechnique)
	if err != nil {
		return 0, err
	}
	return capacityInfo.TotalCapacity, nil
}

// techniqueToInfraTechnique converts domain technique to infrastructure technique string.
func (s *StegoService) techniqueToInfraTechnique(technique stego.StegoTechnique) string {
	switch technique {
	case stego.LSB:
		return "lsb"
	case stego.DCT:
		return "dct"
	case stego.PhaseEncoding:
		return "phase"
	case stego.EchoHiding:
		return "echo"
	case stego.ZeroWidth:
		return "zero-width"  // Text processor expects hyphenated version
	case stego.Palette:
		return "palette"
	default:
		return string(technique) // Fallback to string conversion
	}
}
