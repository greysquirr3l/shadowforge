package stego

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLSBTechnique(t *testing.T) {
	lsb := NewLSBTechnique()

	assert.NotNil(t, lsb)
	assert.Equal(t, 1, lsb.bitsPerChannel)
	assert.Equal(t, "RGB", lsb.channels)
}

func TestLSBTechnique_Name(t *testing.T) {
	lsb := NewLSBTechnique()

	assert.Equal(t, string(stego.LSB), lsb.Name())
	assert.Equal(t, "lsb", lsb.Name())
}

func TestLSBTechnique_SupportsFormat(t *testing.T) {
	tests := []struct {
		name           string
		format         media.MediaFormat
		expectedResult bool
	}{
		{
			name:           "Supports_PNG",
			format:         media.FormatPNG,
			expectedResult: true,
		},
		{
			name:           "Supports_BMP",
			format:         media.FormatBMP,
			expectedResult: true,
		},
		{
			name:           "NotSupports_JPEG",
			format:         media.FormatJPEG,
			expectedResult: false,
		},
		{
			name:           "NotSupports_GIF",
			format:         media.FormatGIF,
			expectedResult: false,
		},
		{
			name:           "NotSupports_WAV",
			format:         media.FormatWAV,
			expectedResult: false,
		},
		{
			name:           "NotSupports_MP3",
			format:         media.FormatMP3,
			expectedResult: false,
		},
		{
			name:           "NotSupports_TXT",
			format:         media.FormatTXT,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lsb := NewLSBTechnique()
			result := lsb.SupportsFormat(tt.format)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestLSBTechnique_CalculateCapacity(t *testing.T) {
	tests := []struct {
		name             string
		width            int
		height           int
		expectedCapacity int // In bytes, minus 4-byte header
	}{
		{
			name:             "Small_10x10",
			width:            10,
			height:           10,
			expectedCapacity: (10 * 10 * 3 * 1 / 8) - 4, // 37 - 4 = 33 bytes
		},
		{
			name:             "Medium_100x100",
			width:            100,
			height:           100,
			expectedCapacity: (100 * 100 * 3 * 1 / 8) - 4, // 3750 - 4 = 3746 bytes
		},
		{
			name:             "Large_1920x1080",
			width:            1920,
			height:           1080,
			expectedCapacity: (1920 * 1080 * 3 * 1 / 8) - 4, // 777600 - 4 = 777596 bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test image
			carrier := createTestImage(tt.width, tt.height)
			lsb := NewLSBTechnique()

			capacity, err := lsb.CalculateCapacity(context.Background(), carrier)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCapacity, capacity)
		})
	}
}

func TestLSBTechnique_CalculateCapacity_InvalidImage(t *testing.T) {
	lsb := NewLSBTechnique()

	_, err := lsb.CalculateCapacity(context.Background(), []byte{0x00, 0x01, 0x02})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode carrier")
}

func TestLSBTechnique_EmbedAndExtract_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		height  int
		payload []byte
	}{
		{
			name:    "Small_Payload_10_Bytes",
			width:   50,
			height:  50,
			payload: []byte("Hello LSB!"),
		},
		{
			name:    "Medium_Payload_100_Bytes",
			width:   200,
			height:  200,
			payload: bytes.Repeat([]byte("X"), 100),
		},
		{
			name:    "Large_Payload_1KB",
			width:   500,
			height:  500,
			payload: bytes.Repeat([]byte("A"), 1024),
		},
		{
			name:    "Binary_Data",
			width:   100,
			height:  100,
			payload: []byte{0x00, 0xFF, 0xAA, 0x55, 0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create carrier image
			carrier := createTestImage(tt.width, tt.height)
			lsb := NewLSBTechnique()

			// Embed payload
			stego, err := lsb.Embed(context.Background(), carrier, tt.payload)
			require.NoError(t, err)
			require.NotNil(t, stego)

			// Extract payload
			extracted, err := lsb.Extract(context.Background(), stego)
			require.NoError(t, err)

			// Verify payload matches
			assert.Equal(t, tt.payload, extracted)
		})
	}
}

func TestLSBTechnique_Embed_EmptyPayload(t *testing.T) {
	carrier := createTestImage(100, 100)
	lsb := NewLSBTechnique()

	// Embed empty payload
	stego, err := lsb.Embed(context.Background(), carrier, []byte{})
	require.NoError(t, err)

	// Extract should fail with invalid length
	_, err = lsb.Extract(context.Background(), stego)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid payload length")
}

func TestLSBTechnique_Embed_PayloadTooLarge(t *testing.T) {
	// Create small carrier (10x10 = ~37 bytes capacity)
	carrier := createTestImage(10, 10)
	// Create payload larger than capacity
	payload := bytes.Repeat([]byte("X"), 100)
	lsb := NewLSBTechnique()

	_, err := lsb.Embed(context.Background(), carrier, payload)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "payload too large")
}

