package media

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"golang.org/x/image/bmp"
)

// ImageProcessor handles image file operations.
type ImageProcessor struct {
	detector *FormatDetector
}

// NewImageProcessor creates a new image processor.
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		detector: NewFormatDetector(),
	}
}

// LoadImage reads and decodes an image file.
func (p *ImageProcessor) LoadImage(data []byte) (image.Image, media.MediaFormat, error) {
	if len(data) == 0 {
		return nil, "", ErrInsufficientData
	}

	// Detect format
	format, mediaType, err := p.detector.DetectFormat(data)
	if err != nil {
		return nil, "", fmt.Errorf("format detection failed: %w", err)
	}

	if mediaType != media.MediaTypeImage {
		return nil, "", ErrUnsupportedFormat
	}

	// Decode based on format
	var img image.Image
	reader := bytes.NewReader(data)

	switch format {
	case media.FormatPNG:
		img, err = png.Decode(reader)
	case media.FormatJPEG:
		img, err = jpeg.Decode(reader)
	case media.FormatGIF:
		img, err = gif.Decode(reader)
	case media.FormatBMP:
		img, err = bmp.Decode(reader)
	default:
		return nil, "", ErrUnsupportedFormat
	}

	if err != nil {
		return nil, "", fmt.Errorf("image decode failed: %w", err)
	}

	return img, format, nil
}

// SaveImage encodes and writes an image.
func (p *ImageProcessor) SaveImage(img image.Image, format media.MediaFormat) ([]byte, error) {
	if img == nil {
		return nil, fmt.Errorf("image is nil")
	}

	var buf bytes.Buffer

	switch format {
	case media.FormatPNG:
		err := png.Encode(&buf, img)
		if err != nil {
			return nil, fmt.Errorf("PNG encode failed: %w", err)
		}

	case media.FormatJPEG:
		opts := &jpeg.Options{Quality: 95}
		err := jpeg.Encode(&buf, img, opts)
		if err != nil {
			return nil, fmt.Errorf("JPEG encode failed: %w", err)
		}

	case media.FormatGIF:
		opts := &gif.Options{NumColors: 256}
		err := gif.Encode(&buf, img, opts)
		if err != nil {
			return nil, fmt.Errorf("GIF encode failed: %w", err)
		}

	case media.FormatBMP:
		err := bmp.Encode(&buf, img)
		if err != nil {
			return nil, fmt.Errorf("BMP encode failed: %w", err)
		}

	default:
		return nil, ErrUnsupportedFormat
	}

	return buf.Bytes(), nil
}

