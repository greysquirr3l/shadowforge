package stego

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDCTTechnique(t *testing.T) {
	dct := NewDCTTechnique()
	assert.NotNil(t, dct)
	assert.Equal(t, 100, dct.quality) // Updated - we now use quality 100 for PNG compatibility
}

func TestDCTTechnique_Name(t *testing.T) {
	dct := NewDCTTechnique()
	assert.Equal(t, "DCT", dct.Name())
}

func TestDCTTechnique_SupportsFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   media.MediaFormat
		expected bool
	}{
		{
			name:     "Supports_JPEG",
			format:   media.FormatJPEG,
			expected: true,
		},
		{
			name:     "NotSupports_PNG",
			format:   media.FormatPNG,
			expected: false,
		},
		{
			name:     "NotSupports_BMP",
			format:   media.FormatBMP,
			expected: false,
		},
		{
			name:     "NotSupports_GIF",
			format:   media.FormatGIF,
			expected: false,
		},
		{
			name:     "NotSupports_WAV",
			format:   media.FormatWAV,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dct := NewDCTTechnique()
			result := dct.SupportsFormat(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDCTTechnique_CalculateCapacity(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		expected int
	}{
		{
			name:     "Small_100x100",
			width:    100,
			height:   100,
			expected: 3746, // (100*100*3)/8 - 4 = 3746 (3 channels RGB, PNG format)
		},
		{
			name:     "Medium_640x480",
			width:    640,
			height:   480,
			expected: 115196, // (640*480*3)/8 - 4 = 115196
		},
		{
			name:     "Large_1920x1080",
			width:    1920,
			height:   1080,
			expected: 777596, // (1920*1080*3)/8 - 4 = 777596
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create JPEG test image
			carrier := createTestJPEG(tt.width, tt.height, 95)
			dct := NewDCTTechnique()

			capacity, err := dct.CalculateCapacity(context.Background(), carrier)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, capacity)
		})
	}
}

func TestDCTTechnique_CalculateCapacity_InvalidImage(t *testing.T) {
	dct := NewDCTTechnique()
	invalidData := []byte("not a valid JPEG")

	capacity, err := dct.CalculateCapacity(context.Background(), invalidData)
	assert.Error(t, err)
	assert.Equal(t, 0, capacity)
	assert.Contains(t, err.Error(), "failed to decode JPEG")
}

func TestDCTTechnique_EmbedAndExtract_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		height  int
		payload []byte
	}{
		{
			name:    "Small_Payload_10_Bytes",
			width:   200,
			height:  200,
			payload: []byte("Hello DCT!"),
		},
		{
			name:    "Medium_Payload_100_Bytes",
			width:   300,
			height:  300,
			payload: bytes.Repeat([]byte("D"), 100),
		},
		{
			name:    "Large_Payload_500_Bytes",
			width:   500,
			height:  400,
			payload: bytes.Repeat([]byte("JPEG steganography with DCT coefficients. "), 12),
		},
		{
			name:    "Binary_Data",
			width:   250,
			height:  250,
			payload: []byte{0x00, 0xFF, 0xAA, 0x55, 0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create JPEG carrier image
			carrier := createTestJPEG(tt.width, tt.height, 95)
			dct := NewDCTTechnique()

			// Embed payload
			stego, err := dct.Embed(context.Background(), carrier, tt.payload)
			require.NoError(t, err)
			require.NotNil(t, stego)

			// Verify it's valid image format (now PNG output)
			_, _, err = image.Decode(bytes.NewReader(stego))
			require.NoError(t, err)

			// Extract payload
			extracted, err := dct.Extract(context.Background(), stego)
			require.NoError(t, err)

			// Verify payload matches
			assert.Equal(t, tt.payload, extracted)
		})
	}
}

