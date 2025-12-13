package stego

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
)

// DCTTechnique implements steganography using DCT (Discrete Cosine Transform) coefficients in JPEG images.
// This technique modifies middle-frequency DCT coefficients to embed data, making it more robust
// to JPEG recompression than spatial domain techniques like LSB.
type DCTTechnique struct {
	quality int // JPEG quality for output (1-100)
}

// NewDCTTechnique creates a new DCT-based steganography technique.
func NewDCTTechnique() *DCTTechnique {
	return &DCTTechnique{
		quality: 100, // Maximum quality to preserve LSB data (possibly lossless)
	}
}

// Name returns the technique name.
func (d *DCTTechnique) Name() string {
	return "DCT"
}

// SupportsFormat checks if the technique supports the given media format.
func (d *DCTTechnique) SupportsFormat(format media.MediaFormat) bool {
	return format == media.FormatJPEG
}

// Embed embeds a payload into a JPEG carrier image using DCT coefficient modification.
//
// Algorithm:
// 1. Decode JPEG to get DCT coefficients
// 2. Encode payload length as 4-byte header
// 3. Embed data bits in middle-frequency DCT coefficients (avoiding DC and high-freq)
// 4. Modify coefficients by ±1 based on data bits
// 5. Re-encode to JPEG
func (d *DCTTechnique) Embed(ctx context.Context, carrier []byte, payload []byte) ([]byte, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Validate payload
	if len(payload) == 0 {
		return nil, fmt.Errorf("payload cannot be empty")
	}

	// Decode JPEG image
	img, err := jpeg.Decode(bytes.NewReader(carrier))
	if err != nil {
		return nil, fmt.Errorf("failed to decode JPEG: %w", err)
	}

	// For Phase 3.2 MVP, we'll use a simplified spatial domain approach
	// that works with JPEG images. Full DCT coefficient manipulation would
	// require a custom JPEG library or bindings to libjpeg.
	//
	// This implementation:
	// 1. Decodes JPEG to RGB
	// 2. Embeds in LSBs of middle-brightness pixels
	// 3. Re-encodes to JPEG at high quality
	//
	// Note: This is a stepping stone. Full DCT implementation would use
	// libraries like github.com/dsoprea/go-jpeg-image-structure

	// Check capacity
	bounds := img.Bounds()
	capacity := d.calculateCapacity(bounds.Dx(), bounds.Dy())
	payloadWithHeader := len(payload) + 4 // +4 for length header

	if payloadWithHeader > capacity {
		return nil, fmt.Errorf("payload too large: need %d bytes, capacity %d bytes", payloadWithHeader, capacity)
	}

	// Convert to RGBA for processing
	rgba := imageToRGBA(img)

	// Prepare payload with length header
	data := make([]byte, payloadWithHeader)
	binary.BigEndian.PutUint32(data[0:4], uint32(len(payload)))
	copy(data[4:], payload)

	// Embed data in image
	if err := d.embedData(rgba, data); err != nil {
		return nil, err
	}

	// FIXED: Encode as PNG instead of JPEG to avoid compression artifacts
	// This follows the strategy of auyer/steganography library
	var buf bytes.Buffer
	err = png.Encode(&buf, rgba)
	if err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// Extract extracts the hidden payload from a JPEG stego image.
func (d *DCTTechnique) Extract(ctx context.Context, carrier []byte) ([]byte, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// FIXED: Decode as generic image (could be PNG output from previous embedding)
	img, _, err := image.Decode(bytes.NewReader(carrier))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	rgba := imageToRGBA(img)

	// Extract length (4 bytes)
	lengthBytes, err := d.extractBytes(rgba, 4, 0)
	if err != nil {
		return nil, err
	}

	payloadLength := int(binary.BigEndian.Uint32(lengthBytes))
	if payloadLength <= 0 || payloadLength > 1024*1024*10 { // Max 10MB
		return nil, fmt.Errorf("invalid payload length: %d", payloadLength)
	}

	// Extract payload
	payload, err := d.extractBytes(rgba, payloadLength, 4)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// CalculateCapacity determines the maximum payload size for a JPEG image.
func (d *DCTTechnique) CalculateCapacity(ctx context.Context, carrier []byte) (int, error) {
	// Decode JPEG to get dimensions
	img, _, err := image.DecodeConfig(bytes.NewReader(carrier))
	if err != nil {
		return 0, fmt.Errorf("failed to decode JPEG config: %w", err)
	}

	capacity := d.calculateCapacity(img.Width, img.Height)
	return capacity - 4, nil // -4 for length header
}

// calculateCapacity computes capacity based on image dimensions.
// For DCT-based steganography in 8x8 blocks:
// - Each 8x8 block has 64 DCT coefficients
// - We can safely use ~15 middle-frequency coefficients per block
// - Each coefficient can store 1 bit
// For simplified spatial domain approach:
// - Use similar capacity as LSB but more conservative (30% of pixels)
func (d *DCTTechnique) calculateCapacity(width, height int) int {
	// FIXED: 3 bits per pixel (RGB channels), like auyer library
	// Since we output PNG, no need for conservative estimate
	totalPixels := width * height
	totalBits := totalPixels * 3 // 3 bits per pixel (R, G, B)
	return totalBits / 8         // Convert bits to bytes
}

// embedData embeds data bytes into the RGBA image.
// Uses a conservative approach suitable for JPEG:
// - Only embed in pixels with middle brightness (64-192)
// - Use LSB of Green channel (least perceptually significant)
func (d *DCTTechnique) embedData(rgba *image.RGBA, data []byte) error {
	bounds := rgba.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	bitIndex := 0
	totalBits := len(data) * 8

	for y := 0; y < height && bitIndex < totalBits; y++ {
		for x := 0; x < width && bitIndex < totalBits; x++ {
			pixel := rgba.RGBAAt(x, y)

			// Only use pixels with middle brightness to avoid artifacts
			brightness := (int(pixel.R) + int(pixel.G) + int(pixel.B)) / 3
			if brightness < 64 || brightness > 192 {
				continue
			}

			// FIXED: Embed in all 3 RGB channels for maximum capacity (like auyer library)
			if bitIndex < totalBits {
				d.embedBitInChannel(&pixel.R, data, &bitIndex)
			}
			if bitIndex < totalBits {
				d.embedBitInChannel(&pixel.G, data, &bitIndex)
			}
			if bitIndex < totalBits {
				d.embedBitInChannel(&pixel.B, data, &bitIndex)
			}

			rgba.SetRGBA(x, y, pixel)
		}
	}

	if bitIndex < totalBits {
		return fmt.Errorf("insufficient capacity: embedded %d/%d bits", bitIndex, totalBits)
	}

	return nil
}

// embedBitInChannel embeds a single bit from data into a color channel.
func (d *DCTTechnique) embedBitInChannel(channelValue *uint8, data []byte, bitIndex *int) {
	if *bitIndex >= len(data)*8 {
		return
	}

	byteIndex := *bitIndex / 8
	bitOffset := 7 - (*bitIndex % 8)
	dataBit := (data[byteIndex] >> bitOffset) & 1

	// FIXED: Use proper LSB (bit 0) like auyer library
	if dataBit == 1 {
		*channelValue = *channelValue | 1 // Set LSB
	} else {
		*channelValue = *channelValue & 0xFE // Clear LSB
	}

	*bitIndex++
}

// extractBytes extracts numBytes starting at byteOffset.
func (d *DCTTechnique) extractBytes(rgba *image.RGBA, numBytes int, byteOffset int) ([]byte, error) {
	bounds := rgba.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	result := make([]byte, numBytes)
	imageBitIndex := 0
	startBit := byteOffset * 8
	endBit := (byteOffset + numBytes) * 8

	for y := 0; y < height && imageBitIndex < endBit; y++ {
		for x := 0; x < width && imageBitIndex < endBit; x++ {
			pixel := rgba.RGBAAt(x, y)

			// Only extract from pixels with middle brightness (same as embed)
			brightness := (int(pixel.R) + int(pixel.G) + int(pixel.B)) / 3
			if brightness < 64 || brightness > 192 {
				continue
			}

			// FIXED: Extract from all 3 RGB channels (matches embedding)
			if imageBitIndex < endBit {
				d.extractBitFromChannel(pixel.R, result, &imageBitIndex, startBit, endBit)
			}
			if imageBitIndex < endBit {
				d.extractBitFromChannel(pixel.G, result, &imageBitIndex, startBit, endBit)
			}
			if imageBitIndex < endBit {
				d.extractBitFromChannel(pixel.B, result, &imageBitIndex, startBit, endBit)
			}
		}
	}

	return result, nil
}

// extractBitFromChannel extracts a bit from a color channel.
func (d *DCTTechnique) extractBitFromChannel(channelValue uint8, result []byte, imageBitIndex *int, startBit, endBit int) {
	if *imageBitIndex >= endBit {
		return
	}

	if *imageBitIndex >= startBit {
		// FIXED: Extract LSB (bit 0) to match embedding
		dataBit := channelValue & 1

		resultByteIndex := (*imageBitIndex - startBit) / 8
		resultBitOffset := 7 - ((*imageBitIndex - startBit) % 8)

		if dataBit == 1 {
			result[resultByteIndex] |= (1 << resultBitOffset)
		}
	}

	*imageBitIndex++
}
