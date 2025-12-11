package media

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/bmp"
)

// TestNewImageProcessor tests the constructor.
func TestNewImageProcessor(t *testing.T) {
	processor := NewImageProcessor()
	require.NotNil(t, processor)
	assert.NotNil(t, processor.detector)
}

// TestImageProcessor_LoadImage_PNG tests loading PNG images.
func TestImageProcessor_LoadImage_PNG(t *testing.T) {
	processor := NewImageProcessor()

	// Create a minimal PNG
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(5, 5, color.RGBA{255, 0, 0, 255})

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)

	// Load the PNG
	loadedImg, format, err := processor.LoadImage(buf.Bytes())
	require.NoError(t, err)
	assert.Equal(t, media.FormatPNG, format)
	assert.NotNil(t, loadedImg)
	assert.Equal(t, 10, loadedImg.Bounds().Dx())
	assert.Equal(t, 10, loadedImg.Bounds().Dy())
}

// TestImageProcessor_LoadImage_JPEG tests loading JPEG images.
func TestImageProcessor_LoadImage_JPEG(t *testing.T) {
	processor := NewImageProcessor()

	// Create a minimal JPEG
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	require.NoError(t, err)

	// Load the JPEG
	loadedImg, format, err := processor.LoadImage(buf.Bytes())
	require.NoError(t, err)
	assert.Equal(t, media.FormatJPEG, format)
	assert.NotNil(t, loadedImg)
}

// TestImageProcessor_LoadImage_GIF tests loading GIF images.
func TestImageProcessor_LoadImage_GIF(t *testing.T) {
	processor := NewImageProcessor()

	// Create a minimal GIF
	palette := color.Palette{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{255, 255, 255, 255},
		color.RGBA{255, 0, 0, 255},
	}
	img := image.NewPaletted(image.Rect(0, 0, 10, 10), palette)
	img.SetColorIndex(5, 5, 2)

	var buf bytes.Buffer
	err := gif.Encode(&buf, img, nil)
	require.NoError(t, err)

	// Load the GIF
	loadedImg, format, err := processor.LoadImage(buf.Bytes())
	require.NoError(t, err)
	assert.Equal(t, media.FormatGIF, format)
	assert.NotNil(t, loadedImg)
}

// TestImageProcessor_LoadImage_BMP tests loading BMP images.
func TestImageProcessor_LoadImage_BMP(t *testing.T) {
	processor := NewImageProcessor()

	// Create a minimal BMP
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(5, 5, color.RGBA{0, 255, 0, 255})

	var buf bytes.Buffer
	err := bmp.Encode(&buf, img)
	require.NoError(t, err)

	// Load the BMP
	loadedImg, format, err := processor.LoadImage(buf.Bytes())
	require.NoError(t, err)
	assert.Equal(t, media.FormatBMP, format)
	assert.NotNil(t, loadedImg)
}

// TestImageProcessor_LoadImage_EmptyData tests loading with empty data.
func TestImageProcessor_LoadImage_EmptyData(t *testing.T) {
	processor := NewImageProcessor()

	img, format, err := processor.LoadImage([]byte{})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInsufficientData)
	assert.Nil(t, img)
	assert.Equal(t, media.MediaFormat(""), format)
}

// TestImageProcessor_LoadImage_InvalidFormat tests loading non-image data.
func TestImageProcessor_LoadImage_InvalidFormat(t *testing.T) {
	processor := NewImageProcessor()

	// Audio data (WAV header)
	wavData := []byte("RIFF\x00\x00\x00\x00WAVE")

	img, format, err := processor.LoadImage(wavData)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedFormat)
	assert.Nil(t, img)
	assert.Equal(t, media.MediaFormat(""), format)
}

// TestImageProcessor_SaveImage_PNG tests saving PNG images.
func TestImageProcessor_SaveImage_PNG(t *testing.T) {
	processor := NewImageProcessor()

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(5, 5, color.RGBA{255, 0, 0, 255})

	// Save as PNG
	data, err := processor.SaveImage(img, media.FormatPNG)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify it's a valid PNG
	assert.True(t, bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}))

	// Verify it can be loaded back
	loadedImg, format, err := processor.LoadImage(data)
	require.NoError(t, err)
	assert.Equal(t, media.FormatPNG, format)
	assert.Equal(t, img.Bounds(), loadedImg.Bounds())
}