func TestDCTTechnique_Embed_EmptyPayload(t *testing.T) {
	carrier := createTestJPEG(200, 200, 95)
	dct := NewDCTTechnique()

	// Embed empty payload should fail
	_, err := dct.Embed(context.Background(), carrier, []byte{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payload cannot be empty")
}

func TestDCTTechnique_Embed_PayloadTooLarge(t *testing.T) {
	carrier := createTestJPEG(100, 100, 95)
	dct := NewDCTTechnique()

	// Calculate capacity
	capacity, _ := dct.CalculateCapacity(context.Background(), carrier)

	// Create payload larger than capacity
	payload := make([]byte, capacity+100)

	// Embed should fail
	_, err := dct.Embed(context.Background(), carrier, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payload too large")
}

func TestDCTTechnique_Embed_InvalidCarrier(t *testing.T) {
	dct := NewDCTTechnique()
	invalidCarrier := []byte("not a valid JPEG image")
	payload := []byte("test payload")

	_, err := dct.Embed(context.Background(), invalidCarrier, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode JPEG")
}

func TestDCTTechnique_Extract_InvalidCarrier(t *testing.T) {
	dct := NewDCTTechnique()
	invalidCarrier := []byte("not a valid JPEG image")

	_, err := dct.Extract(context.Background(), invalidCarrier)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode image") // Updated: now generic image decoding
}

func TestDCTTechnique_Extract_InvalidPayloadLength(t *testing.T) {
	carrier := createTestJPEG(100, 100, 95)
	dct := NewDCTTechnique()

	// First embed a normal payload to create a valid stego image
	normalPayload := []byte("test payload")
	stego, err := dct.Embed(context.Background(), carrier, normalPayload)
	require.NoError(t, err)

	// Now manually corrupt the length header in the stego image
	// Decode the stego image to access pixels directly
	img, _, err := image.Decode(bytes.NewReader(stego))
	require.NoError(t, err)
	rgba := imageToRGBA(img)

	// Manually corrupt the first 4 bytes (length header) by setting them to invalid value
	// We'll simulate setting length to 0xFFFFFFFF (max uint32 = ~4GB)
	corruptedLengthBytes := []byte{0xFF, 0xFF, 0xFF, 0xFF}

	// Embed the corrupted length bytes manually in the first pixels
	bounds := rgba.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	imageBitIndex := 0
	for _, b := range corruptedLengthBytes {
		for bit := 7; bit >= 0; bit-- { // MSB first
			bitValue := (b >> bit) & 1

			// Find next suitable pixel (middle brightness)
			for {
				y := imageBitIndex / (width * 3)
				x := (imageBitIndex / 3) % width
				channel := imageBitIndex % 3

				if y >= height {
					t.Fatal("Not enough pixels to embed corrupted length")
				}

				pixel := rgba.RGBAAt(x, y)
				brightness := (int(pixel.R) + int(pixel.G) + int(pixel.B)) / 3
				if brightness >= 64 && brightness <= 192 {
					// Modify the pixel
					var newValue uint8
					switch channel {
					case 0: // Red
						newValue = (pixel.R & 0xFE) | bitValue
						rgba.SetRGBA(x, y, color.RGBA{R: newValue, G: pixel.G, B: pixel.B, A: pixel.A})
					case 1: // Green
						newValue = (pixel.G & 0xFE) | bitValue
						rgba.SetRGBA(x, y, color.RGBA{R: pixel.R, G: newValue, B: pixel.B, A: pixel.A})
					case 2: // Blue
						newValue = (pixel.B & 0xFE) | bitValue
						rgba.SetRGBA(x, y, color.RGBA{R: pixel.R, G: pixel.G, B: newValue, A: pixel.A})
					}
					imageBitIndex++
					break
				} else {
					imageBitIndex++
				}
			}
		}
	}

	// Re-encode the corrupted image
	var corruptedBuf bytes.Buffer
	err = png.Encode(&corruptedBuf, rgba)
	require.NoError(t, err)

	// Extract should now fail due to invalid length
	_, err = dct.Extract(context.Background(), corruptedBuf.Bytes())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid payload length")
}

func TestDCTTechnique_Extract_ZeroPayloadLength(t *testing.T) {
	carrier := createTestJPEG(100, 100, 95)
	dct := NewDCTTechnique()

	// First embed a normal payload to create a valid stego image
	normalPayload := []byte("test")
	stego, err := dct.Embed(context.Background(), carrier, normalPayload)
	require.NoError(t, err)

	// Now manually set the length header to zero
	img, _, err := image.Decode(bytes.NewReader(stego))
	require.NoError(t, err)
	rgba := imageToRGBA(img)

	// Manually set the first 4 bytes (length header) to zero
	zeroLengthBytes := []byte{0x00, 0x00, 0x00, 0x00}

	// Embed the zero length bytes manually in the first pixels
	bounds := rgba.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	imageBitIndex := 0
	for _, b := range zeroLengthBytes {
		for bit := 7; bit >= 0; bit-- { // MSB first
			bitValue := (b >> bit) & 1

			// Find next suitable pixel (middle brightness)
			for {
				y := imageBitIndex / (width * 3)
				x := (imageBitIndex / 3) % width
				channel := imageBitIndex % 3

				if y >= height {
					t.Fatal("Not enough pixels to embed zero length")
				}

				pixel := rgba.RGBAAt(x, y)
				brightness := (int(pixel.R) + int(pixel.G) + int(pixel.B)) / 3
				if brightness >= 64 && brightness <= 192 {
					// Modify the pixel
					var newValue uint8
					switch channel {
					case 0: // Red
						newValue = (pixel.R & 0xFE) | bitValue
						rgba.SetRGBA(x, y, color.RGBA{R: newValue, G: pixel.G, B: pixel.B, A: pixel.A})
					case 1: // Green
						newValue = (pixel.G & 0xFE) | bitValue
						rgba.SetRGBA(x, y, color.RGBA{R: pixel.R, G: newValue, B: pixel.B, A: pixel.A})
					case 2: // Blue
						newValue = (pixel.B & 0xFE) | bitValue
						rgba.SetRGBA(x, y, color.RGBA{R: pixel.R, G: pixel.G, B: newValue, A: pixel.A})
					}
					imageBitIndex++
					break
				} else {
					imageBitIndex++
				}
			}
		}
	}

	// Re-encode the corrupted image
	var corruptedBuf bytes.Buffer
	err = png.Encode(&corruptedBuf, rgba)
	require.NoError(t, err)

	// Extract should now fail due to zero length
	_, err = dct.Extract(context.Background(), corruptedBuf.Bytes())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid payload length")
}

func TestDCTTechnique_JPEGQuality(t *testing.T) {
	carrier := createTestJPEG(300, 300, 95)
	dct := NewDCTTechnique()
	payload := []byte("Testing JPEG quality preservation")

	// Embed payload
	stego, err := dct.Embed(context.Background(), carrier, payload)
	require.NoError(t, err)

	// Decode and check it's valid image (now PNG output)
	img, _, err := image.Decode(bytes.NewReader(stego))
	require.NoError(t, err)
	assert.NotNil(t, img)

	// Verify dimensions preserved
	bounds := img.Bounds()
	assert.Equal(t, 300, bounds.Dx())
	assert.Equal(t, 300, bounds.Dy())
}

func TestDCTTechnique_MiddleBrightnessOnly(t *testing.T) {
	// Create image with varied brightness zones
	carrier := createVariedBrightnessJPEG(200, 200)
	dct := NewDCTTechnique()
	payload := []byte("Test middle brightness")

	// Embed and extract
	stego, err := dct.Embed(context.Background(), carrier, payload)
	require.NoError(t, err)

	extracted, err := dct.Extract(context.Background(), stego)
	require.NoError(t, err)

	assert.Equal(t, payload, extracted)
}

func TestDCTTechnique_ContextCancellation(t *testing.T) {
	carrier := createTestJPEG(200, 200, 95)
	dct := NewDCTTechnique()
	payload := []byte("test payload")

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Embed should fail immediately
	_, err := dct.Embed(ctx, carrier, payload)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Extract should also fail
	_, err = dct.Extract(ctx, carrier)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestDCTTechnique_MultipleEmbedExtractCycles(t *testing.T) {
	carrier := createTestJPEG(300, 300, 95)
	dct := NewDCTTechnique()
	payload1 := []byte("First embedding cycle")

	// First cycle
	stego1, err := dct.Embed(context.Background(), carrier, payload1)
	require.NoError(t, err)

	extracted1, err := dct.Extract(context.Background(), stego1)
	require.NoError(t, err)
	assert.Equal(t, payload1, extracted1)

	// Second cycle (use original carrier again since DCT requires JPEG input)
	payload2 := []byte("Second cycle data")
	stego2, err := dct.Embed(context.Background(), carrier, payload2)
	require.NoError(t, err)

	extracted2, err := dct.Extract(context.Background(), stego2)
	require.NoError(t, err)
	assert.Equal(t, payload2, extracted2)
}

// Helper function to create a test JPEG image
func createTestJPEG(width, height, quality int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with gradient pattern (varied brightness)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create gradient from dark to light
			brightness := uint8(64 + ((x+y)*128)/(width+height))
			img.Set(x, y, color.RGBA{
				R: brightness,
				G: brightness,
				B: brightness,
				A: 255,
			})
		}
	}

	// Encode as JPEG
	var buf bytes.Buffer
	opts := &jpeg.Options{Quality: quality}
	if err := jpeg.Encode(&buf, img, opts); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// createVariedBrightnessJPEG creates a JPEG with zones of different brightness
func createVariedBrightnessJPEG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create three zones: dark, middle, bright
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var brightness uint8
			third := width / 3

			if x < third {
				// Dark zone (< 64)
				brightness = 32
			} else if x < 2*third {
				// Middle zone (64-192) - this is where embedding happens
				brightness = 128
			} else {
				// Bright zone (> 192)
				brightness = 220
			}

			img.Set(x, y, color.RGBA{
				R: brightness,
				G: brightness,
				B: brightness,
				A: 255,
			})
		}
	}

	// Encode as JPEG
	var buf bytes.Buffer
	opts := &jpeg.Options{Quality: 95}
	if err := jpeg.Encode(&buf, img, opts); err != nil {
		panic(err)
	}

	return buf.Bytes()
}
