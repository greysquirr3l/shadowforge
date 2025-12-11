package media

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTextProcessor tests constructor.
func TestNewTextProcessor(t *testing.T) {
	processor := NewTextProcessor()

	assert.NotNil(t, processor)
	assert.NotNil(t, processor.detector)
}

// TestLoadText_ValidUTF8 tests loading valid UTF-8 text.
func TestLoadText_ValidUTF8(t *testing.T) {
	processor := NewTextProcessor()
	data := []byte("Hello, World!\nThis is a test.")

	textData, err := processor.LoadText(data)

	require.NoError(t, err)
	assert.Equal(t, "Hello, World!\nThis is a test.", textData.Content)
	assert.Equal(t, "UTF-8", textData.Encoding)
	assert.Equal(t, LineEndingLF, textData.LineEnding)
	assert.Equal(t, 29, textData.CharCount) // 15 + 1 newline + 13
	assert.Equal(t, 2, textData.LineCount)
	assert.Equal(t, 6, textData.WordCount)
}

// TestLoadText_EmptyData tests loading empty data.
func TestLoadText_EmptyData(t *testing.T) {
	processor := NewTextProcessor()

	_, err := processor.LoadText([]byte{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestLoadText_InvalidUTF8 tests loading invalid UTF-8.
func TestLoadText_InvalidUTF8(t *testing.T) {
	processor := NewTextProcessor()
	// Invalid UTF-8 sequence
	data := []byte{0xFF, 0xFE, 0xFD}

	_, err := processor.LoadText(data)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UTF-8")
}

// TestLoadText_Markdown tests detecting Markdown.
func TestLoadText_Markdown(t *testing.T) {
	processor := NewTextProcessor()
	data := []byte("# Heading\n\nThis is **bold** and *italic*.\n\n- List item 1\n- List item 2")

	textData, err := processor.LoadText(data)

	require.NoError(t, err)
	assert.True(t, textData.IsMarkdown)
	// Word count includes markdown symbols: "Heading", "This", "is", "**bold**", "and", "*italic*.", "-", "List", "item", "1", "-", "List", "item", "2" = ~14 words
	assert.Greater(t, textData.WordCount, 10) // At least 10 words
	assert.Less(t, textData.WordCount, 20)    // Less than 20 words
}

// TestLoadText_LineEndings tests line ending detection.
func TestLoadText_LineEndings(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected LineEnding
	}{
		{"LF (Unix)", "line1\nline2\nline3", LineEndingLF},
		{"CRLF (Windows)", "line1\r\nline2\r\nline3", LineEndingCRLF},
		{"CR (Old Mac)", "line1\rline2\rline3", LineEndingCR},
		{"No newlines", "single line", LineEndingUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTextProcessor()
			data := []byte(tt.content)

			textData, err := processor.LoadText(data)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, textData.LineEnding)
		})
	}
}

// TestSaveText_Valid tests saving valid text.
func TestSaveText_Valid(t *testing.T) {
	processor := NewTextProcessor()
	textData := &TextData{
		Content:    "Test content",
		Encoding:   "UTF-8",
		LineEnding: LineEndingLF,
	}

	data, err := processor.SaveText(textData)

	require.NoError(t, err)
	assert.Equal(t, []byte("Test content"), data)
}

