// Package media provides infrastructure implementations for media processing.
package media

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
)

// Zero-width character constants for steganography.
const (
	ZWSP = '\u200B' // Zero Width Space - represents 0
	ZWJ  = '\u200D' // Zero Width Joiner - represents 1
	ZWNJ = '\u200C' // Zero Width Non-Joiner - alternative bit representation
)

// TextProcessor handles text file operations for steganography.
type TextProcessor struct {
	detector *FormatDetector
}

// NewTextProcessor creates a new text processor.
func NewTextProcessor() *TextProcessor {
	return &TextProcessor{
		detector: NewFormatDetector(),
	}
}

// TextData represents decoded text with metadata.
type TextData struct {
	Content    string
	Encoding   string
	LineEnding LineEnding
	IsMarkdown bool
	CharCount  int
	LineCount  int
	WordCount  int
}

// LineEnding represents text file line ending style.
type LineEnding int

const (
	LineEndingUnknown LineEnding = iota
	LineEndingLF                 // Unix/Linux: \n
	LineEndingCRLF               // Windows: \r\n
	LineEndingCR                 // Old Mac: \r
)

func (le LineEnding) String() string {
	switch le {
	case LineEndingLF:
		return "LF"
	case LineEndingCRLF:
		return "CRLF"
	case LineEndingCR:
		return "CR"
	default:
		return "Unknown"
	}
}

// LoadText decodes text data and extracts metadata.
func (p *TextProcessor) LoadText(data []byte) (*TextData, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("text data is empty")
	}

	// Validate UTF-8 encoding
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("invalid UTF-8 encoding")
	}

	content := string(data)

	// Detect line ending style
	lineEnding := p.detectLineEnding(content)

	// Check if Markdown
	isMarkdown := hasMarkdownIndicators(data)

	// Count characters, lines, words
	charCount := utf8.RuneCountInString(content)
	lineCount := strings.Count(content, "\n") + 1
	if strings.HasSuffix(content, "\n") && len(content) > 1 {
		lineCount--
	}
	wordCount := p.countWords(content)

	return &TextData{
		Content:    content,
		Encoding:   "UTF-8",
		LineEnding: lineEnding,
		IsMarkdown: isMarkdown,
		CharCount:  charCount,
		LineCount:  lineCount,
		WordCount:  wordCount,
	}, nil
}

// SaveText encodes text data to bytes.
func (p *TextProcessor) SaveText(text *TextData) ([]byte, error) {
	if text == nil {
		return nil, fmt.Errorf("text data is nil")
	}

	if text.Content == "" {
		return nil, fmt.Errorf("text content is empty")
	}

	// Validate UTF-8
	if !utf8.ValidString(text.Content) {
		return nil, fmt.Errorf("invalid UTF-8 in text content")
	}

	return []byte(text.Content), nil
}

// detectLineEnding detects the line ending style used in text.
func (p *TextProcessor) detectLineEnding(text string) LineEnding {
	hasCRLF := strings.Contains(text, "\r\n")
	hasCR := strings.Contains(text, "\r")
	hasLF := strings.Contains(text, "\n")

	if hasCRLF {
		return LineEndingCRLF
	}
	if hasCR && !hasLF {
		return LineEndingCR
	}
	if hasLF {
		return LineEndingLF
	}

	return LineEndingUnknown
}

// countWords counts words in text (space-separated).
func (p *TextProcessor) countWords(text string) int {
	inWord := false
	wordCount := 0

	for _, r := range text {
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			inWord = true
			wordCount++
		}
	}

	return wordCount
}

// NormalizeLineEndings converts line endings to specified style.
func (p *TextProcessor) NormalizeLineEndings(text string, target LineEnding) (string, error) {
	if target == LineEndingUnknown {
		return "", fmt.Errorf("cannot normalize to unknown line ending")
	}

	// First normalize to LF
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	// Then convert to target
	switch target {
	case LineEndingLF:
		return normalized, nil
	case LineEndingCRLF:
		return strings.ReplaceAll(normalized, "\n", "\r\n"), nil
	case LineEndingCR:
		return strings.ReplaceAll(normalized, "\n", "\r"), nil
	default:
		return "", fmt.Errorf("unsupported line ending: %v", target)
	}
}

// CalculateCapacity estimates steganographic capacity for text.
func (p *TextProcessor) CalculateCapacity(data []byte, technique string) (*media.CapacityInfo, error) {
	// Load text to validate and get metadata
	textData, err := p.LoadText(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load text: %w", err)
	}

	var rawCapacity, safeCapacity int64

	switch technique {
	case "zero-width":
		// Zero-width character embedding: 1 bit per word boundary
		// Can insert ZWSP (0) or ZWJ (1) between words
		rawCapacity = int64(textData.WordCount) / 8 // Convert bits to bytes

		// Safe capacity: Use only 60% of word boundaries to avoid detection
		safeCapacity = int64(float64(rawCapacity) * 0.60)

	case "whitespace":
		// Whitespace manipulation: Encode in spaces/tabs
		// Estimate: 1 bit per line (space vs double-space at line end)
		rawCapacity = int64(textData.LineCount) / 8

		// Safe capacity: Use 50% of lines
		safeCapacity = int64(float64(rawCapacity) * 0.50)

	case "line-ending":
		// Line ending encoding: Alternate LF vs CRLF
		// 1 bit per line
		rawCapacity = int64(textData.LineCount) / 8

		// Safe capacity: Use 40% of lines
		safeCapacity = int64(float64(rawCapacity) * 0.40)

	case "homoglyph":
		// Homoglyph substitution: Replace characters with similar-looking Unicode
		// Very limited capacity, depends on character availability
		// Estimate: 10% of characters can be substituted
		substitutableChars := int64(float64(textData.CharCount) * 0.10)
		rawCapacity = substitutableChars / 8

		// Safe capacity: Use only 30% of substitutable chars
		safeCapacity = int64(float64(rawCapacity) * 0.30)

	default:
		return nil, fmt.Errorf("unknown technique: %s", technique)
	}

	// Calculate quality score
	qualityScore := p.calculateQualityScore(textData)

	return &media.CapacityInfo{
		AssetID:        media.AssetID{}, // Empty, to be set by caller
		MediaType:      media.MediaTypeText,
		TotalCapacity:  rawCapacity,
		UsableCapacity: safeCapacity,
		RecommendedMax: safeCapacity,
		QualityImpact:  1.0 - qualityScore,
	}, nil
}

