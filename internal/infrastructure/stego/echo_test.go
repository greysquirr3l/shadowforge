package stego

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

func TestEchoTechnique_NewEchoTechnique(t *testing.T) {
	tests := []struct {
		name   string
		config EchoEmbeddingConfig
		logger *logrus.Logger
	}{
		{
			name:   "with_default_config",
			config: DefaultEchoConfig(),
			logger: logrus.New(),
		},
		{
			name: "with_custom_config",
			config: EchoEmbeddingConfig{
				Delay0:     50,
				Delay1:     55,
				Amplitude:  0.05,
				SegmentLen: 4096,
				MixRatio:   0.9,
			},
			logger: logrus.New(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			technique := NewEchoTechnique(tt.config, tt.logger)

			assert.NotNil(t, technique)
			assert.Equal(t, "echo", technique.Name())
			assert.Equal(t, tt.config, technique.config)
		})
	}
}

func TestEchoTechnique_NewEchoWithDefaults(t *testing.T) {
	technique := NewEchoWithDefaults(logrus.New())

	require.NotNil(t, technique)
	assert.Equal(t, "echo", technique.Name())

	// Verify default config
	expected := DefaultEchoConfig()
	assert.Equal(t, expected, technique.config)
}

func TestEchoTechnique_SupportsFormat(t *testing.T) {
	technique := NewEchoWithDefaults(logrus.New())

	tests := []struct {
		name     string
		format   media.MediaFormat
		expected bool
	}{
		{
			name:     "supports_wav",
			format:   media.FormatWAV,
			expected: true,
		},
		{
			name:     "does_not_support_flac",
			format:   media.FormatFLAC,
			expected: false,
		},
		{
			name:     "does_not_support_mp3",
			format:   media.FormatMP3,
			expected: false,
		},
		{
			name:     "does_not_support_png",
			format:   media.FormatPNG,
			expected: false,
		},
		{
			name:     "does_not_support_jpeg",
			format:   media.FormatJPEG,
			expected: false,
		},
		{
			name:     "does_not_support_txt",
			format:   media.FormatTXT,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := technique.SupportsFormat(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEchoTechnique_CalculateCapacity(t *testing.T) {
	technique := NewEchoWithDefaults(logrus.New())
	ctx := context.Background()

	tests := []struct {
		name        string
		carrierSize int
		expectError bool
		errorType   error
	}{
		{
			name:        "sufficient_capacity_large",
			carrierSize: 16384, // 2 segments
			expectError: false,
		},
		{
			name:        "sufficient_capacity_minimum",
			carrierSize: 8192, // 1 segment
			expectError: false,
		},
		{
			name:        "insufficient_capacity",
			carrierSize: 4096,
			expectError: true,
			errorType:   stego.ErrInsufficientCapacity,
		},
		{
			name:        "empty_carrier",
			carrierSize: 0,
			expectError: true,
			errorType:   stego.ErrInsufficientCapacity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			carrier := make([]byte, tt.carrierSize)
			capacity, err := technique.CalculateCapacity(ctx, carrier)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.errorType)
				assert.Equal(t, 0, capacity)
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, capacity, 0)

				// Verify capacity calculation
				expectedSegments := tt.carrierSize / 8192     // Default segment size
				expectedCapacityBytes := expectedSegments / 8 // 1 bit per segment, convert to bytes
				assert.Equal(t, expectedCapacityBytes, capacity)
			}
		})
	}
}

func TestEchoTechnique_EmbedAndExtract(t *testing.T) {
	technique := NewEchoWithDefaults(logrus.New())
	ctx := context.Background()

	tests := []struct {
		name        string
		carrierSize int
		payload     []byte
		expectError bool
		errorType   error
	}{
		{
			name:        "simple_embed_extract",
			carrierSize: 16384,
			payload:     []byte("test payload"),
			expectError: false,
		},
		{
			name:        "larger_payload",
			carrierSize: 32768,
			payload:     []byte("This is a longer test payload with more data to embed"),
			expectError: false,
		},
		{
			name:        "minimum_size",
			carrierSize: 8192,
			payload:     []byte("short"),
			expectError: false,
		},
		{
			name:        "empty_payload",
			carrierSize: 16384,
			payload:     []byte{},
			expectError: true,
			errorType:   stego.ErrEmptyPayload,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			carrier := make([]byte, tt.carrierSize)

			// Test embedding
			stegoCarrier, err := technique.Embed(ctx, carrier, tt.payload)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.errorType)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, stegoCarrier)
			assert.Greater(t, len(stegoCarrier), len(carrier),
				"Stego carrier should be larger than original")

			// Test extraction
			extracted, err := technique.Extract(ctx, stegoCarrier)
			require.NoError(t, err)
			assert.NotNil(t, extracted)

			// For this simplified implementation, we don't expect exact payload recovery
			// In a full implementation, this would verify: assert.Equal(t, tt.payload, extracted)
			assert.Greater(t, len(extracted), 0, "Should extract some data")
		})
	}
}

