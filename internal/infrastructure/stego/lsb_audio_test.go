package stego

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

func TestLSBAudioTechnique_NewLSBAudioTechnique(t *testing.T) {
	config := DefaultLSBAudioConfig()
	logger := slog.Default()

	technique := NewLSBAudioTechnique(config, logger)

	assert.NotNil(t, technique)
	assert.Equal(t, "lsb_audio", technique.Name())
	assert.Equal(t, config.BitsPerSample, technique.config.BitsPerSample)
	assert.Equal(t, config.ChannelMask, technique.config.ChannelMask)
}

func TestLSBAudioTechnique_NewLSBAudioWithDefaults(t *testing.T) {
	technique := NewLSBAudioWithDefaults(nil)

	assert.NotNil(t, technique)
	assert.Equal(t, 1, technique.config.BitsPerSample)
	assert.Equal(t, uint8(3), technique.config.ChannelMask)
	assert.Equal(t, 1, technique.config.SampleInterval)
	assert.Equal(t, 0.01, technique.config.AmplitudeThreshold)
	assert.False(t, technique.config.UseRandomOrder)
	assert.True(t, technique.config.PreserveSilence)
	assert.Equal(t, 3, technique.config.QualityLevel)
}

func TestLSBAudioTechnique_ConfigValidation(t *testing.T) {
	tests := []struct {
		name              string
		config            LSBAudioEmbeddingConfig
		expectedBits      int
		expectedMask      uint8
		expectedInterval  int
		expectedThreshold float64
		expectedQuality   int
	}{
		{
			name: "valid_config",
			config: LSBAudioEmbeddingConfig{
				BitsPerSample:      2,
				ChannelMask:        1,
				SampleInterval:     2,
				AmplitudeThreshold: 0.05,
				QualityLevel:       4,
			},
			expectedBits:      2,
			expectedMask:      1,
			expectedInterval:  2,
			expectedThreshold: 0.05,
			expectedQuality:   4,
		},
		{
			name:         "invalid_bits_per_sample_high",
			config:       LSBAudioEmbeddingConfig{BitsPerSample: 8},
			expectedBits: 1,
		},
		{
			name:         "invalid_bits_per_sample_low",
			config:       LSBAudioEmbeddingConfig{BitsPerSample: 0},
			expectedBits: 1,
		},
		{
			name:         "invalid_channel_mask",
			config:       LSBAudioEmbeddingConfig{ChannelMask: 0},
			expectedMask: 1,
		},
		{
			name:             "invalid_sample_interval",
			config:           LSBAudioEmbeddingConfig{SampleInterval: 0},
			expectedInterval: 1,
		},
		{
			name:              "invalid_amplitude_threshold_high",
			config:            LSBAudioEmbeddingConfig{AmplitudeThreshold: 2.0},
			expectedThreshold: 0.01,
		},
		{
			name:              "invalid_amplitude_threshold_low",
			config:            LSBAudioEmbeddingConfig{AmplitudeThreshold: -0.5},
			expectedThreshold: 0.01,
		},
		{
			name:            "invalid_quality_level_high",
			config:          LSBAudioEmbeddingConfig{QualityLevel: 10},
			expectedQuality: 3,
		},
		{
			name:            "invalid_quality_level_low",
			config:          LSBAudioEmbeddingConfig{QualityLevel: 0},
			expectedQuality: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			technique := NewLSBAudioTechnique(tt.config, nil)

			if tt.expectedBits != 0 {
				assert.Equal(t, tt.expectedBits, technique.config.BitsPerSample)
			}
			if tt.expectedMask != 0 {
				assert.Equal(t, tt.expectedMask, technique.config.ChannelMask)
			}
			if tt.expectedInterval != 0 {
				assert.Equal(t, tt.expectedInterval, technique.config.SampleInterval)
			}
			if tt.expectedThreshold != 0 {
				assert.Equal(t, tt.expectedThreshold, technique.config.AmplitudeThreshold)
			}
			if tt.expectedQuality != 0 {
				assert.Equal(t, tt.expectedQuality, technique.config.QualityLevel)
			}
		})
	}
}

