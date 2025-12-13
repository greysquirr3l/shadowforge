package stego

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/jpeg"
	"testing"
)

// TestDCT_JPEGRoundTrip - Test if JPEG re-encoding destroys the embedded data
func TestDCT_JPEGRoundTrip(t *testing.T) {
	technique := NewDCTTechnique()

	// Step 1: Create initial JPEG
	carrier := createDiagnosticJPEG(100, 100, 100)
	fmt.Printf("=== Step 1: Initial JPEG Created ===\n")

	// Step 2: Decode to RGBA
	img, err := jpeg.Decode(bytes.NewReader(carrier))
	if err != nil {
		t.Fatal(err)
	}
	rgba := imageToRGBA(img)
	fmt.Printf("=== Step 2: JPEG Decoded to RGBA ===\n")

	// Step 3: Embed test data directly in RGBA
	testPayload := []byte{0x42}
	lengthHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthHeader, uint32(len(testPayload)))
	fullData := append(lengthHeader, testPayload...)

	err = technique.embedData(rgba, fullData)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("=== Step 3: Data Embedded in RGBA ===\n")

	// Step 4: Check data is still there BEFORE re-encoding
	extractedBefore, err := technique.extractBytes(rgba, 4, 0)
	if err != nil {
		t.Fatal(err)
	}
	lengthBefore := binary.BigEndian.Uint32(extractedBefore)
	fmt.Printf("Length before JPEG re-encoding: %d\n", lengthBefore)

	// Step 5: Re-encode to JPEG (this is where data might be lost)
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, rgba, &jpeg.Options{Quality: technique.quality})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("=== Step 5: RGBA Re-encoded to JPEG ===\n")

	// Step 6: Decode the new JPEG
	newImg, err := jpeg.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	newRgba := imageToRGBA(newImg)
	fmt.Printf("=== Step 6: New JPEG Decoded to RGBA ===\n")

	// Step 7: Check if data survived the round-trip
	extractedAfter, err := technique.extractBytes(newRgba, 4, 0)
	if err != nil {
		t.Fatal(err)
	}
	lengthAfter := binary.BigEndian.Uint32(extractedAfter)
	fmt.Printf("Length after JPEG round-trip: %d\n", lengthAfter)

	if lengthBefore != lengthAfter {
		fmt.Printf("❌ JPEG round-trip corrupted the data! %d -> %d\n", lengthBefore, lengthAfter)

		// Compare first few pixels to see what changed
		fmt.Printf("\nPixel comparison (first 3 pixels):\n")
		for i := 0; i < 3; i++ {
			x, y := i%100, i/100
			oldPixel := rgba.RGBAAt(x, y)
			newPixel := newRgba.RGBAAt(x, y)
			fmt.Printf("Pixel[%d,%d]: G=%d->%d (LSB: %d->%d)\n",
				x, y, oldPixel.G, newPixel.G, oldPixel.G&1, newPixel.G&1)
		}
	} else {
		fmt.Printf("✅ JPEG round-trip preserved the data! %d == %d\n", lengthBefore, lengthAfter)
	}
}
