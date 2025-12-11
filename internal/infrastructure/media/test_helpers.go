package media

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
)

// createValidPNG creates a minimal valid PNG image for testing
func createValidPNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a simple pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// createValidJPEG creates a minimal valid JPEG header for testing
func createValidJPEG() []byte {
	// Minimal JPEG header with SOI, APP0, SOF0, SOS markers
	return []byte{
		0xFF, 0xD8, // SOI (Start of Image)
		0xFF, 0xE0, // APP0 marker
		0x00, 0x10, // APP0 length (16 bytes)
		0x4A, 0x46, 0x49, 0x46, 0x00, // "JFIF\0"
		0x01, 0x01, // Version 1.1
		0x00,       // Density units (0 = none)
		0x00, 0x01, // X density
		0x00, 0x01, // Y density
		0x00, 0x00, // Thumbnail dimensions
		0xFF, 0xD9, // EOI (End of Image)
	}
}

// createValidWAV creates a minimal valid WAV file for testing
func createValidWAV(sampleRate, numSamples int) []byte {
	var buf bytes.Buffer

	numChannels := 2
	bitsPerSample := 16
	byteRate := sampleRate * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8
	dataSize := numSamples * numChannels * bitsPerSample / 8

	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(36+dataSize)) // Chunk size
	buf.WriteString("WAVE")

	// fmt subchunk
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))            // Subchunk1Size (16 for PCM)
	binary.Write(&buf, binary.LittleEndian, uint16(1))             // AudioFormat (1 for PCM)
	binary.Write(&buf, binary.LittleEndian, uint16(numChannels))   // NumChannels
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))    // SampleRate
	binary.Write(&buf, binary.LittleEndian, uint32(byteRate))      // ByteRate
	binary.Write(&buf, binary.LittleEndian, uint16(blockAlign))    // BlockAlign
	binary.Write(&buf, binary.LittleEndian, uint16(bitsPerSample)) // BitsPerSample

	// data subchunk
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(dataSize))

	// Write sample data (simple silence or tone)
	for i := 0; i < numSamples*numChannels; i++ {
		binary.Write(&buf, binary.LittleEndian, int16(0))
	}

	return buf.Bytes()
}

// createValidFLAC creates a minimal FLAC header for testing
func createValidFLAC() []byte {
	return []byte{
		0x66, 0x4C, 0x61, 0x43, // "fLaC" magic bytes
		0x80,             // Last metadata block flag + STREAMINFO type
		0x00, 0x00, 0x22, // Block length (34 bytes)
		// Minimal STREAMINFO block (simplified)
		0x00, 0x10, // Min block size
		0x00, 0x10, // Max block size
		0x00, 0x00, 0x00, // Min frame size
		0x00, 0x00, 0x00, // Max frame size
		0x0A, 0xC4, 0x42, // Sample rate (44100) + channels + bits per sample
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Total samples
		// MD5 signature (16 bytes)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

// createValidText creates valid UTF-8 text for testing
func createValidText(wordCount int) []byte {
	text := "Lorem ipsum dolor sit amet consectetur adipiscing elit "
	var buf bytes.Buffer
	for i := 0; i < wordCount/8; i++ { // text has ~8 words
		buf.WriteString(text)
	}
	return buf.Bytes()
}
