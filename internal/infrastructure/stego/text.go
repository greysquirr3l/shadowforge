// Package stego provides steganography technique implementations.
//
// This file implements text steganography using zero-width characters
// to hide data in plain text documents.
package stego

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// Zero-width Unicode characters used for steganography.
const (
	ZWSP = '\u200B' // Zero Width Space - represents bit 0
	ZWJ  = '\u200D' // Zero Width Joiner - represents bit 1
	ZWNJ = '\u200C' // Zero Width Non-Joiner - alternative marker
	WJ   = '\u2060' // Word Joiner - segment separator
)

// TextEmbeddingConfig configures text steganography parameters.
type TextEmbeddingConfig struct {
	Method         TextMethod // Embedding method to use
	SpaceInterval  int        // Insert zero-width chars every N spaces (default: 1)
	WordInterval   int        // Insert between every N words (default: 1)
	UseWordJoiner  bool       // Use Word Joiner as segment separator
	PreserveFormat bool       // Preserve original text formatting
}

// TextMethod represents different text steganography methods.
type TextMethod string

const (
	MethodZeroWidth  TextMethod = "zero_width" // Zero-width character insertion
	MethodWhitespace TextMethod = "whitespace" // Space/tab manipulation
	MethodNewlines   TextMethod = "newlines"   // Line ending manipulation
)

// DefaultTextConfig returns default text steganography configuration.
func DefaultTextConfig() TextEmbeddingConfig {
	return TextEmbeddingConfig{
		Method:         MethodZeroWidth,
		SpaceInterval:  1,
		WordInterval:   1,
		UseWordJoiner:  false,
		PreserveFormat: true,
	}
}

// TextTechnique implements text steganography using zero-width characters.
//
// This technique embeds data by inserting invisible Unicode characters
// between words or at other strategic positions in text. The characters
// are invisible to the reader but can be detected programmatically.
type TextTechnique struct {
	config TextEmbeddingConfig
	logger *logrus.Logger
}

// NewTextTechnique creates a new text steganography technique.
func NewTextTechnique(config TextEmbeddingConfig, logger *logrus.Logger) *TextTechnique {
	if logger == nil {
		logger = logrus.New()
	}

	return &TextTechnique{
		config: config,
		logger: logger,
	}
}

// NewTextWithDefaults creates a text technique with default configuration.
func NewTextWithDefaults(logger *logrus.Logger) *TextTechnique {
	return NewTextTechnique(DefaultTextConfig(), logger)
}

// Name returns the technique identifier.
func (t *TextTechnique) Name() string {
	return "text_zw"
}

// SupportsFormat checks if this technique can be applied to the given format.
func (t *TextTechnique) SupportsFormat(format media.MediaFormat) bool {
	switch format {
	case media.FormatTXT, media.FormatMarkdown:
		return true
	default:
		return false
	}
}

// Embed hides payload data within text using zero-width characters.
func (t *TextTechnique) Embed(ctx context.Context, carrier, payload []byte) ([]byte, error) {
	t.logger.WithFields(logrus.Fields{
		"carrier_size": len(carrier),
		"payload_size": len(payload),
		"method":       string(t.config.Method),
	}).Info("Text embedding started")

	if len(payload) == 0 {
		return nil, stego.ErrEmptyPayload
	}

	if len(carrier) == 0 {
		return nil, stego.ErrEmptyCoverMedia
	}

	coverText := string(carrier)

	switch t.config.Method {
	case MethodZeroWidth:
		result, err := t.embedZeroWidth(coverText, payload)
		if err != nil {
			return nil, err
		}
		return []byte(result), nil

	case MethodWhitespace:
		result, err := t.embedWhitespace(coverText, payload)
		if err != nil {
			return nil, err
		}
		return []byte(result), nil

	case MethodNewlines:
		result, err := t.embedNewlines(coverText, payload)
		if err != nil {
			return nil, err
		}
		return []byte(result), nil

	default:
		return nil, stego.ErrInvalidTechnique
	}
}