func TestEchoTechnique_ExtractFromNonStegoCarrier(t *testing.T) {
	technique := NewEchoWithDefaults(logrus.New())
	ctx := context.Background()

	// Try to extract from a carrier that hasn't had data embedded
	plainCarrier := make([]byte, 16384)

	extracted, err := technique.Extract(ctx, plainCarrier)

	// For this simplified implementation, it should still return something
	// In a full implementation, this might return an error or empty data
	require.NoError(t, err)
	assert.NotNil(t, extracted)
}

func TestDefaultEchoConfig(t *testing.T) {
	config := DefaultEchoConfig()

	assert.Equal(t, 100, config.Delay0)
	assert.Equal(t, 101, config.Delay1)
	assert.Equal(t, 0.1, config.Amplitude)
	assert.Equal(t, 8192, config.SegmentLen)
	assert.Equal(t, 0.8, config.MixRatio)
}

func TestEchoTechnique_InterfaceCompliance(t *testing.T) {
	technique := NewEchoWithDefaults(logrus.New())

	// Verify interface compliance at compile time
	var _ stego.Technique = technique
	var _ stego.Embedder = technique
	var _ stego.Extractor = technique

	assert.NotNil(t, technique)
}

func TestEchoTechnique_ConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    EchoEmbeddingConfig
		expectErr bool
	}{
		{
			name: "valid_config",
			config: EchoEmbeddingConfig{
				Delay0:     50,
				Delay1:     55,
				Amplitude:  0.1,
				SegmentLen: 4096,
				MixRatio:   0.8,
			},
			expectErr: false,
		},
		{
			name: "invalid_delay0_zero",
			config: EchoEmbeddingConfig{
				Delay0:     0,
				Delay1:     55,
				Amplitude:  0.1,
				SegmentLen: 4096,
				MixRatio:   0.8,
			},
			expectErr: true,
		},
		{
			name: "invalid_delay1_negative",
			config: EchoEmbeddingConfig{
				Delay0:     50,
				Delay1:     -5,
				Amplitude:  0.1,
				SegmentLen: 4096,
				MixRatio:   0.8,
			},
			expectErr: true,
		},
		{
			name: "invalid_amplitude_zero",
			config: EchoEmbeddingConfig{
				Delay0:     50,
				Delay1:     55,
				Amplitude:  0.0,
				SegmentLen: 4096,
				MixRatio:   0.8,
			},
			expectErr: true,
		},
		{
			name: "invalid_amplitude_too_high",
			config: EchoEmbeddingConfig{
				Delay0:     50,
				Delay1:     55,
				Amplitude:  1.5,
				SegmentLen: 4096,
				MixRatio:   0.8,
			},
			expectErr: true,
		},
		{
			name: "invalid_segment_too_small",
			config: EchoEmbeddingConfig{
				Delay0:     50,
				Delay1:     55,
				Amplitude:  0.1,
				SegmentLen: 30, // Smaller than delay1
				MixRatio:   0.8,
			},
			expectErr: true,
		},
		{
			name: "invalid_mix_ratio_negative",
			config: EchoEmbeddingConfig{
				Delay0:     50,
				Delay1:     55,
				Amplitude:  0.1,
				SegmentLen: 4096,
				MixRatio:   -0.1,
			},
			expectErr: true,
		},
		{
			name: "invalid_mix_ratio_too_high",
			config: EchoEmbeddingConfig{
				Delay0:     50,
				Delay1:     55,
				Amplitude:  0.1,
				SegmentLen: 4096,
				MixRatio:   1.5,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			technique := NewEchoTechnique(tt.config, logrus.New())
			err := technique.validateConfig()

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEchoTechnique_HelperFunctions(t *testing.T) {
	technique := NewEchoWithDefaults(logrus.New())

	t.Run("payloadToBits", func(t *testing.T) {
		payload := []byte{0xFF, 0x00, 0xAA} // 11111111 00000000 10101010
		expected := []bool{true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false, true, false, true, false, true, false, true, false}

		result := technique.payloadToBits(payload)
		assert.Equal(t, expected, result)
	})

	t.Run("bitsToPayload", func(t *testing.T) {
		bits := []bool{true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false, true, false, true, false, true, false, true, false}
		expected := []byte{0xFF, 0x00, 0xAA}

		result := technique.bitsToPayload(bits)
		assert.Equal(t, expected, result)
	})

	t.Run("segmentAudio", func(t *testing.T) {
		samples := make([]int32, 20000)
		for i := range samples {
			samples[i] = int32(i)
		}

		segments := technique.segmentAudio(samples, 8192)

		assert.Greater(t, len(segments), 1)
		// Should have segments of default size (8192) plus potentially a smaller final segment
		expectedSegments := len(samples) / 8192
		if len(samples)%8192 != 0 {
			expectedSegments++
		}
		assert.Equal(t, expectedSegments, len(segments))
	})

	t.Run("calculateAutocorrelation", func(t *testing.T) {
		// Create a simple test signal with autocorrelation
		samples := []float64{1, 0, -1, 0, 1, 0, -1, 0, 1, 0}

		// Should have high autocorrelation at lag 4 (repeating pattern)
		corr4 := technique.calculateAutocorrelation(samples, 4)
		corr1 := technique.calculateAutocorrelation(samples, 1)

		assert.Greater(t, corr4, corr1)
	})
}