// TestImageProcessor_SaveImage_JPEG tests saving JPEG images.
func TestImageProcessor_SaveImage_JPEG(t *testing.T) {
	processor := NewImageProcessor()

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}

	// Save as JPEG
	data, err := processor.SaveImage(img, media.FormatJPEG)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify it's a valid JPEG (starts with FF D8)
	assert.True(t, bytes.HasPrefix(data, []byte{0xFF, 0xD8}))
}

// TestImageProcessor_SaveImage_GIF tests saving GIF images.
func TestImageProcessor_SaveImage_GIF(t *testing.T) {
	processor := NewImageProcessor()

	// Create paletted image for GIF
	palette := color.Palette{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{255, 255, 255, 255},
	}
	img := image.NewPaletted(image.Rect(0, 0, 10, 10), palette)

	// Save as GIF
	data, err := processor.SaveImage(img, media.FormatGIF)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify it's a valid GIF
	assert.True(t, bytes.HasPrefix(data, []byte("GIF89a")))
}

// TestImageProcessor_SaveImage_BMP tests saving BMP images.
func TestImageProcessor_SaveImage_BMP(t *testing.T) {
	processor := NewImageProcessor()

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// Save as BMP
	data, err := processor.SaveImage(img, media.FormatBMP)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify it's a valid BMP (starts with "BM")
	assert.True(t, bytes.HasPrefix(data, []byte("BM")))
}

// TestImageProcessor_SaveImage_NilImage tests saving nil image.
func TestImageProcessor_SaveImage_NilImage(t *testing.T) {
	processor := NewImageProcessor()

	data, err := processor.SaveImage(nil, media.FormatPNG)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "nil")
}

// TestImageProcessor_SaveImage_UnsupportedFormat tests saving with unsupported format.
func TestImageProcessor_SaveImage_UnsupportedFormat(t *testing.T) {
	processor := NewImageProcessor()

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	data, err := processor.SaveImage(img, media.FormatWAV)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedFormat)
	assert.Nil(t, data)
}

// TestImageProcessor_SanitizeMetadata tests metadata sanitization.
func TestImageProcessor_SanitizeMetadata(t *testing.T) {
	processor := NewImageProcessor()

	// Create a PNG with some data
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(5, 5, color.RGBA{255, 0, 0, 255})

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)
	originalData := buf.Bytes()

	// Sanitize (strips metadata by decode/encode cycle)
	sanitized, err := processor.SanitizeMetadata(originalData)
	require.NoError(t, err)
	assert.NotEmpty(t, sanitized)

	// Verify sanitized image is valid
	loadedImg, format, err := processor.LoadImage(sanitized)
	require.NoError(t, err)
	assert.Equal(t, media.FormatPNG, format)
	assert.Equal(t, img.Bounds(), loadedImg.Bounds())
}

// TestImageProcessor_SanitizeMetadata_InvalidData tests sanitizing invalid data.
func TestImageProcessor_SanitizeMetadata_InvalidData(t *testing.T) {
	processor := NewImageProcessor()

	sanitized, err := processor.SanitizeMetadata([]byte("not an image"))
	assert.Error(t, err)
	assert.Nil(t, sanitized)
}

// TestImageProcessor_CalculateCapacity_LSB tests capacity calculation for LSB technique.
func TestImageProcessor_CalculateCapacity_LSB(t *testing.T) {
	processor := NewImageProcessor()

	// Create a 100x100 PNG image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)

	capacity, err := processor.CalculateCapacity(buf.Bytes(), "lsb")
	require.NoError(t, err)
	assert.NotNil(t, capacity)

	// 100x100 = 10,000 pixels
	// Raw capacity: 10,000 * 4 bits (RGBA) / 8 = 5,000 bytes
	// Safe capacity: 10,000 * 0.8 / 8 = 1,000 bytes
	assert.Equal(t, int64(5000), capacity.TotalCapacity)
	assert.Equal(t, int64(1000), capacity.UsableCapacity)
	assert.Equal(t, media.MediaTypeImage, capacity.MediaType)
	assert.Greater(t, capacity.QualityImpact, 0.0)
	assert.LessOrEqual(t, capacity.QualityImpact, 1.0)
}

