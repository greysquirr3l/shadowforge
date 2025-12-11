package media

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
)

// TestNewMediaService tests service initialization
func TestNewMediaService(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
		service := NewMediaService(logger)

		assert.NotNil(t, service)
		assert.NotNil(t, service.detector)
		assert.NotNil(t, service.imageProcessor)
		assert.NotNil(t, service.audioProcessor)
		assert.NotNil(t, service.textProcessor)
		assert.NotNil(t, service.logger)
	})

	t.Run("WithNilLogger", func(t *testing.T) {
		service := NewMediaService(nil)
		assert.NotNil(t, service)
		// Service should handle nil logger gracefully
	})
}

// TestLoadMedia tests loading media assets
func TestLoadMedia(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewMediaService(logger)
	ctx := context.Background()

	t.Run("LoadImage_PNG_Success", func(t *testing.T) {
		// Create simple 2x2 PNG
		pngData := createTestPNG(2, 2)

		asset, err := service.LoadMedia(ctx, pngData, media.FormatPNG)

		require.NoError(t, err)
		require.NotNil(t, asset)
		assert.Equal(t, media.MediaTypeImage, asset.Type)
		assert.Equal(t, media.FormatPNG, asset.Format)
		assert.NotNil(t, asset.Dimensions)
		assert.Equal(t, 2, asset.Dimensions.Width)
		assert.Equal(t, 2, asset.Dimensions.Height)
	})

	t.Run("LoadImage_JPEG_Success", func(t *testing.T) {
		jpegData := createTestJPEG(4, 4)

		asset, err := service.LoadMedia(ctx, jpegData, media.FormatJPEG)

		require.NoError(t, err)
		require.NotNil(t, asset)
		assert.Equal(t, media.MediaTypeImage, asset.Type)
		assert.Equal(t, media.FormatJPEG, asset.Format)
		assert.NotNil(t, asset.Dimensions)
	})

	t.Run("LoadAudio_WAV_Success", func(t *testing.T) {
		wavData := createTestWAV16Bit()

		asset, err := service.LoadMedia(ctx, wavData, media.FormatWAV)

		require.NoError(t, err)
		require.NotNil(t, asset)
		assert.Equal(t, media.MediaTypeAudio, asset.Type)
		assert.Equal(t, media.FormatWAV, asset.Format)
		assert.NotNil(t, asset.SampleRate)
		assert.Equal(t, 44100, asset.SampleRate.Hz)
		assert.Equal(t, 16, asset.SampleRate.BitDepth)
		assert.Equal(t, 2, asset.SampleRate.Channels)
	})

	t.Run("LoadAudio_FLAC_Success", func(t *testing.T) {
		// FLAC magic bytes
		flacData := []byte("fLaC" + string(make([]byte, 100)))

		asset, err := service.LoadMedia(ctx, flacData, media.FormatFLAC)

		require.NoError(t, err)
		require.NotNil(t, asset)
		assert.Equal(t, media.MediaTypeAudio, asset.Type)
		assert.Equal(t, media.FormatFLAC, asset.Format)
		// FLAC is read-only, no sample rate info
		assert.Nil(t, asset.SampleRate)
	})

	t.Run("LoadText_TXT_Success", func(t *testing.T) {
		textData := []byte("Hello, this is valid UTF-8 text content!")

		asset, err := service.LoadMedia(ctx, textData, media.FormatTXT)

		require.NoError(t, err)
		require.NotNil(t, asset)
		assert.Equal(t, media.MediaTypeText, asset.Type)
		assert.Equal(t, media.FormatTXT, asset.Format)
	})

	t.Run("LoadText_Markdown_Success", func(t *testing.T) {
		mdData := []byte("# Markdown Title\n\n- Item 1\n- Item 2")

		asset, err := service.LoadMedia(ctx, mdData, media.FormatMarkdown)

		require.NoError(t, err)
		require.NotNil(t, asset)
		assert.Equal(t, media.MediaTypeText, asset.Type)
		assert.Equal(t, media.FormatMarkdown, asset.Format)
	})

	t.Run("EmptyData_Error", func(t *testing.T) {
		asset, err := service.LoadMedia(ctx, []byte{}, media.FormatPNG)

		assert.Nil(t, asset)
		assert.ErrorIs(t, err, media.ErrEmptyMediaData)
	})

	t.Run("InvalidFormat_Error", func(t *testing.T) {
		pngData := createTestPNG(2, 2)

		asset, err := service.LoadMedia(ctx, pngData, media.MediaFormat("invalid"))

		assert.Nil(t, asset)
		assert.ErrorIs(t, err, media.ErrInvalidFormat)
	})

	t.Run("FormatMismatch_Error", func(t *testing.T) {
		pngData := createTestPNG(2, 2)

		// Try to load PNG as JPEG
		asset, err := service.LoadMedia(ctx, pngData, media.FormatJPEG)

		assert.Nil(t, asset)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "format mismatch")
	})

	t.Run("InvalidImageData_Error", func(t *testing.T) {
		invalidData := []byte("This is not valid image data")

		asset, err := service.LoadMedia(ctx, invalidData, media.FormatPNG)

		assert.Nil(t, asset)
		assert.Error(t, err)
	})

	t.Run("InvalidTextEncoding_Error", func(t *testing.T) {
		invalidUTF8 := []byte{0xFF, 0xFE, 0xFD, 0x00, 0x01}

		asset, err := service.LoadMedia(ctx, invalidUTF8, media.FormatTXT)

		assert.Nil(t, asset)
		assert.Error(t, err)
	})

	t.Run("CancelledContext_Error", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		pngData := createTestPNG(2, 2)
		asset, err := service.LoadMedia(cancelledCtx, pngData, media.FormatPNG)

		assert.Nil(t, asset)
		assert.Error(t, err)
	})
}

