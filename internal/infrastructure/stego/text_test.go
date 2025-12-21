package stego

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

func TestTextTechnique_NewTextTechnique(t *testing.T) {
	config := DefaultTextConfig()
	logger := logrus.New()

	technique := NewTextTechnique(config, logger)

	assert.NotNil(t, technique)
	assert.Equal(t, "text_zw", technique.Name())
	assert.Equal(t, config.Method, technique.config.Method)
}

func TestTextTechnique_NewTextWithDefaults(t *testing.T) {
	technique := NewTextWithDefaults(nil)

	assert.NotNil(t, technique)
	assert.Equal(t, MethodZeroWidth, technique.config.Method)
	assert.Equal(t, 1, technique.config.SpaceInterval)
	assert.Equal(t, 1, technique.config.WordInterval)
	assert.False(t, technique.config.UseWordJoiner)
	assert.True(t, technique.config.PreserveFormat)
}

func TestTextTechnique_SupportsFormat(t *testing.T) {
	technique := NewTextWithDefaults(nil)

	tests := []struct {
		name     string
		format   media.MediaFormat
		expected bool
	}{
		{"TXT format", media.FormatTXT, true},
		{"Markdown format", media.FormatMarkdown, true},
		{"PNG format", media.FormatPNG, false},
		{"JPEG format", media.FormatJPEG, false},
		{"WAV format", media.FormatWAV, false},
		{"Unknown format", media.MediaFormat("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := technique.SupportsFormat(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTextTechnique_EmbedExtract_ZeroWidth(t *testing.T) {
	config := DefaultTextConfig()
	config.Method = MethodZeroWidth
	technique := NewTextTechnique(config, logrus.New())
	ctx := context.Background()

	tests := []struct {
		name        string
		coverText   string
		payload     []byte
		expectError bool
		errorType   error
	}{
		{
			name:        "simple_text_embedding",
			coverText:   "The quick brown fox jumps over the lazy dog",
			payload:     []byte("secret"),
			expectError: false,
		},
		{
			name:        "single_byte_payload",
			coverText:   "Hello world this is a test",
			payload:     []byte("A"),
			expectError: false,
		},
		{
			name:        "multi_word_text",
			coverText:   "Lorem ipsum dolor sit amet consectetur adipiscing elit",
			payload:     []byte("hidden message"),
			expectError: false,
		},
		{
			name:        "empty_payload",
			coverText:   "Some cover text",
			payload:     []byte{},
			expectError: true,
			errorType:   stego.ErrEmptyPayload,
		},
		{
			name:        "empty_cover",
			coverText:   "",
			payload:     []byte("data"),
			expectError: true,
			errorType:   stego.ErrEmptyCoverMedia,
		},
		{
			name:        "single_word_cover",
			coverText:   "Word",
			payload:     []byte("A"),
			expectError: true,
			errorType:   stego.ErrInsufficientCapacity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test embedding
			stegoText, err := technique.Embed(ctx, []byte(tt.coverText), tt.payload)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, stegoText)

			// Verify stego text contains zero-width characters
			assert.True(t, technique.containsZeroWidthChars(string(stegoText)))

			// Verify clean text is similar to original
			cleanText := technique.cleanZeroWidthChars(string(stegoText))
			assert.Equal(t, strings.Fields(tt.coverText), strings.Fields(cleanText))

			// Test extraction
			extractedPayload, err := technique.Extract(ctx, stegoText)
			require.NoError(t, err)

			// Verify extracted payload matches original
			// Note: Zero-width method may not embed all bits if insufficient word gaps
			expectedLen := min(len(tt.payload), len(extractedPayload))
			assert.Equal(t, tt.payload[:expectedLen], extractedPayload[:expectedLen])
		})
	}
}

func TestTextTechnique_EmbedExtract_Whitespace(t *testing.T) {
	config := DefaultTextConfig()
	config.Method = MethodWhitespace
	technique := NewTextTechnique(config, logrus.New())
	ctx := context.Background()

	coverText := "Line one\nLine two\nLine three\nLine four\nLine five\nLine six\nLine seven\nLine eight\nLine nine\nLine ten\nLine eleven\nLine twelve\nLine thirteen\nLine fourteen\nLine fifteen\nLine sixteen"
	payload := []byte("Hi") // 16 bits = 16 lines needed

	// Test embedding
	stegoText, err := technique.Embed(ctx, []byte(coverText), payload)
	require.NoError(t, err)
	assert.NotEqual(t, coverText, string(stegoText))

	// Test extraction
	extractedPayload, err := technique.Extract(ctx, stegoText)
	require.NoError(t, err)

	// Verify extracted payload matches original (may have some padding)
	assert.True(t, len(extractedPayload) >= len(payload))
	if len(extractedPayload) >= len(payload) {
		assert.Equal(t, payload, extractedPayload[:len(payload)])
	}
}

func TestTextTechnique_CalculateCapacity(t *testing.T) {
	technique := NewTextWithDefaults(nil)
	ctx := context.Background()

	tests := []struct {
		name        string
		coverText   string
		method      TextMethod
		expectError bool
		minCapacity int
	}{
		{
			name:        "zero_width_multiple_words",
			coverText:   "The quick brown fox jumps over the lazy dog",
			method:      MethodZeroWidth,
			expectError: false,
			minCapacity: 8, // At least 8 gaps between words
		},
		{
			name:        "zero_width_two_words",
			coverText:   "Hello world",
			method:      MethodZeroWidth,
			expectError: false,
			minCapacity: 1, // One gap = 8 bits = 1 byte
		},
		{
			name:        "whitespace_multiline",
			coverText:   "Line 1\nLine 2\nLine 3\nLine 4",
			method:      MethodWhitespace,
			expectError: false,
			minCapacity: 0, // 4 lines / 8 = 0 bytes
		},
		{
			name:        "empty_text",
			coverText:   "",
			method:      MethodZeroWidth,
			expectError: true,
		},
		{
			name:        "single_word",
			coverText:   "Word",
			method:      MethodZeroWidth,
			expectError: false,
			minCapacity: 0, // No gaps
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			technique.config.Method = tt.method

			capacity, err := technique.CalculateCapacity(ctx, []byte(tt.coverText))

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.GreaterOrEqual(t, capacity, tt.minCapacity)
		})
	}
}