// TestImageProcessor_CalculateCapacity_LSB2 tests capacity for 2-bit LSB.
func TestImageProcessor_CalculateCapacity_LSB2(t *testing.T) {
	processor := NewImageProcessor()

	// Create a 100x100 PNG
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)

	capacity, err := processor.CalculateCapacity(buf.Bytes(), "lsb-2")
	require.NoError(t, err)
	assert.NotNil(t, capacity)

	// 100x100 = 10,000 pixels
	// Raw capacity: 10,000 * 8 bits (RGBA, 2 bits per channel) / 8 = 10,000 bytes
	// Safe capacity: 10,000 * 0.7 = 7,000 bytes
	assert.Equal(t, int64(10000), capacity.TotalCapacity)
	assert.Equal(t, int64(7000), capacity.UsableCapacity)
}

// TestImageProcessor_CalculateCapacity_DCT tests DCT capacity for JPEG.
func TestImageProcessor_CalculateCapacity_DCT(t *testing.T) {
	processor := NewImageProcessor()

	// Create a 128x128 JPEG (multiple of 8 for DCT blocks)
	img := image.NewRGBA(image.Rect(0, 0, 128, 128))
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	require.NoError(t, err)

	capacity, err := processor.CalculateCapacity(buf.Bytes(), "dct")
	require.NoError(t, err)
	assert.NotNil(t, capacity)

	// 128x128 = 16,384 pixels
	// Blocks: 16,384 / 64 = 256 blocks
	// Raw capacity: 256 / 8 = 32 bytes
	// Safe capacity: 32 * 0.5 = 16 bytes
	assert.Equal(t, int64(32), capacity.TotalCapacity)
	assert.Equal(t, int64(16), capacity.UsableCapacity)
}

// TestImageProcessor_CalculateCapacity_DCT_NonJPEG tests DCT on non-JPEG.
func TestImageProcessor_CalculateCapacity_DCT_NonJPEG(t *testing.T) {
	processor := NewImageProcessor()

	// Create a PNG
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)

	capacity, err := processor.CalculateCapacity(buf.Bytes(), "dct")
	assert.Error(t, err)
	assert.Nil(t, capacity)
	assert.Contains(t, err.Error(), "JPEG")
}

// TestImageProcessor_CalculateCapacity_Palette tests palette technique.
func TestImageProcessor_CalculateCapacity_Palette(t *testing.T) {
	processor := NewImageProcessor()

	// Create a GIF
	palette := color.Palette{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{255, 255, 255, 255},
	}
	img := image.NewPaletted(image.Rect(0, 0, 256, 256), palette)

	var buf bytes.Buffer
	err := gif.Encode(&buf, img, nil)
	require.NoError(t, err)

	capacity, err := processor.CalculateCapacity(buf.Bytes(), "palette")
	require.NoError(t, err)
	assert.NotNil(t, capacity)

	// 256x256 = 65,536 pixels
	// Raw capacity: 65,536 / 256 = 256 bytes
	// Safe capacity: 256 * 0.6 = 153 bytes
	assert.Equal(t, int64(256), capacity.TotalCapacity)
	assert.Equal(t, int64(153), capacity.UsableCapacity)
}

// TestImageProcessor_CalculateCapacity_UnknownTechnique tests unknown technique.
func TestImageProcessor_CalculateCapacity_UnknownTechnique(t *testing.T) {
	processor := NewImageProcessor()

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)

	capacity, err := processor.CalculateCapacity(buf.Bytes(), "unknown-technique")
	assert.Error(t, err)
	assert.Nil(t, capacity)
	assert.Contains(t, err.Error(), "unknown technique")
}

// TestImageProcessor_CalculateCapacity_InvalidData tests capacity with invalid data.
func TestImageProcessor_CalculateCapacity_InvalidData(t *testing.T) {
	processor := NewImageProcessor()

	capacity, err := processor.CalculateCapacity([]byte("not an image"), "lsb")
	assert.Error(t, err)
	assert.Nil(t, capacity)
}

// TestImageProcessor_GetImageInfo tests extracting image info.
func TestImageProcessor_GetImageInfo(t *testing.T) {
	processor := NewImageProcessor()

	// Create a 50x30 PNG
	img := image.NewRGBA(image.Rect(0, 0, 50, 30))
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)

	info, err := processor.GetImageInfo(buf.Bytes())
	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, 50, info.Width)
	assert.Equal(t, 30, info.Height)
	assert.Equal(t, media.FormatPNG, info.Format)
}