func TestLSBAudioTechnique_SupportsFormat(t *testing.T) {
	technique := NewLSBAudioWithDefaults(nil)

	tests := []struct {
		name     string
		format   media.MediaFormat
		expected bool
	}{
		{"WAV format", media.FormatWAV, true},
		{"PNG format", media.FormatPNG, false},
		{"JPEG format", media.FormatJPEG, false},
		{"TXT format", media.FormatTXT, false},
		{"Unknown format", media.MediaFormat("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := technique.SupportsFormat(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLSBAudioTechnique_PayloadToBits(t *testing.T) {
	technique := NewLSBAudioWithDefaults(nil)

	tests := []struct {
		name     string
		payload  []byte
		expected []byte
	}{
		{
			name:     "single_byte_0xFF",
			payload:  []byte{0xFF},
			expected: []byte{1, 1, 1, 1, 1, 1, 1, 1},
		},
		{
			name:     "single_byte_0x00",
			payload:  []byte{0x00},
			expected: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "single_byte_0xAA",
			payload:  []byte{0xAA},
			expected: []byte{1, 0, 1, 0, 1, 0, 1, 0},
		},
		{
			name:     "two_bytes",
			payload:  []byte{0xF0, 0x0F},
			expected: []byte{1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1},
		},
		{
			name:     "empty_payload",
			payload:  []byte{},
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bits := technique.payloadToBits(tt.payload)
			assert.Equal(t, tt.expected, bits)
		})
	}
}

func TestLSBAudioTechnique_BitsToPayload(t *testing.T) {
	technique := NewLSBAudioWithDefaults(nil)

	tests := []struct {
		name     string
		bits     []byte
		expected []byte
	}{
		{
			name:     "single_byte_bits_0xFF",
			bits:     []byte{1, 1, 1, 1, 1, 1, 1, 1},
			expected: []byte{0xFF},
		},
		{
			name:     "single_byte_bits_0x00",
			bits:     []byte{0, 0, 0, 0, 0, 0, 0, 0},
			expected: []byte{0x00},
		},
		{
			name:     "single_byte_bits_0xAA",
			bits:     []byte{1, 0, 1, 0, 1, 0, 1, 0},
			expected: []byte{0xAA},
		},
		{
			name:     "two_bytes_bits",
			bits:     []byte{1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1},
			expected: []byte{0xF0, 0x0F},
		},
		{
			name:     "unaligned_bits_padded",
			bits:     []byte{1, 0, 1}, // Will be padded to 10100000
			expected: []byte{0xA0},
		},
		{
			name:     "empty_bits",
			bits:     []byte{},
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := technique.bitsToPayload(tt.bits)
			assert.Equal(t, tt.expected, payload)
		})
	}
}

func TestLSBAudioTechnique_CreateMockAudioData(t *testing.T) {
	// Helper function to create mock WAV data for testing
	createMockWAVData := func(numSamples int) []byte {
		headerSize := 44
		channels := 2
		bitDepth := 16
		sampleRate := 44100

		audioDataSize := numSamples * channels * (bitDepth / 8)
		totalSize := headerSize + audioDataSize

		data := make([]byte, totalSize)

		// Simple WAV header
		copy(data[0:4], "RIFF")
		fileSize := uint32(totalSize - 8)
		data[4] = byte(fileSize)
		data[5] = byte(fileSize >> 8)
		data[6] = byte(fileSize >> 16)
		data[7] = byte(fileSize >> 24)

		copy(data[8:12], "WAVE")
		copy(data[12:16], "fmt ")

		// fmt chunk size
		data[16] = 16
		data[17] = 0
		data[18] = 0
		data[19] = 0

		// Audio format (PCM)
		data[20] = 1
		data[21] = 0

		// Channels
		data[22] = byte(channels)
		data[23] = 0

		// Sample rate
		data[24] = byte(sampleRate)
		data[25] = byte(sampleRate >> 8)
		data[26] = byte(sampleRate >> 16)
		data[27] = byte(sampleRate >> 24)

		// Byte rate
		byteRate := sampleRate * channels * (bitDepth / 8)
		data[28] = byte(byteRate)
		data[29] = byte(byteRate >> 8)
		data[30] = byte(byteRate >> 16)
		data[31] = byte(byteRate >> 24)

		// Block align
		blockAlign := channels * (bitDepth / 8)
		data[32] = byte(blockAlign)
		data[33] = 0

		// Bits per sample
		data[34] = byte(bitDepth)
		data[35] = 0

		// Data chunk
		copy(data[36:40], "data")

		// Data chunk size
		data[40] = byte(audioDataSize)
		data[41] = byte(audioDataSize >> 8)
		data[42] = byte(audioDataSize >> 16)
		data[43] = byte(audioDataSize >> 24)

		// Generate sample audio data (simple sine wave)
		for i := 0; i < numSamples; i++ {
			// Generate a simple pattern for testing
			sample := int16((i % 1000) - 500) // Simple sawtooth pattern

			for ch := 0; ch < channels; ch++ {
				offset := headerSize + (i*channels+ch)*2
				data[offset] = byte(sample)
				data[offset+1] = byte(sample >> 8)
			}
		}

		return data
	}

	// Test with small audio sample
	audioData := createMockWAVData(1000)
	assert.Greater(t, len(audioData), 44, "Audio data should be larger than header")

	// Verify header
	assert.Equal(t, "RIFF", string(audioData[0:4]))
	assert.Equal(t, "WAVE", string(audioData[8:12]))
	assert.Equal(t, "fmt ", string(audioData[12:16]))
	assert.Equal(t, "data", string(audioData[36:40]))

	t.Logf("Created mock audio data: %d bytes", len(audioData))
}

func TestLSBAudioTechnique_ParseAudioData(t *testing.T) {
	technique := NewLSBAudioWithDefaults(nil)

	// Create mock audio data
	createMockAudio := func(samples int) []byte {
		data := make([]byte, 44+samples*2*2) // Header + stereo 16-bit samples
		copy(data[0:4], "RIFF")
		copy(data[8:12], "WAVE")
		copy(data[12:16], "fmt ")
		copy(data[36:40], "data")
		return data
	}

	tests := []struct {
		name          string
		audioData     []byte
		expectError   bool
		expectedError error
	}{
		{
			name:        "valid_audio_data",
			audioData:   createMockAudio(1000),
			expectError: false,
		},
		{
			name:          "too_small_data",
			audioData:     make([]byte, 20),
			expectError:   true,
			expectedError: stego.ErrEmptyCoverMedia,
		},
		{
			name:          "header_only",
			audioData:     make([]byte, 44),
			expectError:   true,
			expectedError: stego.ErrEmptyCoverMedia,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			samples, sampleRate, bitDepth, channels, err := technique.parseAudioData(tt.audioData)

			if tt.expectError {
				require.Error(t, err)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, 44100, sampleRate)
			assert.Equal(t, 16, bitDepth)
			assert.Equal(t, 2, channels)
			assert.Greater(t, len(samples), 0)
			assert.Greater(t, len(samples[0]), 0)
		})
	}
}

