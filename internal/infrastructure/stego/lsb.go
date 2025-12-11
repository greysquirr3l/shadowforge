// Package stego provides infrastructure implementations of steganographic techniques.
package stego

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// LSBTechnique implements Least Significant Bit steganography for images.
type LSBTechnique struct {
	bitsPerChannel int
	channels       string
}

// NewLSBTechnique creates a new LSB technique with default settings (1-bit LSB, RGB channels).
func NewLSBTechnique() *LSBTechnique {
	return &LSBTechnique{
		bitsPerChannel: 1,
		channels:       "RGB",
	}
}

// Name returns the technique identifier.
func (l *LSBTechnique) Name() string {
	return string(stego.LSB)
}

// SupportsFormat checks if this technique supports the given media format.
func (l *LSBTechnique) SupportsFormat(format media.MediaFormat) bool {
	return format == media.FormatPNG || format == media.FormatBMP
}

// Embed hides payload data within the carrier image using LSB substitution.
func (l *LSBTechnique) Embed(ctx context.Context, carrier, payload []byte) ([]byte, error) {
	// Decode carrier image
	img, _, err := image.Decode(bytes.NewReader(carrier))
	if err != nil {
		return nil, fmt.Errorf("failed to decode carrier image: %w", err)
	}

	rgba := imageToRGBA(img)
	bounds := rgba.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Calculate capacity and validate
	capacity := l.calculateCapacity(width, height)
	totalSize := len(payload) + 4 // +4 for length prefix

	if totalSize > capacity {
		return nil, fmt.Errorf("payload too large: need %d bytes, capacity %d bytes", totalSize, capacity)
	}

	// Prepare data: length (4 bytes) + payload
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(len(payload)))
	dataToEmbed := append(lengthBytes, payload...)

	// Embed data
	if err := l.embedData(rgba, dataToEmbed); err != nil {
		return nil, err
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		return nil, fmt.Errorf("failed to encode result: %w", err)
	}

	return buf.Bytes(), nil
}

// Extract retrieves hidden payload data from the carrier image.
func (l *LSBTechnique) Extract(ctx context.Context, carrier []byte) ([]byte, error) {
	// Decode carrier image
	img, _, err := image.Decode(bytes.NewReader(carrier))
	if err != nil {
		return nil, fmt.Errorf("failed to decode carrier: %w", err)
	}

	rgba := imageToRGBA(img)

	// Extract length (4 bytes)
	lengthBytes, err := l.extractBytes(rgba, 4, 0)
	if err != nil {
		return nil, err
	}

	payloadLength := int(binary.BigEndian.Uint32(lengthBytes))
	if payloadLength <= 0 || payloadLength > 1024*1024*10 { // Max 10MB
		return nil, fmt.Errorf("invalid payload length: %d", payloadLength)
	}

	// Extract payload
	payload, err := l.extractBytes(rgba, payloadLength, 4)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// CalculateCapacity determines the maximum payload size in bytes.
func (l *LSBTechnique) CalculateCapacity(ctx context.Context, carrier []byte) (int, error) {
	img, _, err := image.Decode(bytes.NewReader(carrier))
	if err != nil {
		return 0, fmt.Errorf("failed to decode carrier: %w", err)
	}

	bounds := img.Bounds()
	capacity := l.calculateCapacity(bounds.Dx(), bounds.Dy())
	return capacity - 4, nil // -4 for length prefix
}

// calculateCapacity calculates raw capacity in bytes.
func (l *LSBTechnique) calculateCapacity(width, height int) int {
	totalPixels := width * height
	channelCount := len(l.channels) // "RGB" = 3 channels
	bitsAvailable := totalPixels * channelCount * l.bitsPerChannel
	return bitsAvailable / 8
}

// embedData embeds data bytes into image LSBs.
func (l *LSBTechnique) embedData(rgba *image.RGBA, data []byte) error {
	bounds := rgba.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	bitIndex := 0
	totalBits := len(data) * 8

	for y := 0; y < height && bitIndex < totalBits; y++ {
		for x := 0; x < width && bitIndex < totalBits; x++ {
			pixel := rgba.RGBAAt(x, y)

			// Embed in RGB channels
			pixel.R = l.embedBitsInByte(pixel.R, data, &bitIndex, totalBits)
			pixel.G = l.embedBitsInByte(pixel.G, data, &bitIndex, totalBits)
			pixel.B = l.embedBitsInByte(pixel.B, data, &bitIndex, totalBits)

			rgba.SetRGBA(x, y, pixel)
		}
	}

	return nil
}

// embedBitsInByte embeds bits into a color channel byte.
func (l *LSBTechnique) embedBitsInByte(channelValue uint8, data []byte, bitIndex *int, totalBits int) uint8 {
	result := channelValue

	for i := 0; i < l.bitsPerChannel && *bitIndex < totalBits; i++ {
		byteIndex := *bitIndex / 8
		bitOffset := 7 - (*bitIndex % 8)
		dataBit := (data[byteIndex] >> bitOffset) & 1

		// Clear and set LSB at position i
		mask := ^(uint8(1) << i)
		result = (result & mask) | (dataBit << i)

		*bitIndex++
	}

	return result
}

// extractBytes extracts numBytes starting at byteOffset.
func (l *LSBTechnique) extractBytes(rgba *image.RGBA, numBytes int, byteOffset int) ([]byte, error) {
	bounds := rgba.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	result := make([]byte, numBytes)
	imageBitIndex := 0 // Current position in image bits
	startBit := byteOffset * 8
	endBit := (byteOffset + numBytes) * 8

	for y := 0; y < height && imageBitIndex < endBit; y++ {
		for x := 0; x < width && imageBitIndex < endBit; x++ {
			pixel := rgba.RGBAAt(x, y)

			// Extract from RGB channels
			l.extractBitsFromByte(pixel.R, result, &imageBitIndex, startBit, endBit)
			l.extractBitsFromByte(pixel.G, result, &imageBitIndex, startBit, endBit)
			l.extractBitsFromByte(pixel.B, result, &imageBitIndex, startBit, endBit)
		}
	}

	return result, nil
}

// extractBitsFromByte extracts bits from a color channel byte.
// imageBitIndex tracks absolute position in the image
// startBit is where we start extracting (to skip header)
// We only extract bits when imageBitIndex >= startBit
func (l *LSBTechnique) extractBitsFromByte(channelValue uint8, result []byte, imageBitIndex *int, startBit, endBit int) {
	for i := 0; i < l.bitsPerChannel && *imageBitIndex < endBit; i++ {
		if *imageBitIndex >= startBit {
			dataBit := (channelValue >> i) & 1

			// Map image bit position to result array position
			resultByteIndex := (*imageBitIndex - startBit) / 8
			resultBitOffset := 7 - ((*imageBitIndex - startBit) % 8)

			if dataBit == 1 {
				result[resultByteIndex] |= (1 << resultBitOffset)
			}
		}

		*imageBitIndex++
	}
}

// imageToRGBA converts any image to RGBA format.
func imageToRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	return rgba
}