// SanitizeMetadata strips all metadata from an image, returning a clean copy.
// This prevents forensic analysis via EXIF, IPTC, XMP, or other metadata.
func (p *ImageProcessor) SanitizeMetadata(data []byte) ([]byte, error) {
	// Load image (this strips metadata during decode)
	img, format, err := p.LoadImage(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Re-encode without metadata
	sanitized, err := p.SaveImage(img, format)
	if err != nil {
		return nil, fmt.Errorf("failed to save sanitized image: %w", err)
	}

	return sanitized, nil
}

// CalculateCapacity estimates the steganographic capacity for an image.
// Returns raw capacity (theoretical maximum) and safe capacity (recommended).
func (p *ImageProcessor) CalculateCapacity(data []byte, technique string) (*media.CapacityInfo, error) {
	// Get image info
	info, err := p.detector.ParseImageInfo(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image info: %w", err)
	}

	// Load image to verify it's valid
	img, _, err := p.LoadImage(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	bounds := img.Bounds()
	totalPixels := int64(bounds.Dx() * bounds.Dy())

	var rawCapacity, safeCapacity int64

	switch technique {
	case "lsb":
		// LSB: 1 bit per color channel per pixel
		// For RGB: 3 bits per pixel, for RGBA: 4 bits per pixel
		bitsPerPixel := int64(3) // Default RGB
		if info.ColorSpace == media.ColorSpaceRGBA {
			bitsPerPixel = 4
		} else if info.ColorSpace == media.ColorSpaceGray {
			bitsPerPixel = 1
		}

		rawCapacity = (totalPixels * bitsPerPixel) / 8 // Convert to bytes

		// Safe capacity: Use only 1 bit per pixel (blue channel for RGB/RGBA)
		// and avoid first/last 10% of image to reduce detectability
		safePixels := int64(float64(totalPixels) * 0.80)
		safeCapacity = safePixels / 8

	case "lsb-2":
		// 2-bit LSB: 2 bits per channel
		bitsPerPixel := int64(6) // RGB
		if info.ColorSpace == media.ColorSpaceRGBA {
			bitsPerPixel = 8
		}
		rawCapacity = (totalPixels * bitsPerPixel) / 8
		safeCapacity = int64(float64(rawCapacity) * 0.70)

	case "dct":
		// DCT (JPEG only): Embed in middle-frequency DCT coefficients
		if info.Format != media.FormatJPEG {
			return nil, fmt.Errorf("DCT technique only supported for JPEG")
		}

		// Estimate: 1 bit per 8x8 block
		blocks := totalPixels / 64
		rawCapacity = blocks / 8

		// Safe capacity: Use fewer blocks to maintain quality
		safeCapacity = int64(float64(rawCapacity) * 0.50)

	case "palette":
		// Palette-based (GIF/PNG with palette)
		if info.Format != media.FormatGIF {
			return nil, fmt.Errorf("palette technique primarily for GIF")
		}

		// Estimate: Bits encoded in palette order/indices
		// Rough estimate: 1 byte per 256 pixels
		rawCapacity = totalPixels / 256
		safeCapacity = int64(float64(rawCapacity) * 0.60)

	default:
		return nil, fmt.Errorf("unknown technique: %s", technique)
	}

	// Calculate quality score based on image characteristics
	qualityScore := p.calculateQualityScore(info, totalPixels)

	// Note: We create a simplified CapacityInfo here without AssetID
	// The caller should set AssetID if needed
	return &media.CapacityInfo{
		AssetID:        media.AssetID{}, // Empty AssetID, to be set by caller
		MediaType:      media.MediaTypeImage,
		TotalCapacity:  rawCapacity,
		UsableCapacity: safeCapacity,
		RecommendedMax: safeCapacity,
		QualityImpact:  1.0 - qualityScore, // Impact is inverse of quality
	}, nil
}

// calculateQualityScore assesses image quality for steganography (0.0 - 1.0).
// Higher scores indicate better suitability for steganographic embedding.
func (p *ImageProcessor) calculateQualityScore(info *ImageInfo, totalPixels int64) float64 {
	score := 0.5 // Base score

	// Larger images are better (more capacity, harder to analyze)
	if totalPixels > 1920*1080 {
		score += 0.2
	} else if totalPixels > 800*600 {
		score += 0.1
	} else if totalPixels < 320*240 {
		score -= 0.2
	}

	// RGBA is better than RGB (more embedding channels)
	if info.ColorSpace == media.ColorSpaceRGBA {
		score += 0.1
	}

	// PNG is best (lossless), BMP second, JPEG/GIF less ideal
	switch info.Format {
	case media.FormatPNG:
		score += 0.2
	case media.FormatBMP:
		score += 0.1
	case media.FormatJPEG:
		score -= 0.1 // Lossy compression
	case media.FormatGIF:
		score -= 0.05 // Limited colors
	}

	// Clamp to [0.0, 1.0]
	if score < 0.0 {
		score = 0.0
	}
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// GetImageInfo extracts image metadata without fully decoding.
func (p *ImageProcessor) GetImageInfo(data []byte) (*ImageInfo, error) {
	return p.detector.ParseImageInfo(data)
}

// CreateBlankImage creates a new blank image of specified dimensions.
func (p *ImageProcessor) CreateBlankImage(width, height int, format media.MediaFormat) (image.Image, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: %dx%d", width, height)
	}

	switch format {
	case media.FormatPNG, media.FormatBMP:
		// RGBA for maximum flexibility
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		// Fill with white
		white := color.RGBA{255, 255, 255, 255}
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				img.Set(x, y, white)
			}
		}
		return img, nil

	case media.FormatJPEG:
		// RGB for JPEG (no alpha)
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		white := color.RGBA{255, 255, 255, 255}
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				img.Set(x, y, white)
			}
		}
		return img, nil

	case media.FormatGIF:
		// Paletted image for GIF
		palette := make(color.Palette, 256)
		// Create grayscale palette
		for i := 0; i < 256; i++ {
			c := uint8(i)
			palette[i] = color.RGBA{c, c, c, 255}
		}
		img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
		// Fill with white (index 255)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				img.SetColorIndex(x, y, 255)
			}
		}
		return img, nil

	default:
		return nil, ErrUnsupportedFormat
	}
}