// TestDetectFormat tests automatic format detection
func TestDetectFormat(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewMediaService(logger)
	ctx := context.Background()

	t.Run("DetectPNG", func(t *testing.T) {
		pngData := createTestPNG(2, 2)

		format, err := service.DetectFormat(ctx, pngData)

		require.NoError(t, err)
		assert.Equal(t, media.FormatPNG, format)
	})

	t.Run("DetectJPEG", func(t *testing.T) {
		jpegData := createTestJPEG(4, 4)

		format, err := service.DetectFormat(ctx, jpegData)

		require.NoError(t, err)
		assert.Equal(t, media.FormatJPEG, format)
	})

	t.Run("DetectWAV", func(t *testing.T) {
		wavData := createTestWAV16Bit()

		format, err := service.DetectFormat(ctx, wavData)

		require.NoError(t, err)
		assert.Equal(t, media.FormatWAV, format)
	})

	t.Run("DetectText", func(t *testing.T) {
		textData := []byte("Plain text content")

		format, err := service.DetectFormat(ctx, textData)

		require.NoError(t, err)
		assert.Equal(t, media.FormatTXT, format)
	})

	t.Run("EmptyData_Error", func(t *testing.T) {
		format, err := service.DetectFormat(ctx, []byte{})

		assert.Equal(t, media.MediaFormat(""), format)
		assert.ErrorIs(t, err, media.ErrEmptyMediaData)
	})

	t.Run("UnknownFormat_Error", func(t *testing.T) {
		unknownData := []byte{0x00, 0x01, 0x02, 0x03, 0x04}

		format, err := service.DetectFormat(ctx, unknownData)

		assert.Equal(t, media.MediaFormat(""), format)
		assert.Error(t, err)
	})

	t.Run("CancelledContext_Error", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		pngData := createTestPNG(2, 2)
		format, err := service.DetectFormat(cancelledCtx, pngData)

		assert.Equal(t, media.MediaFormat(""), format)
		assert.Error(t, err)
	})
}