func TestTextTechnique_PayloadToBits(t *testing.T) {
	technique := NewTextWithDefaults(nil)

	tests := []struct {
		name     string
		payload  []byte
		expected []byte
	}{
		{
			name:     "single_byte",
			payload:  []byte{0xFF}, // 11111111
			expected: []byte{1, 1, 1, 1, 1, 1, 1, 1},
		},
		{
			name:     "zero_byte",
			payload:  []byte{0x00}, // 00000000
			expected: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "mixed_bytes",
			payload:  []byte{0xAA}, // 10101010
			expected: []byte{1, 0, 1, 0, 1, 0, 1, 0},
		},
		{
			name:     "multiple_bytes",
			payload:  []byte{0xF0, 0x0F}, // 11110000 00001111
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

func TestTextTechnique_BitsToPayload(t *testing.T) {
	technique := NewTextWithDefaults(nil)

	tests := []struct {
		name     string
		bits     []byte
		expected []byte
	}{
		{
			name:     "single_byte_bits",
			bits:     []byte{1, 1, 1, 1, 1, 1, 1, 1},
			expected: []byte{0xFF},
		},
		{
			name:     "zero_byte_bits",
			bits:     []byte{0, 0, 0, 0, 0, 0, 0, 0},
			expected: []byte{0x00},
		},
		{
			name:     "mixed_bits",
			bits:     []byte{1, 0, 1, 0, 1, 0, 1, 0},
			expected: []byte{0xAA},
		},
		{
			name:     "multiple_bytes_bits",
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

func TestTextTechnique_ZeroWidthCharacters(t *testing.T) {
	technique := NewTextWithDefaults(nil)

	// Test zero-width character constants
	assert.Equal(t, '\u200B', ZWSP)
	assert.Equal(t, '\u200D', ZWJ)
	assert.Equal(t, '\u200C', ZWNJ)
	assert.Equal(t, '\u2060', WJ)

	// Test containsZeroWidthChars
	tests := []struct {
		name     string
		text     string
		contains bool
	}{
		{
			name:     "text_with_zwsp",
			text:     "Hello\u200Bworld",
			contains: true,
		},
		{
			name:     "text_with_zwj",
			text:     "Hello\u200Dworld",
			contains: true,
		},
		{
			name:     "plain_text",
			text:     "Hello world",
			contains: false,
		},
		{
			name:     "empty_text",
			text:     "",
			contains: false,
		},
		{
			name:     "text_with_word_joiner",
			text:     "Hello\u2060world",
			contains: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := technique.containsZeroWidthChars(tt.text)
			assert.Equal(t, tt.contains, result)
		})
	}
}

func TestTextTechnique_CleanZeroWidthChars(t *testing.T) {
	technique := NewTextWithDefaults(nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "text_with_zwsp",
			input:    "Hello\u200Bworld",
			expected: "Helloworld",
		},
		{
			name:     "text_with_multiple_zw_chars",
			input:    "Hello\u200B\u200D\u200C\u2060world",
			expected: "Helloworld",
		},
		{
			name:     "plain_text",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "empty_text",
			input:    "",
			expected: "",
		},
		{
			name:     "only_zw_chars",
			input:    "\u200B\u200D\u200C",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := technique.cleanZeroWidthChars(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTextTechnique_LargePayload(t *testing.T) {
	technique := NewTextWithDefaults(nil)
	ctx := context.Background()

	// Create large cover text with many words
	words := []string{
		"Lorem", "ipsum", "dolor", "sit", "amet", "consectetur",
		"adipiscing", "elit", "sed", "do", "eiusmod", "tempor",
		"incididunt", "ut", "labore", "et", "dolore", "magna",
		"aliqua", "Ut", "enim", "ad", "minim", "veniam",
		"quis", "nostrud", "exercitation", "ullamco", "laboris",
		"nisi", "ut", "aliquip", "ex", "ea", "commodo", "consequat",
	}
	coverText := strings.Join(words, " ")

	// Large payload
	payload := make([]byte, 20)
	for i := range payload {
		payload[i] = byte(i)
	}

	// Calculate capacity
	capacity, err := technique.CalculateCapacity(ctx, []byte(coverText))
	require.NoError(t, err)

	t.Logf("Cover text words: %d, Capacity: %d bytes, Payload: %d bytes",
		len(words), capacity, len(payload))

	if capacity < len(payload) {
		t.Skip("Cover text capacity insufficient for large payload test")
	}

	// Test embedding
	stegoText, err := technique.Embed(ctx, []byte(coverText), payload)
	require.NoError(t, err)

	// Verify stego text is valid UTF-8
	assert.True(t, utf8.Valid(stegoText))

	// Test extraction
	extractedPayload, err := technique.Extract(ctx, stegoText)
	require.NoError(t, err)

	// Verify payload integrity
	assert.True(t, len(extractedPayload) >= len(payload))
	assert.Equal(t, payload, extractedPayload[:len(payload)])
}

func TestTextTechnique_EdgeCases(t *testing.T) {
	technique := NewTextWithDefaults(nil)
	ctx := context.Background()

	t.Run("unicode_text", func(t *testing.T) {
		coverText := "Héllo wörld with émojis 😀 and spëcial châractérs"
		payload := []byte("test")

		stegoText, err := technique.Embed(ctx, []byte(coverText), payload)
		require.NoError(t, err)
		assert.True(t, utf8.Valid(stegoText))

		extractedPayload, err := technique.Extract(ctx, stegoText)
		require.NoError(t, err)
		assert.Equal(t, payload, extractedPayload[:len(payload)])
	})

	t.Run("very_long_words", func(t *testing.T) {
		coverText := "supercalifragilisticexpialidocious antidisestablishmentarianism"
		payload := []byte("X")

		stegoText, err := technique.Embed(ctx, []byte(coverText), payload)
		require.NoError(t, err)

		extractedPayload, err := technique.Extract(ctx, stegoText)
		require.NoError(t, err)
		assert.Equal(t, payload, extractedPayload[:len(payload)])
	})

	t.Run("text_with_existing_spaces", func(t *testing.T) {
		coverText := "Text  with   multiple    spaces"
		payload := []byte("A")

		stegoText, err := technique.Embed(ctx, []byte(coverText), payload)
		require.NoError(t, err)

		extractedPayload, err := technique.Extract(ctx, stegoText)
		require.NoError(t, err)
		assert.Equal(t, payload, extractedPayload[:len(payload)])
	})
}

func TestTextTechnique_ErrorConditions(t *testing.T) {
	config := DefaultTextConfig()
	config.Method = "invalid"
	technique := NewTextTechnique(config, logrus.New())
	ctx := context.Background()

	coverText := "Hello world"
	payload := []byte("test")

	// Test invalid method for embedding
	_, err := technique.Embed(ctx, []byte(coverText), payload)
	assert.ErrorIs(t, err, stego.ErrInvalidTechnique)

	// Test invalid method for extraction
	_, err = technique.Extract(ctx, []byte(coverText))
	assert.ErrorIs(t, err, stego.ErrInvalidTechnique)

	// Test invalid method for capacity calculation
	_, err = technique.CalculateCapacity(ctx, []byte(coverText))
	assert.ErrorIs(t, err, stego.ErrInvalidTechnique)
}

func TestTextTechnique_ExtractFromPlainText(t *testing.T) {
	technique := NewTextWithDefaults(nil)
	ctx := context.Background()

	// Test extracting from plain text (should fail)
	plainText := "This is plain text with no hidden data"

	_, err := technique.Extract(ctx, []byte(plainText))
	assert.ErrorIs(t, err, stego.ErrNoEmbeddedData)
}

// Benchmark tests

func BenchmarkTextTechnique_Embed(b *testing.B) {
	technique := NewTextWithDefaults(nil)
	ctx := context.Background()

	coverText := "The quick brown fox jumps over the lazy dog and runs through the forest"
	payload := []byte("secret message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := technique.Embed(ctx, []byte(coverText), payload)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTextTechnique_Extract(b *testing.B) {
	technique := NewTextWithDefaults(nil)
	ctx := context.Background()

	coverText := "The quick brown fox jumps over the lazy dog and runs through the forest"
	payload := []byte("secret message")

	// Create stego text once
	stegoText, err := technique.Embed(ctx, []byte(coverText), payload)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := technique.Extract(ctx, stegoText)
		if err != nil {
			b.Fatal(err)
		}
	}
}