// Extract retrieves hidden data from text by detecting zero-width characters.
func (t *TextTechnique) Extract(ctx context.Context, carrier []byte) ([]byte, error) {
	t.logger.WithFields(logrus.Fields{
		"carrier_size": len(carrier),
		"method":       string(t.config.Method),
	}).Info("Text extraction started")

	if len(carrier) == 0 {
		return nil, stego.ErrEmptyCoverMedia
	}

	stegoText := string(carrier)

	switch t.config.Method {
	case MethodZeroWidth:
		result, err := t.extractZeroWidth(stegoText)
		if err != nil {
			return nil, err
		}
		return result, nil

	case MethodWhitespace:
		result, err := t.extractWhitespace(stegoText)
		if err != nil {
			return nil, err
		}
		return result, nil

	case MethodNewlines:
		result, err := t.extractNewlines(stegoText)
		if err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, stego.ErrInvalidTechnique
	}
}

// CalculateCapacity determines embedding capacity for text steganography.
func (t *TextTechnique) CalculateCapacity(ctx context.Context, carrier []byte) (int, error) {
	if len(carrier) == 0 {
		return 0, stego.ErrInsufficientCapacity
	}

	coverText := string(carrier)

	switch t.config.Method {
	case MethodZeroWidth:
		return t.calculateZeroWidthCapacity(coverText), nil
	case MethodWhitespace:
		return t.calculateWhitespaceCapacity(coverText), nil
	case MethodNewlines:
		return t.calculateNewlineCapacity(coverText), nil
	default:
		return 0, stego.ErrInvalidTechnique
	}
}

// Zero-width character embedding methods

// embedZeroWidth embeds data using zero-width Unicode characters.
func (t *TextTechnique) embedZeroWidth(coverText string, payload []byte) (string, error) {
	bits := t.payloadToBits(payload)
	var result strings.Builder

	// Pre-allocate buffer size estimate
	result.Grow(len(coverText) + len(bits))

	bitIndex := 0
	words := strings.Fields(coverText)

	if len(words) == 0 {
		return "", stego.ErrEmptyCoverMedia
	}

	for i, word := range words {
		result.WriteString(word)

		// Insert zero-width characters between words
		if i < len(words)-1 && bitIndex < len(bits) {
			// Add space first
			result.WriteRune(' ')

			// Add zero-width characters based on bit values
			bitsToEmbed := min(8, len(bits)-bitIndex) // Embed up to 8 bits per gap

			for j := 0; j < bitsToEmbed; j++ {
				if bitIndex < len(bits) {
					if bits[bitIndex] == 0 {
						result.WriteRune(ZWSP) // Zero Width Space for 0
					} else {
						result.WriteRune(ZWJ) // Zero Width Joiner for 1
					}
					bitIndex++
				}
			}
		} else if i < len(words)-1 {
			// Just add regular space if no more bits to embed
			result.WriteRune(' ')
		}
	}

	if bitIndex < len(bits) {
		t.logger.WithFields(logrus.Fields{
			"embedded_bits": bitIndex,
			"total_bits":    len(bits),
		}).Warn("Not all payload bits could be embedded")

		// If we couldn't embed any bits, return an error
		if bitIndex == 0 {
			return "", stego.ErrInsufficientCapacity
		}
	}

	t.logger.WithFields(logrus.Fields{
		"embedded_bits": bitIndex,
		"result_length": result.Len(),
	}).Info("Zero-width embedding completed")

	return result.String(), nil
}

// extractZeroWidth extracts data from zero-width Unicode characters.
func (t *TextTechnique) extractZeroWidth(stegoText string) ([]byte, error) {
	var bits []byte

	for _, char := range stegoText {
		switch char {
		case ZWSP:
			bits = append(bits, 0)
		case ZWJ:
			bits = append(bits, 1)
		case WJ:
			// Word joiner can be used as segment separator - ignore
			continue
		}
	}

	if len(bits) == 0 {
		return nil, stego.ErrNoEmbeddedData
	}

	// Convert bits to bytes
	payload := t.bitsToPayload(bits)

	t.logger.WithFields(logrus.Fields{
		"extracted_bits": len(bits),
		"payload_bytes":  len(payload),
	}).Info("Zero-width extraction completed")

	return payload, nil
}

// calculateZeroWidthCapacity calculates capacity for zero-width embedding.
func (t *TextTechnique) calculateZeroWidthCapacity(text string) int {
	words := strings.Fields(text)
	if len(words) <= 1 {
		return 0
	}

	// Can embed in gaps between words
	gaps := len(words) - 1

	// Assume 8 bits per gap (conservative estimate)
	capacityBits := gaps * 8
	capacityBytes := capacityBits / 8

	return capacityBytes
}

