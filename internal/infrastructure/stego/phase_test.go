package stego

import (
	"context"
	"math"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
	"github.com/greysquirr3l/shadowforge/pkg/testutil"
)

func TestPhaseTechnique_NewPhaseTechnique(t *testing.T) {
	tests := []struct {
		name   string
		config PhaseEmbeddingConfig
		logger *logrus.Logger
	}{
		{
			name:   "with_default_config",
			config: DefaultPhaseConfig(),
			logger: logrus.New(),
		},
		{
			name: "with_custom_config",
			config: PhaseEmbeddingConfig{
				SegmentSize:    512,
				PhaseThreshold: 0.1,
				MinFrequency:   5,
				MaxFrequency:   256,
				OverlapFactor:  0.25,
			},
			logger: logrus.New(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			technique := NewPhaseTechnique(tt.config, tt.logger)

			assert.NotNil(t, technique)
			assert.Equal(t, "phase", technique.Name())
			assert.Equal(t, tt.config, technique.config)
		})
	}
}

func TestPhaseTechnique_NewPhaseWithDefaults(t *testing.T) {
	technique := NewPhaseWithDefaults(logrus.New())

	require.NotNil(t, technique)
	assert.Equal(t, "phase", technique.Name())

	// Verify default config
	expected := DefaultPhaseConfig()
	assert.Equal(t, expected, technique.config)
}

func TestPhaseTechnique_SupportsFormat(t *testing.T) {
	technique := NewPhaseWithDefaults(logrus.New())

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

func TestPhaseTechnique_CalculateCapacity(t *testing.T) {
	technique := NewPhaseWithDefaults(logrus.New())
	ctx := context.Background()

	tests := []struct {
		name        string
		numSamples  int
		expectError bool
		errorType   error
	}{
		{
			name:        "sufficient_capacity",
			numSamples:  100000, // 100000/256 - 32 = 358 bits
			expectError: false,
		},
		{
			name:        "small_but_sufficient",
			numSamples:  11000, // 11000/256 - 32 = 10 bits (>= 8 bits minimum)
			expectError: false,
		},
		{
			name:        "insufficient_capacity",
			numSamples:  2000, // 2000/256 - 32 < 8 bits (insufficient)
			expectError: true,
			errorType:   stego.ErrInsufficientCapacity,
		},
		{
			name:        "empty_carrier",
			numSamples:  0,
			expectError: true,
			errorType:   stego.ErrInsufficientCapacity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create WAV carrier with specified number of samples
			carrier := testutil.GenerateStereo16BitWAV(tt.numSamples)
			capacity, err := technique.CalculateCapacity(ctx, carrier)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.errorType)
				assert.Equal(t, 0, capacity)
			} else {
				require.NoError(t, err)
				assert.Greater(t, capacity, 0)
				// For DSSS: capacity = (samples / chipLength) - 32 bits
				// chipLength = 256
				expectedCapacity := (tt.numSamples / 256) - 32
				assert.Equal(t, expectedCapacity, capacity, "Capacity should match DSSS formula")
			}
		})
	}
}

func TestPhaseTechnique_EmbedAndExtract(t *testing.T) {
	technique := NewPhaseWithDefaults(logrus.New())
	ctx := context.Background()

	tests := []struct {
		name        string
		numSamples  int
		payload     []byte
		expectError bool
		errorType   error
	}{
		{
			name:        "simple_embed_extract",
			numSamples:  100000,
			payload:     []byte("test payload"),
			expectError: false,
		},
		{
			name:        "larger_payload",
			numSamples:  500000,
			payload:     []byte("This is a longer test payload with more data to embed in the audio file"),
			expectError: false,
		},
		{
			name:        "minimum_size",
			numSamples:  50000,
			payload:     []byte("short"),
			expectError: false,
		},
		{
			name:        "empty_payload",
			numSamples:  100000,
			payload:     []byte{},
			expectError: true,
			errorType:   stego.ErrEmptyPayload,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate proper WAV carrier with white noise
			carrier := testutil.GenerateStereo16BitWAV(tt.numSamples)

			// Test embedding
			stegoCarrier, err := technique.Embed(ctx, carrier, tt.payload)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.errorType)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, stegoCarrier)
			// Phase encoding modifies samples in-place, maintains WAV size
			assert.Equal(t, len(carrier), len(stegoCarrier),
				"Phase encoding should maintain WAV file size")

			// Test extraction
			extracted, err := technique.Extract(ctx, stegoCarrier)
			require.NoError(t, err)
			assert.NotNil(t, extracted)

			// Verify exact payload recovery
			assert.Equal(t, tt.payload, extracted, "Extracted payload should match original")
		})
	}
}

func TestPhaseTechnique_ExtractFromNonStegoCarrier(t *testing.T) {
	technique := NewPhaseWithDefaults(logrus.New())
	ctx := context.Background()

	// Generate proper WAV carrier without embedded data
	plainCarrier := testutil.GenerateStereo16BitWAV(50000)

	// Extracting from non-stego carrier should fail or return invalid data
	_, err := technique.Extract(ctx, plainCarrier)

	// For phase encoding, random phase values will decode to garbage length
	// The implementation should detect this and return an error
	require.Error(t, err, "Should fail to extract from non-stego carrier")
}

func TestPhaseTechnique_EmbedInsufficientCapacity(t *testing.T) {
	technique := NewPhaseWithDefaults(logrus.New())
	ctx := context.Background()

	// Generate small WAV file - not enough capacity
	smallCarrier := testutil.GenerateStereo16BitWAV(100)
	payload := []byte("test payload")

	_, err := technique.Embed(ctx, smallCarrier, payload)

	// Should fail due to insufficient capacity
	require.Error(t, err, "Should fail with insufficient capacity")
}

func TestDefaultPhaseConfig(t *testing.T) {
	config := DefaultPhaseConfig()

	assert.Equal(t, 256, config.SegmentSize) // ChipLength for DSSS - 24dB process gain
	assert.Equal(t, math.Pi/8, config.PhaseThreshold)
	assert.Equal(t, 10, config.MinFrequency)
	assert.Equal(t, 512, config.MaxFrequency)
	assert.Equal(t, 0.5, config.OverlapFactor)
}

func TestPhaseTechnique_InterfaceCompliance(t *testing.T) {
	technique := NewPhaseWithDefaults(logrus.New())

	// Verify interface compliance at compile time
	var _ stego.Technique = technique
	var _ stego.Embedder = technique
	var _ stego.Extractor = technique

	assert.NotNil(t, technique)
}

func TestPhaseTechnique_ConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config PhaseEmbeddingConfig
	}{
		{
			name: "valid_small_segments",
			config: PhaseEmbeddingConfig{
				SegmentSize:    256,
				PhaseThreshold: math.Pi / 16,
				MinFrequency:   1,
				MaxFrequency:   128,
				OverlapFactor:  0.25,
			},
		},
		{
			name: "valid_large_segments",
			config: PhaseEmbeddingConfig{
				SegmentSize:    4096,
				PhaseThreshold: math.Pi / 4,
				MinFrequency:   50,
				MaxFrequency:   2048,
				OverlapFactor:  0.75,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			technique := NewPhaseTechnique(tt.config, logrus.New())
			assert.NotNil(t, technique)
			assert.Equal(t, tt.config, technique.config)
		})
	}
}