// TestImageProcessor_CreateBlankImage_PNG tests creating blank PNG.
func TestImageProcessor_CreateBlankImage_PNG(t *testing.T) {
	processor := NewImageProcessor()

	img, err := processor.CreateBlankImage(100, 200, media.FormatPNG)
	require.NoError(t, err)
	assert.NotNil(t, img)
	assert.Equal(t, 100, img.Bounds().Dx())
	assert.Equal(t, 200, img.Bounds().Dy())

	// Verify it's RGBA
	_, ok := img.(*image.RGBA)
	assert.True(t, ok)
}

// TestImageProcessor_CreateBlankImage_JPEG tests creating blank JPEG.
func TestImageProcessor_CreateBlankImage_JPEG(t *testing.T) {
	processor := NewImageProcessor()

	img, err := processor.CreateBlankImage(150, 100, media.FormatJPEG)
	require.NoError(t, err)
	assert.NotNil(t, img)
	assert.Equal(t, 150, img.Bounds().Dx())
	assert.Equal(t, 100, img.Bounds().Dy())
}

// TestImageProcessor_CreateBlankImage_GIF tests creating blank GIF.
func TestImageProcessor_CreateBlankImage_GIF(t *testing.T) {
	processor := NewImageProcessor()

	img, err := processor.CreateBlankImage(80, 60, media.FormatGIF)
	require.NoError(t, err)
	assert.NotNil(t, img)
	assert.Equal(t, 80, img.Bounds().Dx())
	assert.Equal(t, 60, img.Bounds().Dy())

	// Verify it's paletted
	palettedImg, ok := img.(*image.Paletted)
	assert.True(t, ok)
	assert.NotNil(t, palettedImg.Palette)
	assert.Len(t, palettedImg.Palette, 256)
}

// TestImageProcessor_CreateBlankImage_BMP tests creating blank BMP.
func TestImageProcessor_CreateBlankImage_BMP(t *testing.T) {
	processor := NewImageProcessor()

	img, err := processor.CreateBlankImage(120, 90, media.FormatBMP)
	require.NoError(t, err)
	assert.NotNil(t, img)
	assert.Equal(t, 120, img.Bounds().Dx())
	assert.Equal(t, 90, img.Bounds().Dy())
}

// TestImageProcessor_CreateBlankImage_InvalidDimensions tests invalid dimensions.
func TestImageProcessor_CreateBlankImage_InvalidDimensions(t *testing.T) {
	processor := NewImageProcessor()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"zero width", 0, 100},
		{"zero height", 100, 0},
		{"negative width", -10, 100},
		{"negative height", 100, -10},
		{"both zero", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := processor.CreateBlankImage(tt.width, tt.height, media.FormatPNG)
			assert.Error(t, err)
			assert.Nil(t, img)
			assert.Contains(t, err.Error(), "invalid dimensions")
		})
	}
}

// TestImageProcessor_CreateBlankImage_UnsupportedFormat tests unsupported format.
func TestImageProcessor_CreateBlankImage_UnsupportedFormat(t *testing.T) {
	processor := NewImageProcessor()

	img, err := processor.CreateBlankImage(100, 100, media.FormatWAV)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedFormat)
	assert.Nil(t, img)
}

// TestImageProcessor_ValidateImage_PNG tests PNG validation.
func TestImageProcessor_ValidateImage_PNG(t *testing.T) {
	processor := NewImageProcessor()

	// Create valid PNG
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)

	err = processor.ValidateImage(buf.Bytes())
	assert.NoError(t, err)
}

// TestImageProcessor_ValidateImage_JPEG tests JPEG validation.
func TestImageProcessor_ValidateImage_JPEG(t *testing.T) {
	processor := NewImageProcessor()

	// Create valid JPEG
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	require.NoError(t, err)

	err = processor.ValidateImage(buf.Bytes())
	assert.NoError(t, err)
}

// TestImageProcessor_ValidateImage_GIF tests GIF validation.
func TestImageProcessor_ValidateImage_GIF(t *testing.T) {
	processor := NewImageProcessor()

	// Create valid GIF
	palette := color.Palette{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{255, 255, 255, 255},
	}
	img := image.NewPaletted(image.Rect(0, 0, 10, 10), palette)

	var buf bytes.Buffer
	err := gif.Encode(&buf, img, nil)
	require.NoError(t, err)

	err = processor.ValidateImage(buf.Bytes())
	assert.NoError(t, err)
}

