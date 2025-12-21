// Package testutil provides test utilities for generating test data files
// including WAV audio files, images, and other media types used in tests.
package testutil

import (
	"bytes"
	"encoding/binary"
	"math/rand"
)

// WAVHeader represents a standard WAV file header structure.
type WAVHeader struct {
	ChunkID       [4]byte // "RIFF"
	ChunkSize     uint32  // File size - 8
	Format        [4]byte // "WAVE"
	Subchunk1ID   [4]byte // "fmt "
	Subchunk1Size uint32  // 16 for PCM
	AudioFormat   uint16  // 1 = PCM
	NumChannels   uint16  // Mono = 1, Stereo = 2
	SampleRate    uint32  // 8000, 44100, 48000, etc.
	ByteRate      uint32  // SampleRate * NumChannels * BitsPerSample/8
	BlockAlign    uint16  // NumChannels * BitsPerSample/8
	BitsPerSample uint16  // 8, 16, 24, 32
	Subchunk2ID   [4]byte // "data"
	Subchunk2Size uint32  // NumSamples * NumChannels * BitsPerSample/8
}

// WAVConfig contains configuration for WAV file generation.
type WAVConfig struct {
	SampleRate    int  // Sample rate in Hz (e.g., 44100)
	BitDepth      int  // Bits per sample (8, 16, 24, 32)
	Channels      int  // Number of audio channels (1=mono, 2=stereo)
	NumSamples    int  // Number of samples per channel
	GenerateTone  bool // If true, generate a test tone instead of silence
	ToneFrequency int  // Frequency of test tone in Hz (e.g., 440 for A4)
}

// DefaultWAVConfig returns a standard WAV configuration for testing.
func DefaultWAVConfig() WAVConfig {
	return WAVConfig{
		SampleRate:    44100,
		BitDepth:      16,
		Channels:      2,
		NumSamples:    1000,
		GenerateTone:  false,
		ToneFrequency: 0,
	}
}

// GenerateWAV creates a WAV file with the specified configuration.
// This is the recommended way to generate WAV files for testing.
func GenerateWAV(config WAVConfig) []byte {
	return generateWAVWithSamples(config.SampleRate, config.BitDepth, config.Channels, config.NumSamples)
}

// GenerateSimpleWAV creates a minimal valid WAV file with default parameters.
// Use this for simple tests that don't need custom configuration.
func GenerateSimpleWAV() []byte {
	return GenerateWAV(DefaultWAVConfig())
}

// GenerateStereo16BitWAV creates a stereo 16-bit WAV file at 44.1kHz.
// This is the most common format for testing.
func GenerateStereo16BitWAV(numSamples int) []byte {
	return generateWAVWithSamples(44100, 16, 2, numSamples)
}

// GenerateMono16BitWAV creates a mono 16-bit WAV file at 44.1kHz.
func GenerateMono16BitWAV(numSamples int) []byte {
	return generateWAVWithSamples(44100, 16, 1, numSamples)
}

// Generate8BitWAV creates an 8-bit mono WAV file.
func Generate8BitWAV(numSamples int) []byte {
	return generateWAVWithSamples(8000, 8, 1, numSamples)
}

// Generate24BitWAV creates a 24-bit stereo WAV file at 48kHz.
func Generate24BitWAV(numSamples int) []byte {
	return generateWAVWithSamples(48000, 24, 2, numSamples)
}

// Generate32BitWAV creates a 32-bit stereo WAV file at 48kHz.
func Generate32BitWAV(numSamples int) []byte {
	return generateWAVWithSamples(48000, 32, 2, numSamples)
}

// generateWAVWithSamples creates a minimal valid WAV file for testing.
// This is the core implementation used by all WAV generation functions.
// Generates WHITE NOISE which has ideal properties for echo hiding detection:
// - No periodic structure that would confuse autocorrelation
// - Flat autocorrelation at non-zero lags, making echoes clearly detectable
func generateWAVWithSamples(sampleRate, bitDepth, channels, numSamples int) []byte {
	bytesPerSample := bitDepth / 8
	dataSize := uint32(numSamples * channels * bytesPerSample)

	header := WAVHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     36 + dataSize,
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1, // PCM
		NumChannels:   uint16(channels),
		SampleRate:    uint32(sampleRate),
		ByteRate:      uint32(sampleRate * channels * bytesPerSample),
		BlockAlign:    uint16(channels * bytesPerSample),
		BitsPerSample: uint16(bitDepth),
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: dataSize,
	}

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, &header); err != nil {
		return nil
	}

	// Use deterministic seed for reproducible tests
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < numSamples; i++ {
		// Generate pure WHITE noise - random values in [-1, 1]
		// White noise has near-zero autocorrelation at non-zero lags,
		// making it ideal for detecting added echoes via autocorrelation.
		sampleValue := (rng.Float64()*2 - 1) * 0.7 // Scale to 70% amplitude

		// Clamp to [-1, 1]
		if sampleValue > 1.0 {
			sampleValue = 1.0
		}
		if sampleValue < -1.0 {
			sampleValue = -1.0
		}

		// Write same sample value to all channels
		for ch := 0; ch < channels; ch++ {
			switch bitDepth {
			case 8:
				// 8-bit unsigned (0-255, center at 128)
				value := uint8(128 + sampleValue*127)
				buf.WriteByte(value)
			case 16:
				// 16-bit signed (-32768 to 32767)
				value := int16(sampleValue * 32767)
				if err := binary.Write(&buf, binary.LittleEndian, value); err != nil {
					return nil
				}
			case 24:
				// 24-bit signed
				value := int32(sampleValue * 8388607) // 2^23 - 1
				buf.WriteByte(byte(value))
				buf.WriteByte(byte(value >> 8))
				buf.WriteByte(byte(value >> 16))
			case 32:
				// 32-bit signed
				value := int32(sampleValue * 2147483647)
				if err := binary.Write(&buf, binary.LittleEndian, value); err != nil {
					return nil
				}
			}
		}
	}

	return buf.Bytes()
}

// GetWAVCapacity calculates the approximate capacity in bytes for steganography.
// This assumes LSB encoding (1 bit per sample).
func GetWAVCapacity(config WAVConfig) int {
	totalSamples := config.NumSamples * config.Channels
	// LSB: 1 bit per sample = total_samples / 8 bytes
	return totalSamples / 8
}

// GetWAVCapacityForFile calculates capacity from a generated WAV file.
func GetWAVCapacityForFile(wavData []byte) int {
	if len(wavData) < 44 {
		return 0
	}

	// Read header to get sample count
	reader := bytes.NewReader(wavData)
	var header WAVHeader
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return 0
	}

	// Calculate total samples
	bytesPerSample := int(header.BitsPerSample) / 8
	totalSamples := int(header.Subchunk2Size) / (int(header.NumChannels) * bytesPerSample)

	// LSB capacity: 1 bit per sample
	return totalSamples * int(header.NumChannels) / 8
}
