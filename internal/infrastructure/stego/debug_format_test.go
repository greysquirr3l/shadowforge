package stego

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/jpeg"
	"testing"
)

// TestDCT_DebugImageFormat - Debug the image format and conversion process
func TestDCT_DebugImageFormat(t *testing.T) {
	technique := NewDCTTechnique()

	// Create minimal test carrier: 100x100 JPEG
	carrier := createDiagnosticJPEG(100, 100, 100)

	// Decode the JPEG to see what format we get
	img, err := jpeg.Decode(bytes.NewReader(carrier))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Decoded image type: %T\n", img)
	fmt.Printf("Bounds: %v\n", img.Bounds())

	// Convert to RGBA and check first pixel
	rgba := imageToRGBA(img)
	fmt.Printf("RGBA image type: %T\n", rgba)

	// Check first few pixels
	for y := 0; y < 3 && y < rgba.Bounds().Dy(); y++ {
		for x := 0; x < 3 && x < rgba.Bounds().Dx(); x++ {
			pixel := rgba.RGBAAt(x, y)
			brightness := (int(pixel.R) + int(pixel.G) + int(pixel.B)) / 3
			fmt.Printf("Pixel[%d,%d]: R=%d, G=%d, B=%d, Brightness=%d, Usable=%t\n",
				x, y, pixel.R, pixel.G, pixel.B, brightness, brightness >= 64 && brightness <= 192)
		}
	}

	// Test the embedding and extraction of length header only
	payload := []byte{0x42} // Single test byte
	fmt.Printf("\nOriginal payload length: %d\n", len(payload))

	// Manually embed just the length header (4 bytes)
	lengthHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthHeader, uint32(len(payload)))
	fmt.Printf("Length header bytes: %v\n", lengthHeader)

	// Embed length header
	err = technique.embedData(rgba, lengthHeader)
	if err != nil {
		t.Fatal(err)
	}

	// Extract length header
	extractedLength, err := technique.extractBytes(rgba, 4, 0)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Extracted length bytes: %v\n", extractedLength)
	extractedLengthValue := binary.BigEndian.Uint32(extractedLength)
	fmt.Printf("Extracted length value: %d\n", extractedLengthValue)
}