func TestLSBAudioTechnique_CalculateCapacity(t *testing.T) {
	technique := NewLSBAudioWithDefaults(nil)
	ctx := context.Background()

	// Create mock audio with known sample count
	createMockAudio := func(samples int) []byte {
		headerSize := 44
		audioDataSize := samples * 2 * 2 // stereo 16-bit
		totalSize := headerSize + audioDataSize

		data := make([]byte, totalSize)
		copy(data[0:4], "RIFF")
		copy(data[8:12], "WAVE")
		copy(data[12:16], "fmt ")
		copy(data[36:40], "data")

		// Set data chunk size
		data[40] = byte(audioDataSize)
		data[41] = byte(audioDataSize >> 8)
		data[42] = byte(audioDataSize >> 16)
		data[43] = byte(audioDataSize >> 24)

		return data
	}

	tests := []struct {
		name        string
		audioData   []byte
		expectError bool
		minCapacity int
	}{
		{
			name:        "small_audio_1000_samples",
			audioData:   createMockAudio(1000),
			expectError: false,
			minCapacity: 100, // Conservative estimate
		},
		{
			name:        "large_audio_10000_samples",
			audioData:   createMockAudio(10000),
			expectError: false,
			minCapacity: 1000,
		},
		{
			name:        "empty_audio",
			audioData:   []byte{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity, err := technique.CalculateCapacity(ctx, tt.audioData)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.GreaterOrEqual(t, capacity, tt.minCapacity)
			t.Logf("Calculated capacity: %d bytes for audio size: %d", capacity, len(tt.audioData))
		})
	}
}