// calculateQualityScore assesses text quality for steganography (0.0 - 1.0).
func (p *TextProcessor) calculateQualityScore(text *TextData) float64 {
	score := 0.5 // Base score

	// Longer text is better
	if text.CharCount > 1000 {
		score += 0.2
	} else if text.CharCount > 500 {
		score += 0.1
	}

	// More words provide more embedding points
	if text.WordCount > 200 {
		score += 0.15
	} else if text.WordCount > 100 {
		score += 0.08
	}

	// More lines provide more opportunities
	if text.LineCount > 50 {
		score += 0.1
	} else if text.LineCount > 20 {
		score += 0.05
	}

	// Markdown provides more structure for hiding data
	if text.IsMarkdown {
		score += 0.05
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// ValidateText validates text data.
func (p *TextProcessor) ValidateText(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("text data is empty")
	}

	// Validate UTF-8
	if !utf8.Valid(data) {
		return fmt.Errorf("invalid UTF-8 encoding")
	}

	// Additional validation: Check for null bytes (binary data indicator)
	if bytes.Contains(data, []byte{0}) {
		return fmt.Errorf("text contains null bytes (possibly binary data)")
	}

	return nil
}

// EmbedZeroWidth embeds data using zero-width characters.
func (p *TextProcessor) EmbedZeroWidth(coverText string, data []byte) (string, error) {
	if coverText == "" {
		return "", fmt.Errorf("cover text is empty")
	}

	if len(data) == 0 {
		return "", fmt.Errorf("data to embed is empty")
	}

	var result strings.Builder
	result.Grow(len(coverText) + len(data)*8*3) // Estimate size

	bitIndex := 0
	totalBits := len(data) * 8

	// Iterate through cover text
	for _, char := range coverText {
		result.WriteRune(char)

		// Insert zero-width character after spaces (word boundaries)
		if unicode.IsSpace(char) && bitIndex < totalBits {
			// Extract bit
			byteIndex := bitIndex / 8
			bitOffset := 7 - (bitIndex % 8)
			bit := (data[byteIndex] >> bitOffset) & 1

			// Embed bit using zero-width character
			if bit == 0 {
				result.WriteRune(ZWSP) // 0 -> Zero Width Space
			} else {
				result.WriteRune(ZWJ) // 1 -> Zero Width Joiner
			}

			bitIndex++
		}
	}

	// Check if we embedded all data
	if bitIndex < totalBits {
		return "", fmt.Errorf("insufficient capacity: embedded %d of %d bits", bitIndex, totalBits)
	}

	return result.String(), nil
}

// ExtractZeroWidth extracts data from zero-width characters.
func (p *TextProcessor) ExtractZeroWidth(stegoText string) ([]byte, error) {
	if stegoText == "" {
		return nil, fmt.Errorf("stego text is empty")
	}

	var bits []byte

	// Extract zero-width characters
	for _, char := range stegoText {
		switch char {
		case ZWSP:
			bits = append(bits, 0)
		case ZWJ:
			bits = append(bits, 1)
		}
		// Ignore other characters (including ZWNJ for now)
	}

	if len(bits) == 0 {
		return nil, fmt.Errorf("no zero-width characters found")
	}

	// Convert bits to bytes
	numBytes := len(bits) / 8
	if len(bits)%8 != 0 {
		numBytes++
	}

	data := make([]byte, numBytes)
	for i, bit := range bits {
		if bit == 1 {
			byteIndex := i / 8
			bitOffset := 7 - (i % 8)
			data[byteIndex] |= 1 << bitOffset
		}
	}

	// Trim padding from last byte if needed
	if len(bits)%8 != 0 {
		data = data[:len(bits)/8]
	}

	return data, nil
}

// ManipulateWhitespace applies a modifier function to whitespace in text.
func (p *TextProcessor) ManipulateWhitespace(text string, modifier func(line int, ws string) string) (string, error) {
	if text == "" {
		return "", fmt.Errorf("text is empty")
	}

	if modifier == nil {
		return "", fmt.Errorf("modifier function is nil")
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Extract trailing whitespace
		trimmed := strings.TrimRight(line, " \t")
		ws := line[len(trimmed):]

		// Apply modifier
		newWS := modifier(i, ws)

		// Rebuild line
		lines[i] = trimmed + newWS
	}

	return strings.Join(lines, "\n"), nil
}

// GetTextInfo extracts text metadata.
func (p *TextProcessor) GetTextInfo(data []byte) (*TextInfo, error) {
	textData, err := p.LoadText(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load text: %w", err)
	}

	return &TextInfo{
		Encoding:   textData.Encoding,
		LineEnding: textData.LineEnding.String(),
		IsMarkdown: textData.IsMarkdown,
		CharCount:  textData.CharCount,
		LineCount:  textData.LineCount,
		WordCount:  textData.WordCount,
	}, nil
}

// TextInfo contains metadata about text.
type TextInfo struct {
	Encoding   string
	LineEnding string
	IsMarkdown bool
	CharCount  int
	LineCount  int
	WordCount  int
}