// ValidateImage checks if image data is valid and can be processed.
func (p *ImageProcessor) ValidateImage(data []byte) error {
	if len(data) == 0 {
		return ErrInsufficientData
	}

	// Verify format is supported
	format, mediaType, err := p.detector.DetectFormat(data)
	if err != nil {
		return fmt.Errorf("format detection failed: %w", err)
	}

	if mediaType != media.MediaTypeImage {
		return fmt.Errorf("not an image file")
	}

	// Verify it can be decoded
	_, _, err = p.LoadImage(data)
	if err != nil {
		return fmt.Errorf("image validation failed: %w", err)
	}

	// Additional checks based on format
	switch format {
	case media.FormatPNG:
		return p.validatePNG(data)
	case media.FormatJPEG:
		return p.validateJPEG(data)
	case media.FormatGIF:
		return p.validateGIF(data)
	case media.FormatBMP:
		return p.validateBMP(data)
	}

	return nil
}

// validatePNG performs PNG-specific validation.
func (p *ImageProcessor) validatePNG(data []byte) error {
	// Check PNG signature
	if len(data) < 8 {
		return ErrInsufficientData
	}

	expectedSig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if !bytes.Equal(data[:8], expectedSig) {
		return ErrCorruptedHeader
	}

	// Verify IHDR chunk exists
	if len(data) < 33 {
		return ErrInsufficientData
	}

	chunkType := string(data[12:16])
	if chunkType != "IHDR" {
		return fmt.Errorf("missing IHDR chunk")
	}

	return nil
}

// validateJPEG performs JPEG-specific validation.
func (p *ImageProcessor) validateJPEG(data []byte) error {
	if len(data) < 2 {
		return ErrInsufficientData
	}

	// Check SOI marker (0xFF 0xD8)
	if data[0] != 0xFF || data[1] != 0xD8 {
		return ErrCorruptedHeader
	}

	// Check for EOI marker at end (0xFF 0xD9)
	if len(data) >= 2 {
		endIdx := len(data) - 2
		if data[endIdx] != 0xFF || data[endIdx+1] != 0xD9 {
			// EOI not required to be at absolute end, just should exist
			// This is a warning, not an error
		}
	}

	return nil
}

// validateGIF performs GIF-specific validation.
func (p *ImageProcessor) validateGIF(data []byte) error {
	if len(data) < 6 {
		return ErrInsufficientData
	}

	// Check GIF signature
	sig := string(data[:6])
	if sig != "GIF87a" && sig != "GIF89a" {
		return ErrCorruptedHeader
	}

	return nil
}

// validateBMP performs BMP-specific validation.
func (p *ImageProcessor) validateBMP(data []byte) error {
	if len(data) < 14 {
		return ErrInsufficientData
	}

	// Check BM signature
	if data[0] != 'B' || data[1] != 'M' {
		return ErrCorruptedHeader
	}

	// Check file size matches
	fileSize := binary.LittleEndian.Uint32(data[2:6])
	if uint32(len(data)) != fileSize {
		// Size mismatch - warning but not fatal
	}

	return nil
}