// TestSaveText_NilData tests saving nil data.
func TestSaveText_NilData(t *testing.T) {
	processor := NewTextProcessor()

	_, err := processor.SaveText(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

// TestSaveText_EmptyContent tests saving empty content.
func TestSaveText_EmptyContent(t *testing.T) {
	processor := NewTextProcessor()
	textData := &TextData{
		Content: "",
	}

	_, err := processor.SaveText(textData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestSaveText_InvalidUTF8 tests saving invalid UTF-8.
func TestSaveText_InvalidUTF8(t *testing.T) {
	processor := NewTextProcessor()
	// Create invalid UTF-8 string (Go strings are supposed to be valid UTF-8, but we can force it)
	textData := &TextData{
		Content: string([]byte{0xFF, 0xFE, 0xFD}),
	}

	_, err := processor.SaveText(textData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UTF-8")
}

// TestRoundTrip tests load and save cycle.
func TestRoundTrip(t *testing.T) {
	processor := NewTextProcessor()
	original := []byte("Hello, 世界!\n\nThis is a multi-line\ntest with Unicode: 🎉")

	// Load
	textData, err := processor.LoadText(original)
	require.NoError(t, err)

	// Save
	saved, err := processor.SaveText(textData)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original, saved)
}

// TestNormalizeLineEndings tests line ending normalization.
func TestNormalizeLineEndings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		target   LineEnding
		expected string
	}{
		{
			name:     "CRLF to LF",
			input:    "line1\r\nline2\r\nline3",
			target:   LineEndingLF,
			expected: "line1\nline2\nline3",
		},
		{
			name:     "LF to CRLF",
			input:    "line1\nline2\nline3",
			target:   LineEndingCRLF,
			expected: "line1\r\nline2\r\nline3",
		},
		{
			name:     "Mixed to LF",
			input:    "line1\r\nline2\nline3\rline4",
			target:   LineEndingLF,
			expected: "line1\nline2\nline3\nline4",
		},
		{
			name:     "LF to CR",
			input:    "line1\nline2\nline3",
			target:   LineEndingCR,
			expected: "line1\rline2\rline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTextProcessor()

			result, err := processor.NormalizeLineEndings(tt.input, tt.target)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNormalizeLineEndings_UnknownTarget tests normalizing to unknown target.
func TestNormalizeLineEndings_UnknownTarget(t *testing.T) {
	processor := NewTextProcessor()

	_, err := processor.NormalizeLineEndings("test", LineEndingUnknown)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown")
}

// TestCalculateCapacity_ZeroWidth tests zero-width capacity calculation.
func TestCalculateCapacity_ZeroWidth(t *testing.T) {
	processor := NewTextProcessor()
	text := []byte("The quick brown fox jumps over the lazy dog. " +
		"This sentence has many words for testing zero-width embedding.")

	capacity, err := processor.CalculateCapacity(text, "zero-width")

	require.NoError(t, err)
	assert.Greater(t, capacity.TotalCapacity, int64(0))
	assert.Greater(t, capacity.UsableCapacity, int64(0))
	assert.LessOrEqual(t, capacity.UsableCapacity, capacity.TotalCapacity)

	// Verify calculation: ~17 words / 8 bits per byte
	expectedRaw := int64(17 / 8)
	assert.Equal(t, expectedRaw, capacity.TotalCapacity)
}

// TestCalculateCapacity_Whitespace tests whitespace capacity.
func TestCalculateCapacity_Whitespace(t *testing.T) {
	processor := NewTextProcessor()
	// Need 8+ lines for capacity > 0 (lines / 8 bits per byte)
	text := []byte("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\n")

	capacity, err := processor.CalculateCapacity(text, "whitespace")

	require.NoError(t, err)
	assert.Greater(t, capacity.TotalCapacity, int64(0))
}

// TestCalculateCapacity_LineEnding tests line-ending capacity.
func TestCalculateCapacity_LineEnding(t *testing.T) {
	processor := NewTextProcessor()
	text := []byte(strings.Repeat("Line\n", 20))

	capacity, err := processor.CalculateCapacity(text, "line-ending")

	require.NoError(t, err)
	assert.Greater(t, capacity.TotalCapacity, int64(0))

	// Verify: 20 lines / 8 bits = 2 bytes
	expectedRaw := int64(20 / 8)
	assert.Equal(t, expectedRaw, capacity.TotalCapacity)
}

// TestCalculateCapacity_Homoglyph tests homoglyph capacity.
func TestCalculateCapacity_Homoglyph(t *testing.T) {
	processor := NewTextProcessor()
	// Need 80+ chars for capacity > 0 (chars * 0.1 / 8 bits per byte = 80 * 0.1 / 8 = 1 byte)
	text := []byte("This is a test with lots of characters for homoglyph substitution. Need more characters here. And even more.")

	capacity, err := processor.CalculateCapacity(text, "homoglyph")

	require.NoError(t, err)
	assert.Greater(t, capacity.TotalCapacity, int64(0))
}

// TestTextCalculateCapacity_UnknownTechnique tests unknown technique.
func TestTextCalculateCapacity_UnknownTechnique(t *testing.T) {
	processor := NewTextProcessor()
	text := []byte("Test text")

	_, err := processor.CalculateCapacity(text, "unknown")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown technique")
}

// TestCalculateCapacity_InvalidText tests capacity with invalid text.
func TestCalculateCapacity_InvalidText(t *testing.T) {
	processor := NewTextProcessor()
	text := []byte{0xFF, 0xFE}

	_, err := processor.CalculateCapacity(text, "zero-width")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load text")
}

// TestTextCalculateQualityScore tests quality scoring.
func TestTextCalculateQualityScore(t *testing.T) {
	tests := []struct {
		name     string
		textData *TextData
		minScore float64
		maxScore float64
	}{
		{
			name: "High quality (long text, many words)",
			textData: &TextData{
				CharCount:  1500,
				WordCount:  300,
				LineCount:  60,
				IsMarkdown: true,
			},
			minScore: 0.9,
			maxScore: 1.0,
		},
		{
			name: "Medium quality",
			textData: &TextData{
				CharCount:  700,
				WordCount:  150,
				LineCount:  30,
				IsMarkdown: false,
			},
			minScore: 0.6,
			maxScore: 0.8,
		},
		{
			name: "Low quality (short text)",
			textData: &TextData{
				CharCount: 100,
				WordCount: 20,
				LineCount: 5,
			},
			minScore: 0.3,
			maxScore: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTextProcessor()

			score := processor.calculateQualityScore(tt.textData)

			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, tt.maxScore)
			assert.LessOrEqual(t, score, 1.0)
			assert.GreaterOrEqual(t, score, 0.0)
		})
	}
}

// TestValidateText_Valid tests validating valid text.
func TestValidateText_Valid(t *testing.T) {
	processor := NewTextProcessor()
	data := []byte("Valid UTF-8 text 世界")

	err := processor.ValidateText(data)

	assert.NoError(t, err)
}

// TestValidateText_Empty tests validating empty text.
func TestValidateText_Empty(t *testing.T) {
	processor := NewTextProcessor()

	err := processor.ValidateText([]byte{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestValidateText_InvalidUTF8 tests validating invalid UTF-8.
func TestValidateText_InvalidUTF8(t *testing.T) {
	processor := NewTextProcessor()
	data := []byte{0xFF, 0xFE, 0xFD}

	err := processor.ValidateText(data)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UTF-8")
}

// TestValidateText_NullBytes tests validation with null bytes.
func TestValidateText_NullBytes(t *testing.T) {
	processor := NewTextProcessor()
	data := []byte("Text with\x00null byte")

	err := processor.ValidateText(data)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "null bytes")
}

// TestEmbedZeroWidth_Basic tests basic zero-width embedding.
func TestEmbedZeroWidth_Basic(t *testing.T) {
	processor := NewTextProcessor()
	coverText := "The quick brown fox jumps over the lazy dog"
	data := []byte{0xAB} // 10101011

	stegoText, err := processor.EmbedZeroWidth(coverText, data)

	require.NoError(t, err)
	assert.NotEqual(t, coverText, stegoText)
	assert.Greater(t, len(stegoText), len(coverText))

	// Should contain zero-width characters
	assert.True(t, strings.ContainsRune(stegoText, ZWSP) || strings.ContainsRune(stegoText, ZWJ))
}

// TestEmbedZeroWidth_EmptyCover tests embedding with empty cover.
func TestEmbedZeroWidth_EmptyCover(t *testing.T) {
	processor := NewTextProcessor()

	_, err := processor.EmbedZeroWidth("", []byte{0xAB})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestEmbedZeroWidth_EmptyData tests embedding empty data.
func TestEmbedZeroWidth_EmptyData(t *testing.T) {
	processor := NewTextProcessor()

	_, err := processor.EmbedZeroWidth("Cover text", []byte{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestEmbedZeroWidth_InsufficientCapacity tests embedding with insufficient capacity.
func TestEmbedZeroWidth_InsufficientCapacity(t *testing.T) {
	processor := NewTextProcessor()
	coverText := "short"       // Only 1 space
	data := []byte{0xAB, 0xCD} // 16 bits, needs 16 spaces

	_, err := processor.EmbedZeroWidth(coverText, data)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient capacity")
}

// TestExtractZeroWidth_Basic tests basic zero-width extraction.
func TestExtractZeroWidth_Basic(t *testing.T) {
	processor := NewTextProcessor()

	// Create stego text with known pattern: 10101011 (0xAB)
	stegoText := "Word" + string(ZWJ) + " " + "test" + string(ZWSP) + " " +
		"with" + string(ZWJ) + " " + "zero" + string(ZWSP) + " " +
		"width" + string(ZWJ) + " " + "chars" + string(ZWSP) + " " +
		"here" + string(ZWJ) + " " + "now" + string(ZWJ)

	data, err := processor.ExtractZeroWidth(stegoText)

	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Equal(t, byte(0xAB), data[0])
}

// TestExtractZeroWidth_EmptyText tests extraction from empty text.
func TestExtractZeroWidth_EmptyText(t *testing.T) {
	processor := NewTextProcessor()

	_, err := processor.ExtractZeroWidth("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestExtractZeroWidth_NoZeroWidth tests extraction with no zero-width chars.
func TestExtractZeroWidth_NoZeroWidth(t *testing.T) {
	processor := NewTextProcessor()
	text := "Normal text without zero-width characters"

	_, err := processor.ExtractZeroWidth(text)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no zero-width")
}

// TestZeroWidthRoundTrip tests embed and extract cycle.
func TestZeroWidthRoundTrip(t *testing.T) {
	processor := NewTextProcessor()
	// Need 32+ spaces for "Test" (4 bytes = 32 bits)
	coverText := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 4) // 4 * 9 spaces = 36 spaces
	original := []byte("Test")

	// Embed
	stegoText, err := processor.EmbedZeroWidth(coverText, original)
	require.NoError(t, err)

	// Extract
	extracted, err := processor.ExtractZeroWidth(stegoText)
	require.NoError(t, err)

	// Compare (may have padding, so check prefix)
	assert.Equal(t, original, extracted[:len(original)])
}

// TestManipulateWhitespace_Basic tests whitespace manipulation.
func TestManipulateWhitespace_Basic(t *testing.T) {
	processor := NewTextProcessor()
	text := "Line 1  \nLine 2\t\nLine 3   "

	// Modifier: add one space to each line
	modifier := func(line int, ws string) string {
		return ws + " "
	}

	result, err := processor.ManipulateWhitespace(text, modifier)

	require.NoError(t, err)
	assert.NotEqual(t, text, result)
	assert.Contains(t, result, "Line 1   ") // Original 2 spaces + 1 added
}

// TestManipulateWhitespace_EmptyText tests with empty text.
func TestManipulateWhitespace_EmptyText(t *testing.T) {
	processor := NewTextProcessor()

	modifier := func(line int, ws string) string {
		return ws
	}

	_, err := processor.ManipulateWhitespace("", modifier)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestManipulateWhitespace_NilModifier tests with nil modifier.
func TestManipulateWhitespace_NilModifier(t *testing.T) {
	processor := NewTextProcessor()

	_, err := processor.ManipulateWhitespace("Test", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

// TestGetTextInfo tests text info extraction.
func TestGetTextInfo(t *testing.T) {
	processor := NewTextProcessor()
	data := []byte("# Heading\n\nParagraph with **bold**.\n\n- Item 1\n- Item 2")

	info, err := processor.GetTextInfo(data)

	require.NoError(t, err)
	assert.Equal(t, "UTF-8", info.Encoding)
	assert.Equal(t, "LF", info.LineEnding)
	assert.True(t, info.IsMarkdown)
	assert.Greater(t, info.CharCount, 0)
	assert.Greater(t, info.LineCount, 0)
	assert.Greater(t, info.WordCount, 0)
}

// TestGetTextInfo_InvalidData tests info extraction with invalid data.
func TestGetTextInfo_InvalidData(t *testing.T) {
	processor := NewTextProcessor()
	data := []byte{0xFF, 0xFE}

	_, err := processor.GetTextInfo(data)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load text")
}

// TestLineEnding_String tests LineEnding string representation.
func TestLineEnding_String(t *testing.T) {
	tests := []struct {
		ending   LineEnding
		expected string
	}{
		{LineEndingLF, "LF"},
		{LineEndingCRLF, "CRLF"},
		{LineEndingCR, "CR"},
		{LineEndingUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ending.String())
		})
	}
}

// TestCountWords tests word counting.
func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"Simple sentence", "Hello world", 2},
		{"With punctuation", "Hello, world!", 2},
		{"Multiple spaces", "Hello    world", 2},
		{"With newlines", "Hello\nworld\n", 2},
		{"Empty string", "", 0},
		{"Only spaces", "   ", 0},
		{"Unicode", "Hello 世界", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTextProcessor()

			count := processor.countWords(tt.text)

			assert.Equal(t, tt.expected, count)
		})
	}
}
