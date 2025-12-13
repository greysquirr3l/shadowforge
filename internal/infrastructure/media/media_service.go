package media

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
)

// MediaService implements the domain media.Service interface.
// It coordinates between FormatDetector and media-specific processors
// while maintaining security through validation and proper error handling.
type MediaService struct {
	detector       *FormatDetector
	imageProcessor *ImageProcessor
	audioProcessor *AudioProcessor
	textProcessor  *TextProcessor
	logger         *logrus.Logger
}

// NewMediaService creates a new MediaService.
func NewMediaService(logger *logrus.Logger) *MediaService {
	detector := NewFormatDetector()

	return &MediaService{
		detector:       detector,
		imageProcessor: NewImageProcessor(),
		audioProcessor: NewAudioProcessor(),
		textProcessor:  NewTextProcessor(),
		logger:         logger,
	}
}

// LoadMedia loads media data and creates a MediaAsset.
// This method validates all inputs and ensures proper domain object creation.
func (s *MediaService) LoadMedia(ctx context.Context, data []byte, format media.MediaFormat) (*media.MediaAsset, error) {
	// Security: Check context first
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Security: Validate inputs
	if len(data) == 0 {
		return nil, media.ErrEmptyMediaData
	}

	if !format.IsValid() {
		return nil, media.ErrInvalidFormat
	}

	s.logger.Debug("loading media",
		slog.String("format", string(format)),
		slog.Int("size", len(data)))

	// Route to appropriate handler based on format
	switch format {
	case media.FormatPNG, media.FormatJPEG, media.FormatGIF, media.FormatBMP:
		return s.loadImageMedia(ctx, data, format)
	case media.FormatWAV, media.FormatFLAC, media.FormatMP3:
		return s.loadAudioMedia(ctx, data, format)
	case media.FormatTXT, media.FormatMarkdown:
		return s.loadTextMedia(ctx, data, format)
	default:
		return nil, media.ErrInvalidFormat
	}
}