// TestCalculateCapacity tests capacity calculations
func TestCalculateCapacity(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewMediaService(logger)
	ctx := context.Background()

	t.Run("ImageCapacity_Success", func(t *testing.T) {
		pngData := createTestPNG(100, 100)
		asset, _ := service.LoadMedia(ctx, pngData, media.FormatPNG)

		capacityInfo, err := service.CalculateCapacity(ctx, asset)

		require.NoError(t, err)
		require.NotNil(t, capacityInfo)
		assert.Equal(t, asset.ID, capacityInfo.AssetID)
		assert.Equal(t, media.MediaTypeImage, capacityInfo.MediaType)

		// 100x100 = 10,000 pixels, LSB = 10,000 bits = 1,250 bytes
		assert.Equal(t, int64(1250), capacityInfo.TotalCapacity)
		assert.Equal(t, int64(875), capacityInfo.UsableCapacity) // 70%
		assert.Equal(t, int64(625), capacityInfo.RecommendedMax) // 50%
	})

	t.Run("AudioCapacity_WAV_Success", func(t *testing.T) {
		wavData := createTestWAV16Bit()
		asset, _ := service.LoadMedia(ctx, wavData, media.FormatWAV)

		capacityInfo, err := service.CalculateCapacity(ctx, asset)

		require.NoError(t, err)
		require.NotNil(t, capacityInfo)
		assert.Equal(t, media.MediaTypeAudio, capacityInfo.MediaType)
		assert.Greater(t, capacityInfo.TotalCapacity, int64(0))
	})

	t.Run("TextCapacity_Success", func(t *testing.T) {
		textData := []byte("This is sample text with multiple words for testing capacity calculation.")
		asset, _ := service.LoadMedia(ctx, textData, media.FormatTXT)

		capacityInfo, err := service.CalculateCapacity(ctx, asset)

		require.NoError(t, err)
		require.NotNil(t, capacityInfo)
		assert.Equal(t, media.MediaTypeText, capacityInfo.MediaType)
		assert.Greater(t, capacityInfo.TotalCapacity, int64(0))
	})

	t.Run("NilAsset_Error", func(t *testing.T) {
		capacityInfo, err := service.CalculateCapacity(ctx, nil)

		assert.Nil(t, capacityInfo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil asset")
	})

	t.Run("CancelledContext_Error", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		pngData := createTestPNG(10, 10)
		asset, _ := service.LoadMedia(ctx, pngData, media.FormatPNG)

		capacityInfo, err := service.CalculateCapacity(cancelledCtx, asset)

		assert.Nil(t, capacityInfo)
		assert.Error(t, err)
	})

	t.Run("ImageWithoutDimensions_ZeroCapacity", func(t *testing.T) {
		// Create asset without dimensions
		assetID := media.NewAssetID()
		asset, _ := media.NewMediaAsset(assetID, media.MediaTypeImage, media.FormatPNG, []byte("data"))

		capacityInfo, err := service.CalculateCapacity(ctx, asset)

		require.NoError(t, err)
		assert.Equal(t, int64(0), capacityInfo.TotalCapacity)
	})

	t.Run("AudioWithoutSampleRate_ZeroCapacity", func(t *testing.T) {
		// Create asset without sample rate
		assetID := media.NewAssetID()
		asset, _ := media.NewMediaAsset(assetID, media.MediaTypeAudio, media.FormatFLAC, []byte("data"))

		capacityInfo, err := service.CalculateCapacity(ctx, asset)

		require.NoError(t, err)
		assert.Equal(t, int64(0), capacityInfo.TotalCapacity)
	})
}

