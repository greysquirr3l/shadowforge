package stego

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

// TestJPEG_BitPreservation - Test if JPEG preserves bit modifications at different positions
func TestJPEG_BitPreservation(t *testing.T) {
	// Create test image with known pixel values
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// Set first pixel to a known value: RGB(128, 128, 128)
	testColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	img.Set(0, 0, testColor)

	// Fill rest with same color
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, testColor)
		}
	}

	fmt.Printf("Original G value: %d (%08b)\n", 128, 128)

	// Modify bit 1 (DCT current approach)
	modifiedG := uint8(128)
	modifiedG = (modifiedG & 0xFD) | (1 << 1) // Set bit 1
	fmt.Printf("Modified G value: %d (%08b) [bit 1 set]\n", modifiedG, modifiedG)

	// Set the modified pixel
	img.Set(0, 0, color.RGBA{R: 128, G: modifiedG, B: 128, A: 255})

	// Test different JPEG qualities
	qualities := []int{100, 95, 90}

	for _, quality := range qualities {
		t.Run(fmt.Sprintf("Quality_%d", quality), func(t *testing.T) {
			// Encode as JPEG
			var buf bytes.Buffer
			err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
			if err != nil {
				t.Fatal(err)
			}

			// Decode back
			decodedImg, err := jpeg.Decode(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatal(err)
			}

			// Check if bit modification was preserved
			rgbaImg := decodedImg.(*image.YCbCr)
			bounds := rgbaImg.Bounds()
			_, g, _, _ := rgbaImg.At(bounds.Min.X, bounds.Min.Y).RGBA()

			// Convert back to uint8 values
			decodedG := uint8(g >> 8)

			fmt.Printf("Quality %d: G %d (%08b) -> %d (%08b)\n",
				quality, modifiedG, modifiedG, decodedG, decodedG)

			// Check if bit 1 was preserved
			originalBit1 := (modifiedG >> 1) & 1
			decodedBit1 := (decodedG >> 1) & 1

			if originalBit1 != decodedBit1 {
				fmt.Printf("  ⚠️  Bit 1 NOT preserved! %d -> %d\n", originalBit1, decodedBit1)
			} else {
				fmt.Printf("  ✅ Bit 1 preserved! %d -> %d\n", originalBit1, decodedBit1)
			}

			// Check bit 0 for comparison
			originalBit0 := modifiedG & 1
			decodedBit0 := decodedG & 1

			if originalBit0 != decodedBit0 {
				fmt.Printf("  ⚠️  Bit 0 NOT preserved! %d -> %d\n", originalBit0, decodedBit0)
			} else {
				fmt.Printf("  ✅ Bit 0 preserved! %d -> %d\n", originalBit0, decodedBit0)
			}
		})
	}
}