// TestImageProcessor_ValidateImage_BMP tests BMP validation.
func TestImageProcessor_ValidateImage_BMP(t *testing.T) {
	processor := NewImageProcessor()

	// Create valid BMP
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer
	err := bmp.Encode(&buf, img)
	require.NoError(t, err)

	err = processor.ValidateImage(buf.Bytes())
	assert.NoError(t, err)
}

// TestImageProcessor_ValidateImage_EmptyData tests validation with empty data.
func TestImageProcessor_ValidateImage_EmptyData(t *testing.T) {
	processor := NewImageProcessor()

	err := processor.ValidateImage([]byte{})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInsufficientData)
}

// TestImageProcessor_ValidateImage_InvalidFormat tests non-image data.
func TestImageProcessor_ValidateImage_InvalidFormat(t *testing.T) {
	processor := NewImageProcessor()

	err := processor.ValidateImage([]byte("not an image"))
	assert.Error(t, err)
}

// TestImageProcessor_ValidateImage_CorruptedPNG tests corrupted PNG.
func TestImageProcessor_ValidateImage_CorruptedPNG(t *testing.T) {
	processor := NewImageProcessor()

	// Invalid PNG signature
	corruptedPNG := []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00, 0x00, 0x00}

	err := processor.ValidateImage(corruptedPNG)
	assert.Error(t, err)
}

// TestImageProcessor_ValidateImage_CorruptedJPEG tests corrupted JPEG.
func TestImageProcessor_ValidateImage_CorruptedJPEG(t *testing.T) {
	processor := NewImageProcessor()

	// Invalid JPEG (wrong SOI marker)
	corruptedJPEG := []byte{0xFF, 0x00, 0x00, 0x00}

	err := processor.ValidateImage(corruptedJPEG)
	assert.Error(t, err)
}

// TestImageProcessor_RoundTrip tests full encode/decode cycle.
func TestImageProcessor_RoundTrip(t *testing.T) {
	processor := NewImageProcessor()

	formats := []media.MediaFormat{
		media.FormatPNG,
		media.FormatJPEG,
		media.FormatBMP,
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			// Create original image
			original := image.NewRGBA(image.Rect(0, 0, 50, 50))
			for y := 0; y < 50; y++ {
				for x := 0; x < 50; x++ {
					original.Set(x, y, color.RGBA{
						uint8(x * 5), uint8(y * 5), 128, 255,
					})
				}
			}

			// Save
			data, err := processor.SaveImage(original, format)
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Load back
			loaded, loadedFormat, err := processor.LoadImage(data)
			require.NoError(t, err)
			assert.Equal(t, format, loadedFormat)
			assert.Equal(t, original.Bounds(), loaded.Bounds())

			// For lossless formats (PNG, BMP), verify pixel accuracy
			// JPEG is lossy, so we skip exact pixel comparison
			if format == media.FormatPNG || format == media.FormatBMP {
				// Check a sample pixel
				origColor := original.At(25, 25)
				loadedColor := loaded.At(25, 25)
				assert.Equal(t, origColor, loadedColor)
			}
		})
	}
}

// TestImageProcessor_CalculateQualityScore tests quality scoring.
func TestImageProcessor_CalculateQualityScore(t *testing.T) {
	processor := NewImageProcessor()

	tests := []struct {
		name         string
		width        int
		height       int
		format       media.MediaFormat
		colorSpace   media.ColorSpace
		expectHigher float64 // Minimum expected score
	}{
		{
			name:         "large PNG RGBA",
			width:        2000,
			height:       2000,
			format:       media.FormatPNG,
			colorSpace:   media.ColorSpaceRGBA,
			expectHigher: 0.8, // Large + PNG + RGBA = high score
		},
		{
			name:         "small JPEG RGB",
			width:        100,
			height:       100,
			format:       media.FormatJPEG,
			colorSpace:   media.ColorSpaceRGB,
			expectHigher: 0.0, // Small + JPEG = lower score
		},
		{
			name:         "medium BMP RGB",
			width:        1000,
			height:       800,
			format:       media.FormatBMP,
			colorSpace:   media.ColorSpaceRGB,
			expectHigher: 0.5, // Medium size + BMP
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ImageInfo{
				Width:      tt.width,
				Height:     tt.height,
				Format:     tt.format,
				ColorSpace: tt.colorSpace,
			}

			totalPixels := int64(tt.width * tt.height)
			score := processor.calculateQualityScore(info, totalPixels)

			assert.GreaterOrEqual(t, score, 0.0)
			assert.LessOrEqual(t, score, 1.0)
			assert.GreaterOrEqual(t, score, tt.expectHigher)
		})
	}
}