// TestSanitizeMetadata tests metadata sanitization
func TestSanitizeMetadata(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewMediaService(logger)
	ctx := context.Background()

	t.Run("SanitizeImage_Success", func(t *testing.T) {
		pngData := createTestPNG(10, 10)
		asset, _ := service.LoadMedia(ctx, pngData, media.FormatPNG)

		originalSanitized := asset.IsSanitized

		err := service.SanitizeMetadata(ctx, asset)

		require.NoError(t, err)
		assert.True(t, asset.IsSanitized)
		assert.NotEqual(t, originalSanitized, asset.IsSanitized)
		// ModifiedAt should be updated
		assert.False(t, asset.ModifiedAt.IsZero())
	})

	t.Run("SanitizeAudio_Skipped", func(t *testing.T) {
		wavData := createTestWAV16Bit()
		asset, _ := service.LoadMedia(ctx, wavData, media.FormatWAV)

		err := service.SanitizeMetadata(ctx, asset)

		require.NoError(t, err)
		// Audio sanitization not supported yet
		assert.False(t, asset.IsSanitized)
	})

	t.Run("SanitizeText_Skipped", func(t *testing.T) {
		textData := []byte("Plain text")
		asset, _ := service.LoadMedia(ctx, textData, media.FormatTXT)

		err := service.SanitizeMetadata(ctx, asset)

		require.NoError(t, err)
		// Text sanitization not supported yet
		assert.False(t, asset.IsSanitized)
	})

	t.Run("NilAsset_Error", func(t *testing.T) {
		err := service.SanitizeMetadata(ctx, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil asset")
	})

	t.Run("CancelledContext_Error", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		pngData := createTestPNG(10, 10)
		asset, _ := service.LoadMedia(ctx, pngData, media.FormatPNG)

		err := service.SanitizeMetadata(cancelledCtx, asset)

		assert.Error(t, err)
	})
}

// TestValidateMedia tests media validation
func TestValidateMedia(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewMediaService(logger)
	ctx := context.Background()

	t.Run("ValidateImage_Success", func(t *testing.T) {
		pngData := createTestPNG(10, 10)
		asset, _ := service.LoadMedia(ctx, pngData, media.FormatPNG)

		err := service.ValidateMedia(ctx, asset)

		require.NoError(t, err)
		assert.True(t, asset.FormatValidated)
	})

	t.Run("ValidateAudio_Success", func(t *testing.T) {
		wavData := createTestWAV16Bit()
		asset, _ := service.LoadMedia(ctx, wavData, media.FormatWAV)

		err := service.ValidateMedia(ctx, asset)

		require.NoError(t, err)
		assert.True(t, asset.FormatValidated)
	})

	t.Run("ValidateText_Success", func(t *testing.T) {
		textData := []byte("Valid UTF-8 text")
		asset, _ := service.LoadMedia(ctx, textData, media.FormatTXT)

		err := service.ValidateMedia(ctx, asset)

		require.NoError(t, err)
		assert.True(t, asset.FormatValidated)
	})

	t.Run("NilAsset_Error", func(t *testing.T) {
		err := service.ValidateMedia(ctx, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil asset")
	})

	t.Run("EmptyAssetData_Error", func(t *testing.T) {
		assetID := media.NewAssetID()
		// Create asset with 1 byte, then clear it to bypass NewMediaAsset validation
		asset, _ := media.NewMediaAsset(assetID, media.MediaTypeImage, media.FormatPNG, []byte{0x00})
		asset.Data = []byte{} // Clear data after creation

		err := service.ValidateMedia(ctx, asset)

		assert.ErrorIs(t, err, media.ErrEmptyMediaData)
	})

	t.Run("InvalidImageData_Error", func(t *testing.T) {
		assetID := media.NewAssetID()
		invalidData := []byte("not valid image data")
		asset, _ := media.NewMediaAsset(assetID, media.MediaTypeImage, media.FormatPNG, invalidData)

		err := service.ValidateMedia(ctx, asset)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid image data")
	})

	t.Run("InvalidTextEncoding_Error", func(t *testing.T) {
		assetID := media.NewAssetID()
		invalidUTF8 := []byte{0xFF, 0xFE, 0xFD}
		asset, _ := media.NewMediaAsset(assetID, media.MediaTypeText, media.FormatTXT, invalidUTF8)

		err := service.ValidateMedia(ctx, asset)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid text encoding")
	})

	t.Run("CancelledContext_Error", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		pngData := createTestPNG(10, 10)
		asset, _ := service.LoadMedia(ctx, pngData, media.FormatPNG)

		err := service.ValidateMedia(cancelledCtx, asset)

		assert.Error(t, err)
	})
}