// loadImageMedia handles image loading with full security validation.
func (s *MediaService) loadImageMedia(ctx context.Context, data []byte, expectedFormat media.MediaFormat) (*media.MediaAsset, error) {
	// Security: Load and validate image format
	img, detectedFormat, err := s.imageProcessor.LoadImage(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Security: Verify format matches expected
	if expectedFormat != detectedFormat {
		return nil, fmt.Errorf("format mismatch: expected %s, got %s", expectedFormat, detectedFormat)
	}

	// Security: Get metadata through safe parser
	imageInfo, err := s.imageProcessor.GetImageInfo(data)
	if err != nil {
		return nil, fmt.Errorf("failed to get image info: %w", err)
	}

	// Security: Create asset with proper validation
	assetID := media.NewAssetID()
	asset, err := media.NewMediaAsset(assetID, media.MediaTypeImage, detectedFormat, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	// Security: Create and validate dimensions value object
	dimensions, err := media.NewDimensions(imageInfo.Width, imageInfo.Height)
	if err != nil {
		return nil, fmt.Errorf("invalid dimensions: %w", err)
	}

	// Security: Set dimensions through domain method (enforces business rules)
	if err := asset.SetDimensions(dimensions); err != nil {
		return nil, fmt.Errorf("failed to set dimensions: %w", err)
	}

	// Security: Set color space through domain method
	if err := asset.SetColorSpace(&imageInfo.ColorSpace); err != nil {
		return nil, fmt.Errorf("failed to set color space: %w", err)
	}

	s.logger.Info("image media loaded",
		slog.String("asset_id", asset.ID.String()),
		slog.String("format", string(detectedFormat)),
		slog.Int("width", imageInfo.Width),
		slog.Int("height", imageInfo.Height))

	// Security: Ensure image object is properly disposed
	_ = img // image.Image is used for validation only

	return asset, nil
}

// loadAudioMedia handles audio loading with security validation.
func (s *MediaService) loadAudioMedia(ctx context.Context, data []byte, format media.MediaFormat) (*media.MediaAsset, error) {
	// Security: Create asset first with validation
	assetID := media.NewAssetID()
	asset, err := media.NewMediaAsset(assetID, media.MediaTypeAudio, format, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	// Security: For non-WAV formats, validate only (read-only)
	if format != media.FormatWAV {
		if err := s.audioProcessor.ValidateAudio(data); err != nil {
			return nil, fmt.Errorf("invalid audio data: %w", err)
		}

		s.logger.Info("audio media loaded (read-only)",
			slog.String("asset_id", asset.ID.String()),
			slog.String("format", string(format)))

		return asset, nil
	}

	// Security: For WAV, load and validate full PCM data
	pcmData, err := s.audioProcessor.LoadWAV(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load WAV: %w", err)
	}

	// Security: Create SampleRate value object with validation
	sampleRate, err := media.NewSampleRate(pcmData.SampleRate, pcmData.BitsPerSample, pcmData.Channels)
	if err != nil {
		return nil, fmt.Errorf("invalid sample rate parameters: %w", err)
	}

	// Security: Set sample rate through domain method (enforces business rules)
	if err := asset.SetSampleRate(sampleRate); err != nil {
		return nil, fmt.Errorf("failed to set sample rate: %w", err)
	}

	s.logger.Info("audio media loaded",
		slog.String("asset_id", asset.ID.String()),
		slog.String("format", string(format)),
		slog.Int("sample_rate", pcmData.SampleRate),
		slog.Int("bit_depth", pcmData.BitsPerSample),
		slog.Int("channels", pcmData.Channels))

	return asset, nil
}

// loadTextMedia handles text loading with encoding validation.
func (s *MediaService) loadTextMedia(ctx context.Context, data []byte, format media.MediaFormat) (*media.MediaAsset, error) {
	// Security: Validate encoding before creating asset
	if err := s.textProcessor.ValidateText(data); err != nil {
		return nil, fmt.Errorf("invalid text encoding: %w", err)
	}

	// Security: Create asset with validation
	assetID := media.NewAssetID()
	asset, err := media.NewMediaAsset(assetID, media.MediaTypeText, format, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	s.logger.Info("text media loaded",
		slog.String("asset_id", asset.ID.String()),
		slog.String("format", string(format)),
		slog.Int("size", len(data)))

	return asset, nil
}

// DetectFormat automatically detects the media format from raw data.
// Security: Uses magic byte detection with validation.
func (s *MediaService) DetectFormat(ctx context.Context, data []byte) (media.MediaFormat, error) {
	// Security: Check context
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Security: Validate input
	if len(data) == 0 {
		return "", media.ErrEmptyMediaData
	}

	// Security: Detect format using magic bytes (secure)
	format, mediaType, err := s.detector.DetectFormat(data)
	if err != nil {
		return "", fmt.Errorf("format detection failed: %w", err)
	}

	s.logger.Debug("format detected",
		slog.String("format", string(format)),
		slog.String("type", string(mediaType)),
		slog.Int("data_size", len(data)))

	return format, nil
}

// CalculateCapacity calculates the embedding capacity for a media asset.
// Security: Validates asset before calculation.
func (s *MediaService) CalculateCapacity(ctx context.Context, asset *media.MediaAsset) (*media.CapacityInfo, error) {
	// Security: Check context
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Security: Validate asset
	if asset == nil {
		return nil, fmt.Errorf("nil asset")
	}

	// Calculate capacity based on media type
	var totalCapacity int64
	switch asset.Type {
	case media.MediaTypeImage:
		totalCapacity = s.calculateImageCapacity(asset)
	case media.MediaTypeAudio:
		totalCapacity = s.calculateAudioCapacity(asset)
	case media.MediaTypeText:
		totalCapacity = s.calculateTextCapacity(asset)
	default:
		return nil, media.ErrInvalidFormat
	}

	// Security: Apply conservative safety factors
	usableCapacity := int64(float64(totalCapacity) * 0.7) // 70% usable
	recommendedMax := int64(float64(totalCapacity) * 0.5) // 50% recommended

	// Security: Create capacity info with validation
	capacityInfo, err := media.NewCapacityInfo(
		asset.ID,
		asset.Type,
		totalCapacity,
		usableCapacity,
		recommendedMax,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create capacity info: %w", err)
	}

	s.logger.Debug("capacity calculated",
		slog.String("asset_id", asset.ID.String()),
		slog.Int64("total", totalCapacity),
		slog.Int64("usable", usableCapacity),
		slog.Int64("recommended", recommendedMax))

	return capacityInfo, nil
}

// calculateImageCapacity calculates capacity for image assets.
// Security: Conservative estimation based on LSB embedding.
func (s *MediaService) calculateImageCapacity(asset *media.MediaAsset) int64 {
	if asset.Dimensions == nil {
		return 0
	}

	totalPixels := asset.Dimensions.Area()
	// Security: Use LSB (1 bit per pixel) as baseline
	return int64(totalPixels) / 8 // bits to bytes
}

// calculateAudioCapacity calculates capacity for audio assets.
// Security: Conservative estimation based on LSB embedding.
func (s *MediaService) calculateAudioCapacity(asset *media.MediaAsset) int64 {
	if asset.SampleRate == nil {
		return 0
	}

	// Security: Validate sample rate components
	if asset.SampleRate.Hz == 0 || asset.SampleRate.BitDepth == 0 || asset.SampleRate.Channels == 0 {
		return 0
	}

	// Security: Calculate estimated samples safely
	bytesPerSample := asset.SampleRate.BitDepth / 8
	if bytesPerSample == 0 {
		return 0
	}

	estimatedSamples := int64(len(asset.Data)) / int64(bytesPerSample) / int64(asset.SampleRate.Channels)
	// Security: Use LSB (1 bit per sample) as baseline
	return estimatedSamples / 8 // bits to bytes
}

// calculateTextCapacity calculates capacity for text assets.
// Security: Very conservative estimation.
func (s *MediaService) calculateTextCapacity(asset *media.MediaAsset) int64 {
	// Security: Estimate words conservatively
	wordCount := len(asset.Data) / 5 // rough average
	// Security: Use 1 bit per word boundary as baseline
	return int64(wordCount) / 8 // bits to bytes
}

// SanitizeMetadata removes all metadata from a media asset.
// Security: Only applies to images; logs sanitization.
func (s *MediaService) SanitizeMetadata(ctx context.Context, asset *media.MediaAsset) error {
	// Security: Check context
	if err := ctx.Err(); err != nil {
		return err
	}

	// Security: Validate asset
	if asset == nil {
		return fmt.Errorf("nil asset")
	}

	// Security: Only images support metadata sanitization currently
	if asset.Type != media.MediaTypeImage {
		s.logger.Debug("metadata sanitization skipped",
			slog.String("type", string(asset.Type)),
			slog.String("reason", "not supported for this type"))
		return nil
	}

	// Security: Sanitize through processor
	sanitizedData, err := s.imageProcessor.SanitizeMetadata(asset.Data)
	if err != nil {
		return fmt.Errorf("failed to sanitize image metadata: %w", err)
	}

	// Security: Update asset data
	asset.Data = sanitizedData
	asset.IsSanitized = true
	asset.ModifiedAt = time.Now()

	s.logger.Info("metadata sanitized",
		slog.String("asset_id", asset.ID.String()),
		slog.Int("original_size", len(asset.Data)),
		slog.Int("sanitized_size", len(sanitizedData)))

	return nil
}

// ValidateMedia validates the integrity of a media asset.
// Security: Performs format-specific validation.
func (s *MediaService) ValidateMedia(ctx context.Context, asset *media.MediaAsset) error {
	// Security: Check context
	if err := ctx.Err(); err != nil {
		return err
	}

	// Security: Validate asset
	if asset == nil {
		return fmt.Errorf("nil asset")
	}

	if len(asset.Data) == 0 {
		return media.ErrEmptyMediaData
	}

	// Security: Perform type-specific validation
	switch asset.Type {
	case media.MediaTypeImage:
		_, _, err := s.imageProcessor.LoadImage(asset.Data)
		if err != nil {
			return fmt.Errorf("invalid image data: %w", err)
		}

	case media.MediaTypeAudio:
		if err := s.audioProcessor.ValidateAudio(asset.Data); err != nil {
			return fmt.Errorf("invalid audio data: %w", err)
		}

	case media.MediaTypeText:
		if err := s.textProcessor.ValidateText(asset.Data); err != nil {
			return fmt.Errorf("invalid text encoding: %w", err)
		}

	default:
		return media.ErrInvalidFormat
	}

	// Security: Mark as validated
	asset.FormatValidated = true

	s.logger.Debug("media validated",
		slog.String("asset_id", asset.ID.String()),
		slog.String("type", string(asset.Type)))

	return nil
}

// AnalyzeQuality analyzes the quality score of a media asset.
// Security: Returns conservative quality scores.
func (s *MediaService) AnalyzeQuality(ctx context.Context, asset *media.MediaAsset) (float64, error) {
	// Security: Check context
	if err := ctx.Err(); err != nil {
		return 0.0, err
	}

	// Security: Validate asset
	if asset == nil {
		return 0.0, fmt.Errorf("nil asset")
	}

	var quality float64

	// Security: Analyze quality based on media type
	switch asset.Type {
	case media.MediaTypeImage:
		// Security: Get image info safely
		imageInfo, err := s.imageProcessor.GetImageInfo(asset.Data)
		if err != nil {
			return 0.0, fmt.Errorf("failed to get image info for quality analysis: %w", err)
		}

		totalPixels := int64(imageInfo.Width * imageInfo.Height)
		quality = s.imageProcessor.calculateQualityScore(imageInfo, totalPixels)

	case media.MediaTypeAudio:
		// Security: Only WAV supports quality analysis currently
		if asset.Format == media.FormatWAV {
			audioInfo, err := s.audioProcessor.GetAudioInfo(asset.Data)
			if err != nil {
				return 0.0, fmt.Errorf("failed to get audio info for quality analysis: %w", err)
			}

			// Security: Calculate samples safely
			pcmData, loadErr := s.audioProcessor.LoadWAV(asset.Data)
			if loadErr != nil {
				return 0.0, fmt.Errorf("failed to load WAV for quality analysis: %w", loadErr)
			}

			totalSamples := int64(len(pcmData.Samples[0]))
			quality = s.audioProcessor.calculateQualityScore(audioInfo, totalSamples)
		} else {
			// Security: Default quality for non-WAV
			quality = 0.5
		}

	case media.MediaTypeText:
		// Security: Simple quality metric for text - longer texts are higher quality
		wordCount := len([]rune(string(asset.Data))) / 5 // Rough word estimate
		if wordCount < 100 {
			quality = 0.3
		} else if wordCount < 1000 {
			quality = 0.5
		} else if wordCount < 10000 {
			quality = 0.7
		} else {
			quality = 0.9
		}

	default:
		return 0.0, media.ErrInvalidFormat
	}

	s.logger.Debug("quality analyzed",
		slog.String("asset_id", asset.ID.String()),
		slog.Float64("quality", quality))

	return quality, nil
}