// Additional coverage tests for validation functions

func TestValidatePNG(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name: "Valid_PNG_Minimal",
			data: []byte{
				// PNG signature
				0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
				// IHDR chunk
				0x00, 0x00, 0x00, 0x0D,
				0x49, 0x48, 0x44, 0x52,
				0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
				0x08, 0x02, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				// IEND chunk
				0x00, 0x00, 0x00, 0x00,
				0x49, 0x45, 0x4E, 0x44,
				0xAE, 0x42, 0x60, 0x82,
			},
			expectError: false,
		},
		{
			name:        "TooShort_Error",
			data:        []byte{0x89, 0x50, 0x4E, 0x47},
			expectError: true,
		},
		{
			name: "InvalidSignature_Error",
			data: []byte{
				0x00, 0x00, 0x00, 0x00, 0x0D, 0x0A, 0x1A, 0x0A,
			},
			expectError: true,
		},
		{
			name: "MissingIHDR_Error",
			data: []byte{
				0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
				0x00, 0x00, 0x00, 0x0D,
				0x00, 0x00, 0x00, 0x00, // Wrong chunk type (not IHDR)
				0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
				0x08, 0x02, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewImageProcessor()
			err := processor.validatePNG(tt.data)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateJPEG(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name: "Valid_JPEG_Minimal",
			data: []byte{
				// SOI
				0xFF, 0xD8,
				// APP0 (minimal JFIF)
				0xFF, 0xE0, 0x00, 0x10,
				0x4A, 0x46, 0x49, 0x46, 0x00,
				0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00,
				// EOI
				0xFF, 0xD9,
			},
			expectError: false,
		},
		{
			name:        "TooShort_Error",
			data:        []byte{0xFF},
			expectError: true,
		},
		{
			name: "InvalidSOI_Error",
			data: []byte{
				0x00, 0x00, 0xFF, 0xD9,
			},
			expectError: true,
		},
		{
			name: "NoEOI_Acceptable",
			data: []byte{
				0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10,
				0x4A, 0x46, 0x49, 0x46, 0x00,
				0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00,
				// EOI not strictly required
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewImageProcessor()
			err := processor.validateJPEG(tt.data)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateGIF(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name: "Valid_GIF89a",
			data: []byte{
				// GIF89a signature
				0x47, 0x49, 0x46, 0x38, 0x39, 0x61,
				// Logical screen descriptor (7 bytes minimum)
				0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
				// Trailer
				0x3B,
			},
			expectError: false,
		},
		{
			name:        "TooShort_Error",
			data:        []byte{0x47, 0x49, 0x46},
			expectError: true,
		},
		{
			name: "InvalidSignature_Error",
			data: []byte{
				0x00, 0x00, 0x00, 0x38, 0x39, 0x61,
				0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
				0x3B,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewImageProcessor()
			err := processor.validateGIF(tt.data)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateBMP(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name: "Valid_BMP",
			data: []byte{
				// BM signature
				0x42, 0x4D,
				// File size
				0x36, 0x00, 0x00, 0x00,
				// Reserved
				0x00, 0x00, 0x00, 0x00,
				// Pixel data offset
				0x36, 0x00, 0x00, 0x00,
				// DIB header size
				0x28, 0x00, 0x00, 0x00,
			},
			expectError: false,
		},
		{
			name:        "TooShort_Error",
			data:        []byte{0x42, 0x4D},
			expectError: true,
		},
		{
			name: "InvalidSignature_Error",
			data: []byte{
				0x00, 0x00,
				0x36, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x36, 0x00, 0x00, 0x00,
				0x28, 0x00, 0x00, 0x00,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewImageProcessor()
			err := processor.validateBMP(tt.data)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