// TestAnalyzeQuality tests quality analysis
func TestAnalyzeQuality(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewMediaService(logger)
	ctx := context.Background()

	t.Run("AnalyzeImageQuality_Success", func(t *testing.T) {
		// Large image = higher quality
		pngData := createTestPNG(1920, 1080)
		asset, _ := service.LoadMedia(ctx, pngData, media.FormatPNG)

		quality, err := service.AnalyzeQuality(ctx, asset)

		require.NoError(t, err)
		assert.Greater(t, quality, 0.0)
		assert.LessOrEqual(t, quality, 1.0)
		// Large image should have decent quality
		assert.Greater(t, quality, 0.5)
	})

	t.Run("AnalyzeAudioQuality_WAV_Success", func(t *testing.T) {
		wavData := createTestWAV16Bit()
		asset, _ := service.LoadMedia(ctx, wavData, media.FormatWAV)

		quality, err := service.AnalyzeQuality(ctx, asset)

		require.NoError(t, err)
		assert.Greater(t, quality, 0.0)
		assert.LessOrEqual(t, quality, 1.0)
	})

	t.Run("AnalyzeAudioQuality_FLAC_DefaultQuality", func(t *testing.T) {
		flacData := []byte("fLaC" + string(make([]byte, 100)))
		asset, _ := service.LoadMedia(ctx, flacData, media.FormatFLAC)

		quality, err := service.AnalyzeQuality(ctx, asset)

		require.NoError(t, err)
		assert.Equal(t, 0.5, quality) // Default for non-WAV
	})

	t.Run("AnalyzeTextQuality_SmallText", func(t *testing.T) {
		textData := []byte("Short text")
		asset, _ := service.LoadMedia(ctx, textData, media.FormatTXT)

		quality, err := service.AnalyzeQuality(ctx, asset)

		require.NoError(t, err)
		assert.Equal(t, 0.3, quality) // Small text = low quality
	})

	t.Run("AnalyzeTextQuality_LargeText", func(t *testing.T) {
		// Create large text (>10000 words)
		largeText := make([]byte, 100000)
		for i := range largeText {
			largeText[i] = 'a'
		}
		asset, _ := service.LoadMedia(ctx, largeText, media.FormatTXT)

		quality, err := service.AnalyzeQuality(ctx, asset)

		require.NoError(t, err)
		assert.Equal(t, 0.9, quality) // Large text = high quality
	})

	t.Run("NilAsset_Error", func(t *testing.T) {
		quality, err := service.AnalyzeQuality(ctx, nil)

		assert.Equal(t, 0.0, quality)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil asset")
	})

	t.Run("CancelledContext_Error", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		pngData := createTestPNG(10, 10)
		asset, _ := service.LoadMedia(ctx, pngData, media.FormatPNG)

		quality, err := service.AnalyzeQuality(cancelledCtx, asset)

		assert.Equal(t, 0.0, quality)
		assert.Error(t, err)
	})
}

// Helper functions to create test media data
// These match the helpers in other test files

func createTestPNG(width, height int) []byte {
	imageProcessor := NewImageProcessor()
	img, _ := imageProcessor.CreateBlankImage(width, height, media.FormatPNG)
	data, _ := imageProcessor.SaveImage(img, media.FormatPNG)
	return data
}

func createTestJPEG(width, height int) []byte {
	imageProcessor := NewImageProcessor()
	img, _ := imageProcessor.CreateBlankImage(width, height, media.FormatJPEG)
	data, _ := imageProcessor.SaveImage(img, media.FormatJPEG)
	return data
}

func createTestWAV16Bit() []byte {
	audioProcessor := NewAudioProcessor()

	// Create silence: 0.1 seconds, 44.1kHz, 16-bit, stereo
	pcmData, _ := audioProcessor.CreateSilence(44100, 16, 2, 100)
	data, _ := audioProcessor.SaveWAV(pcmData)
	return data
}