func TestLSBAudioTechnique_EmbedExtract(t *testing.T) {
	technique := NewLSBAudioWithDefaults(nil)
	ctx := context.Background()

	// Create mock audio data
	createMockAudio := func(samples int) []byte {
		headerSize := 44
		audioDataSize := samples * 2 * 2 // stereo 16-bit
		totalSize := headerSize + audioDataSize

		data := make([]byte, totalSize)

		// WAV header
		copy(data[0:4], "RIFF")
		fileSize := uint32(totalSize - 8)
		data[4] = byte(fileSize)
		data[5] = byte(fileSize >> 8)
		data[6] = byte(fileSize >> 16)
		data[7] = byte(fileSize >> 24)

		copy(data[8:12], "WAVE")
		copy(data[12:16], "fmt ")

		data[16] = 16 // fmt chunk size
		data[20] = 1  // PCM format
		data[22] = 2  // stereo

		// Sample rate 44100
		sampleRate := uint32(44100)
		data[24] = byte(sampleRate)
		data[25] = byte(sampleRate >> 8)
		data[26] = byte(sampleRate >> 16)
		data[27] = byte(sampleRate >> 24)

		// Byte rate
		byteRate := uint32(44100 * 2 * 2)
		data[28] = byte(byteRate)
		data[29] = byte(byteRate >> 8)
		data[30] = byte(byteRate >> 16)
		data[31] = byte(byteRate >> 24)

		data[32] = 4  // block align
		data[34] = 16 // bits per sample

		copy(data[36:40], "data")

		// Data chunk size
		data[40] = byte(audioDataSize)
		data[41] = byte(audioDataSize >> 8)
		data[42] = byte(audioDataSize >> 16)
		data[43] = byte(audioDataSize >> 24)

		// Sample data (non-zero to avoid amplitude threshold issues)
		for i := 0; i < samples*2*2; i += 2 {
			sample := int16(1000 + (i/2)%1000) // Non-zero samples
			data[headerSize+i] = byte(sample)
			data[headerSize+i+1] = byte(sample >> 8)
		}

		return data
	}

	tests := []struct {
		name        string
		audioData   []byte
		payload     []byte
		expectError bool
		errorType   error
	}{
		{
			name:        "small_payload_success",
			audioData:   createMockAudio(5000),
			payload:     []byte("test"),
			expectError: false,
		},
		{
			name:        "single_byte_payload",
			audioData:   createMockAudio(1000),
			payload:     []byte("A"),
			expectError: false,
		},
		{
			name:        "empty_payload",
			audioData:   createMockAudio(1000),
			payload:     []byte{},
			expectError: true,
			errorType:   stego.ErrEmptyPayload,
		},
		{
			name:        "empty_carrier",
			audioData:   []byte{},
			payload:     []byte("test"),
			expectError: true,
			errorType:   stego.ErrEmptyCoverMedia,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test embedding
			stegoAudio, err := technique.Embed(ctx, tt.audioData, tt.payload)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, stegoAudio)
			assert.Greater(t, len(stegoAudio), 44, "Stego audio should include header")

			// Test extraction
			extractedPayload, err := technique.Extract(ctx, stegoAudio)
			require.NoError(t, err)
			assert.NotEmpty(t, extractedPayload)

			// Verify payload (may have padding due to byte boundary)
			assert.True(t, len(extractedPayload) >= len(tt.payload))
			if len(extractedPayload) >= len(tt.payload) {
				assert.Equal(t, tt.payload, extractedPayload[:len(tt.payload)])
			}
		})
	}
}

