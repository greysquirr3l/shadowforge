package stego

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDCT_MinimalDiagnostic - Create minimal test case to isolate bit manipulation bug
func TestDCT_MinimalDiagnostic(t *testing.T) {
	technique := NewDCTTechnique()

	// Create minimal test carrier: 100x100 JPEG with middle brightness
	carrier := createDiagnosticJPEG(100, 100, 100)

	// Test single byte with known bit pattern
	testByte := byte(0x00) // Binary: 00000000
	payload := []byte{testByte}

	fmt.Printf("=== DCT Diagnostic Test ===\n")
	fmt.Printf("Original payload: %02x (binary: %08b)\n", testByte, testByte)

	// Step 1: Embed single byte
	fmt.Printf("\n--- EMBEDDING PHASE ---\n")
	ctx := context.Background()
	stego, err := technique.Embed(ctx, carrier, payload)
	require.NoError(t, err)
	fmt.Printf("Embedding completed without error\n")

	// Step 2: Extract and compare
	fmt.Printf("\n--- EXTRACTION PHASE ---\n")
	extractedPayload, err := technique.Extract(ctx, stego)
	require.NoError(t, err)

	fmt.Printf("Extracted payload length: %d bytes\n", len(extractedPayload))
	if len(extractedPayload) > 0 {
		extractedByte := extractedPayload[0]
		fmt.Printf("Extracted payload: %02x (binary: %08b)\n", extractedByte, extractedByte)
		fmt.Printf("Expected:         %02x (binary: %08b)\n", testByte, testByte)

		if extractedByte != testByte {
			// Analyze bit differences
			xor := extractedByte ^ testByte
			fmt.Printf("XOR difference:   %02x (binary: %08b)\n", xor, xor)
			fmt.Printf("Bits that differ: ")
			for i := 0; i < 8; i++ {
				if (xor>>i)&1 == 1 {
					fmt.Printf("bit_%d ", i)
				}
			}
			fmt.Printf("\n")
		}
	}

	// Assert correctness
	assert.Equal(t, payload, extractedPayload, "Single byte should be extracted correctly")
}

// TestDCT_BitPositionTest - Test if the issue is bit position (0 vs 1)
func TestDCT_BitPositionTest(t *testing.T) {
	technique := NewDCTTechnique()

	// Test all single-bit patterns to understand which bits are affected
	testBytes := []byte{
		0x01, // 00000001 - bit 0 set
		0x02, // 00000010 - bit 1 set
		0x04, // 00000100 - bit 2 set
		0x08, // 00001000 - bit 3 set
	}

	for i, testByte := range testBytes {
		t.Run(fmt.Sprintf("SingleBit_%d_Value_%02x", i, testByte), func(t *testing.T) {
			payload := []byte{testByte}
			carrier := createDiagnosticJPEG(100, 100, 100)
			ctx := context.Background()

			stego, err := technique.Embed(ctx, carrier, payload)
			require.NoError(t, err)

			extractedPayload, err := technique.Extract(ctx, stego)
			require.NoError(t, err)

			if len(extractedPayload) > 0 {
				fmt.Printf("Input: %02x (%08b) → Output: %02x (%08b)\n",
					testByte, testByte, extractedPayload[0], extractedPayload[0])
			}

			assert.Equal(t, payload, extractedPayload)
		})
	}
}

// createDiagnosticJPEG creates a JPEG image with specified dimensions and quality for diagnostics
func createDiagnosticJPEG(width, height, quality int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create varied image instead of solid gray to avoid JPEG optimization issues
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create slight color variation while keeping in the 64-192 brightness range
			r := 120 + (x % 8)       // 120-127
			g := 120 + (y % 8)       // 120-127
			b := 120 + ((x + y) % 8) // 120-127
			img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255})
		}
	}

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		panic(fmt.Sprintf("Failed to encode diagnostic JPEG: %v", err))
	}

	return buf.Bytes()
}