func TestLSBTechnique_Embed_InvalidCarrier(t *testing.T) {
	lsb := NewLSBTechnique()

	_, err := lsb.Embed(context.Background(), []byte{0x00, 0x01}, []byte("test"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode carrier image")
}

func TestLSBTechnique_Extract_InvalidCarrier(t *testing.T) {
	lsb := NewLSBTechnique()

	_, err := lsb.Extract(context.Background(), []byte{0x00, 0x01})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode carrier")
}

func TestLSBTechnique_Extract_InvalidPayloadLength(t *testing.T) {
	// Create carrier with invalid length encoding (too large)
	carrier := createTestImage(50, 50)
	lsb := NewLSBTechnique()

	// Manually create a carrier with invalid length header (>10MB)
	img, _, _ := image.Decode(bytes.NewReader(carrier))
	rgba := imageToRGBA(img)
	// Embed invalid length: 20MB (0x01400000)
	invalidLength := []byte{0x01, 0x40, 0x00, 0x00}
	_ = lsb.embedData(rgba, invalidLength)
	var buf bytes.Buffer
	_ = png.Encode(&buf, rgba)
	invalidCarrier := buf.Bytes()

	// Try to extract
	_, err := lsb.Extract(context.Background(), invalidCarrier)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid payload length")
}

func TestLSBTechnique_Extract_ZeroPayloadLength(t *testing.T) {
	// Create carrier with zero length encoding
	carrier := createTestImage(50, 50)
	lsb := NewLSBTechnique()

	// Manually create a carrier with zero length header
	img, _, _ := image.Decode(bytes.NewReader(carrier))
	rgba := imageToRGBA(img)
	// Embed zero length
	zeroLength := []byte{0x00, 0x00, 0x00, 0x00}
	_ = lsb.embedData(rgba, zeroLength)
	var buf bytes.Buffer
	_ = png.Encode(&buf, rgba)
	invalidCarrier := buf.Bytes()

	// Try to extract
	_, err := lsb.Extract(context.Background(), invalidCarrier)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid payload length")
}

func TestLSBTechnique_ImageToRGBA_AlreadyRGBA(t *testing.T) {
	// Create RGBA image
	rgba := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// Convert should return the same image
	result := imageToRGBA(rgba)

	assert.Equal(t, rgba, result)
}

func TestLSBTechnique_ImageToRGBA_NonRGBA(t *testing.T) {
	// Create non-RGBA image (Gray)
	gray := image.NewGray(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			gray.SetGray(x, y, color.Gray{Y: 128})
		}
	}

	// Convert to RGBA
	rgba := imageToRGBA(gray)

	assert.NotNil(t, rgba)
	assert.Equal(t, 10, rgba.Bounds().Dx())
	assert.Equal(t, 10, rgba.Bounds().Dy())

	// Verify pixel conversion worked
	r, g, b, a := rgba.At(5, 5).RGBA()
	assert.NotZero(t, r)
	assert.NotZero(t, g)
	assert.NotZero(t, b)
	assert.NotZero(t, a)
}

func TestLSBTechnique_EmbedExtractBits_Correctness(t *testing.T) {
	// Test embedding and extracting individual bits
	lsb := NewLSBTechnique()
	data := []byte{0xAB, 0xCD} // 10101011 11001101

	// Embed bits into channel values
	bitIndex := 0
	totalBits := len(data) * 8
	channelValue := uint8(0xFF)

	// Verify bit was embedded (LSB should be 1 from first bit of 0xAB)
	result := lsb.embedBitsInByte(channelValue, data, &bitIndex, totalBits)
	assert.Equal(t, uint8(1), result&1)

	// Extract bit back
	extractedData := make([]byte, 2)
	extractBitIndex := 0
	lsb.extractBitsFromByte(result, extractedData, &extractBitIndex, 0, 16)

	// First bit should match
	firstBit := (extractedData[0] >> 7) & 1
	expectedFirstBit := (data[0] >> 7) & 1
	assert.Equal(t, expectedFirstBit, firstBit)
}

func TestLSBTechnique_MultiByteEmbedding(t *testing.T) {
	// Test that multi-byte payloads are embedded correctly across pixels
	carrier := createTestImage(100, 100)
	payload := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}
	lsb := NewLSBTechnique()

	stego, err := lsb.Embed(context.Background(), carrier, payload)
	require.NoError(t, err)

	extracted, err := lsb.Extract(context.Background(), stego)
	require.NoError(t, err)

	// Verify each byte matches
	assert.Equal(t, payload, extracted)
	for i := range payload {
		assert.Equal(t, payload[i], extracted[i], "Byte at index %d should match", i)
	}
}

func TestLSBTechnique_ContextCancellation(t *testing.T) {
	carrier := createTestImage(100, 100)
	payload := []byte("test payload")
	lsb := NewLSBTechnique()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Embed should still work (we don't currently check ctx, but test for future-proofing)
	_, err := lsb.Embed(ctx, carrier, payload)
	// Current implementation doesn't check context, so this should succeed
	// In future, we might want to add context checking
	assert.NoError(t, err)
}

// Helper function to create test PNG images
func createTestImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with gradient pattern for better testing
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(128)
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}