func TestLSBAudioTechnique_DifferentConfigurations(t *testing.T) {
	ctx := context.Background()

	// Create mock audio
	createMockAudio := func() []byte {
		headerSize := 44
		samples := 2000
		audioDataSize := samples * 2 * 2
		totalSize := headerSize + audioDataSize

		data := make([]byte, totalSize)
		copy(data[0:4], "RIFF")
		copy(data[8:12], "WAVE")
		copy(data[12:16], "fmt ")
		copy(data[36:40], "data")

		fileSize := uint32(totalSize - 8)
		data[4] = byte(fileSize)
		data[5] = byte(fileSize >> 8)
		data[6] = byte(fileSize >> 16)
		data[7] = byte(fileSize >> 24)

		data[16] = 16 // fmt chunk size
		data[20] = 1  // PCM
		data[22] = 2  // stereo
		data[32] = 4  // block align
		data[34] = 16 // bits per sample

		data[40] = byte(audioDataSize)
		data[41] = byte(audioDataSize >> 8)
		data[42] = byte(audioDataSize >> 16)
		data[43] = byte(audioDataSize >> 24)

		// Non-zero sample data
		for i := 0; i < audioDataSize; i += 2 {
			sample := int16(5000) // Constant non-zero amplitude
			data[headerSize+i] = byte(sample)
			data[headerSize+i+1] = byte(sample >> 8)
		}

		return data
	}

	tests := []struct {
		name   string
		config LSBAudioEmbeddingConfig
	}{
		{
			name: "single_bit_left_channel",
			config: LSBAudioEmbeddingConfig{
				BitsPerSample:      1,
				ChannelMask:        1, // Left only
				SampleInterval:     1,
				AmplitudeThreshold: 0.0,
				QualityLevel:       5,
			},
		},
		{
			name: "multi_bit_both_channels",
			config: LSBAudioEmbeddingConfig{
				BitsPerSample:      2,
				ChannelMask:        3, // Both channels
				SampleInterval:     1,
				AmplitudeThreshold: 0.0,
				QualityLevel:       5,
			},
		},
		{
			name: "sparse_embedding",
			config: LSBAudioEmbeddingConfig{
				BitsPerSample:      1,
				ChannelMask:        3,
				SampleInterval:     4, // Every 4th sample
				AmplitudeThreshold: 0.0,
				QualityLevel:       3,
			},
		},
	}

	audioData := createMockAudio()
	payload := []byte("test")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			technique := NewLSBAudioTechnique(tt.config, nil)

			// Check capacity
			capacity, err := technique.CalculateCapacity(ctx, audioData)
			require.NoError(t, err)

			if capacity < len(payload) {
				t.Skip("Insufficient capacity for this configuration")
			}

			// Test embed/extract cycle
			stegoAudio, err := technique.Embed(ctx, audioData, payload)
			require.NoError(t, err)

			extractedPayload, err := technique.Extract(ctx, stegoAudio)
			require.NoError(t, err)

			assert.True(t, len(extractedPayload) >= len(payload))
			assert.Equal(t, payload, extractedPayload[:len(payload)])

			t.Logf("Config %s: capacity=%d, embedded successfully", tt.name, capacity)
		})
	}
}

func TestLSBAudioTechnique_ErrorConditions(t *testing.T) {
	technique := NewLSBAudioWithDefaults(nil)
	ctx := context.Background()

	t.Run("extract_from_plain_audio", func(t *testing.T) {
		// Create audio with no embedded data
		plainAudio := make([]byte, 1000)
		copy(plainAudio[0:4], "RIFF")
		copy(plainAudio[8:12], "WAVE")

		_, err := technique.Extract(ctx, plainAudio)
		// LSB audio extraction from plain audio may succeed with garbage data
		// or may return ErrNoEmbeddedData - both are acceptable behaviors
		if err != nil {
			assert.ErrorIs(t, err, stego.ErrNoEmbeddedData)
		}
	})

	t.Run("invalid_audio_format", func(t *testing.T) {
		invalidAudio := []byte("not audio data")

		_, err := technique.Embed(ctx, invalidAudio, []byte("test"))
		assert.Error(t, err)

		_, err = technique.Extract(ctx, invalidAudio)
		assert.Error(t, err)

		_, err = technique.CalculateCapacity(ctx, invalidAudio)
		assert.Error(t, err)
	})
}

// Benchmark tests

func BenchmarkLSBAudioTechnique_Embed(b *testing.B) {
	technique := NewLSBAudioWithDefaults(nil)
	ctx := context.Background()

	// Create reasonably sized audio data
	headerSize := 44
	samples := 44100 // 1 second of audio
	audioDataSize := samples * 2 * 2
	audioData := make([]byte, headerSize+audioDataSize)

	copy(audioData[0:4], "RIFF")
	copy(audioData[8:12], "WAVE")
	copy(audioData[12:16], "fmt ")
	copy(audioData[36:40], "data")

	payload := []byte("benchmark test payload for LSB audio embedding")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := technique.Embed(ctx, audioData, payload)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLSBAudioTechnique_Extract(b *testing.B) {
	technique := NewLSBAudioWithDefaults(nil)
	ctx := context.Background()

	// Create audio data with embedded payload
	headerSize := 44
	samples := 44100
	audioDataSize := samples * 2 * 2
	audioData := make([]byte, headerSize+audioDataSize)

	copy(audioData[0:4], "RIFF")
	copy(audioData[8:12], "WAVE")
	copy(audioData[12:16], "fmt ")
	copy(audioData[36:40], "data")

	payload := []byte("benchmark payload")
	stegoAudio, err := technique.Embed(ctx, audioData, payload)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := technique.Extract(ctx, stegoAudio)
		if err != nil {
			b.Fatal(err)
		}
	}
}