// Whitespace manipulation methods

// embedWhitespace embeds data by manipulating spaces and tabs.
func (t *TextTechnique) embedWhitespace(coverText string, payload []byte) (string, error) {
	// TODO: Implement whitespace manipulation
	// For now, use a simple approach similar to zero-width

	bits := t.payloadToBits(payload)
	lines := strings.Split(coverText, "\n")
	var result strings.Builder

	bitIndex := 0
	for i, line := range lines {
		result.WriteString(line)

		// Add trailing spaces/tabs to encode bits
		if bitIndex < len(bits) && strings.TrimSpace(line) != "" {
			if bits[bitIndex] == 0 {
				result.WriteString(" ") // Single space for 0
			} else {
				result.WriteString("\t") // Tab for 1
			}
			bitIndex++
		}

		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// extractWhitespace extracts data from whitespace manipulation.
func (t *TextTechnique) extractWhitespace(stegoText string) ([]byte, error) {
	lines := strings.Split(stegoText, "\n")
	var bits []byte

	for _, line := range lines {
		if len(line) == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Check trailing whitespace
		trimmed := strings.TrimRightFunc(line, func(r rune) bool {
			return r == ' ' || r == '\t'
		})

		if len(line) > len(trimmed) {
			trailing := line[len(trimmed):]
			if trailing == " " {
				bits = append(bits, 0)
			} else if trailing == "\t" {
				bits = append(bits, 1)
			}
		}
	}

	if len(bits) == 0 {
		return nil, stego.ErrNoEmbeddedData
	}

	return t.bitsToPayload(bits), nil
}

// calculateWhitespaceCapacity calculates capacity for whitespace manipulation.
func (t *TextTechnique) calculateWhitespaceCapacity(text string) int {
	lines := strings.Split(text, "\n")
	nonEmptyLines := 0

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	// One bit per non-empty line
	return nonEmptyLines / 8 // Convert to bytes
}

// Line ending manipulation methods

// embedNewlines embeds data by manipulating line endings.
func (t *TextTechnique) embedNewlines(coverText string, payload []byte) (string, error) {
	// TODO: Implement line ending manipulation (CRLF vs LF)
	// For now, return original text
	return coverText, nil
}

// extractNewlines extracts data from line ending manipulation.
func (t *TextTechnique) extractNewlines(stegoText string) ([]byte, error) {
	// TODO: Implement line ending detection
	return nil, stego.ErrNoEmbeddedData
}

// calculateNewlineCapacity calculates capacity for newline manipulation.
func (t *TextTechnique) calculateNewlineCapacity(text string) int {
	lineCount := strings.Count(text, "\n")
	return lineCount / 8 // One bit per line ending
}

// Helper functions

// payloadToBits converts byte payload to individual bits.
func (t *TextTechnique) payloadToBits(payload []byte) []byte {
	bits := make([]byte, len(payload)*8)
	for i, b := range payload {
		for j := 0; j < 8; j++ {
			bit := (b >> (7 - j)) & 1
			bits[i*8+j] = bit
		}
	}
	return bits
}

// bitsToPayload converts individual bits back to byte payload.
func (t *TextTechnique) bitsToPayload(bits []byte) []byte {
	if len(bits) == 0 {
		return []byte{}
	}

	// Pad to byte boundary if needed
	for len(bits)%8 != 0 {
		bits = append(bits, 0)
	}

	payload := make([]byte, len(bits)/8)
	for i := 0; i < len(payload); i++ {
		var b byte
		for j := 0; j < 8; j++ {
			bit := bits[i*8+j]
			b |= (bit << (7 - j))
		}
		payload[i] = b
	}
	return payload
}

// containsZeroWidthChars checks if text contains zero-width characters.
func (t *TextTechnique) containsZeroWidthChars(text string) bool {
	for _, char := range text {
		switch char {
		case ZWSP, ZWJ, ZWNJ, WJ:
			return true
		}
	}
	return false
}

// cleanZeroWidthChars removes zero-width characters from text.
func (t *TextTechnique) cleanZeroWidthChars(text string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ZWSP, ZWJ, ZWNJ, WJ:
			return -1 // Remove character
		default:
			return r
		}
	}, text)
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Compile-time interface compliance check
var _ stego.Technique = (*TextTechnique)(nil)
