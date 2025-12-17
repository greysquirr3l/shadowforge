package selection

import (
	"fmt"
	"strings"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// SuggestionEngine provides intelligent capacity expansion suggestions.
type SuggestionEngine struct{}

// NewSuggestionEngine creates a new suggestion engine.
func NewSuggestionEngine() *SuggestionEngine {
	return &SuggestionEngine{}
}

// GenerateSuggestions generates actionable suggestions for bridging capacity gap.
func (e *SuggestionEngine) GenerateSuggestions(
	requiredCapacity int64,
	availableCapacity int64,
) *CapacitySuggestions {
	gap := requiredCapacity - availableCapacity

	suggestions := &CapacitySuggestions{
		RequiredCapacity:  requiredCapacity,
		AvailableCapacity: availableCapacity,
		CapacityGap:       gap,
		MediaSuggestions:  make([]*MediaSuggestion, 0),
		Alternatives:      make([]string, 0),
	}

	// Generate media-specific suggestions
	suggestions.MediaSuggestions = append(suggestions.MediaSuggestions,
		e.suggestImages(gap),
		e.suggestAudio(gap),
		e.suggestText(gap),
	)

	// Generate alternative strategies
	suggestions.Alternatives = append(suggestions.Alternatives,
		"Reduce Reed-Solomon redundancy (⚠️ lowers fault tolerance)",
		"Split payload into multiple separate operations",
		"Use higher-capacity techniques where available (e.g., Phase Encoding for audio)",
	)

	return suggestions
}

func (e *SuggestionEngine) suggestImages(gap int64) *MediaSuggestion {
	// Assume average LSB capacity: ~0.125 bytes/pixel (1 bit/pixel)
	// For 1920x1080 image: ~2,073,600 pixels × 0.125 = ~259KB capacity
	bytesPerPixel := 0.125
	pngDimensions := "1920x1080"
	pngPixels := 1920 * 1080
	pngCapacity := int64(float64(pngPixels) * bytesPerPixel)

	filesNeeded := (gap + pngCapacity - 1) / pngCapacity // Ceiling division

	return &MediaSuggestion{
		MediaType: media.MediaTypeImage,
		Message: fmt.Sprintf("Add %d PNG file(s) at %s (~%s capacity each)",
			filesNeeded, pngDimensions, formatBytes(pngCapacity)),
		ExpectedCapacity: pngCapacity * filesNeeded,
		Examples: []string{
			"Stock photo sites: unsplash.com, pexels.com, pixabay.com",
			"Screenshot tools: native OS screenshot (⌘+Shift+4 on macOS)",
		},
	}
}

func (e *SuggestionEngine) suggestAudio(gap int64) *MediaSuggestion {
	// Assume average LSB-Audio capacity: ~5.5KB/second for 44.1kHz 16-bit stereo
	// For 3-minute audio: ~180 seconds × 5.5KB = ~990KB capacity
	bytesPerSecond := 5500
	duration := 180 // 3 minutes
	audioCapacity := int64(bytesPerSecond * duration)

	filesNeeded := (gap + audioCapacity - 1) / audioCapacity

	return &MediaSuggestion{
		MediaType: media.MediaTypeAudio,
		Message: fmt.Sprintf("Add %d WAV file(s), %d+ minutes each (~%s capacity)",
			filesNeeded, duration/60, formatBytes(audioCapacity)),
		ExpectedCapacity: audioCapacity * filesNeeded,
		Examples: []string{
			"Free audio: freesound.org, zapsplat.com",
			"Generate silence: sox -n -r 44100 -c 2 output.wav trim 0.0 180.0",
		},
	}
}

func (e *SuggestionEngine) suggestText(gap int64) *MediaSuggestion {
	// Assume average zero-width capacity: ~0.05 bytes/character
	// For 10KB text file: ~10,000 chars × 0.05 = ~500 bytes capacity
	bytesPerChar := 0.05
	textSize := 10000 // 10KB file
	textCapacity := int64(float64(textSize) * bytesPerChar)

	filesNeeded := (gap + textCapacity - 1) / textCapacity

	return &MediaSuggestion{
		MediaType: media.MediaTypeText,
		Message: fmt.Sprintf("Add %d text file(s), %dKB+ each (~%s capacity)",
			filesNeeded, textSize/1024, formatBytes(textCapacity)),
		ExpectedCapacity: textCapacity * filesNeeded,
		Examples: []string{
			"Lorem ipsum generators: lipsum.com",
			"Project documentation files (README.md, CHANGELOG.md)",
		},
	}
}

// FormatSources formats suggestion sources for display.
func (e *SuggestionEngine) FormatSources(suggestion *MediaSuggestion) string {
	if len(suggestion.Examples) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("Sources:\n")
	for _, example := range suggestion.Examples {
		builder.WriteString(fmt.Sprintf("  • %s\n", example))
	}

	return builder.String()
}

// FormatQualityTradeoff provides guidance on quality vs detectability.
func (e *SuggestionEngine) FormatQualityTradeoff(technique stego.StegoTechnique) string {
	switch technique {
	case stego.LSB:
		return "LSB: Higher quality images → more capacity. Use lossless formats (PNG, BMP)."
	case stego.DCT:
		return "DCT: Higher JPEG quality → more usable coefficients. Aim for quality 90+."
	case stego.PhaseEncoding, stego.EchoHiding:
		return "Audio: Longer duration → more capacity. 16-bit 44.1kHz stereo recommended."
	case stego.ZeroWidth:
		return "Text: More characters → more insertion points. Natural language text works best."
	case stego.Palette:
		return "Palette: More colors → more capacity, but palette reordering detectable with statistical analysis."
	default:
		return "Balance capacity needs with acceptable detectability risk."
	}
}

// CapacitySuggestions contains suggestions for expanding capacity.
type CapacitySuggestions struct {
	RequiredCapacity  int64
	AvailableCapacity int64
	CapacityGap       int64
	MediaSuggestions  []*MediaSuggestion
	Alternatives      []string
}

// MediaSuggestion represents a suggestion for a specific media type.
type MediaSuggestion struct {
	MediaType        media.MediaType
	Message          string
	ExpectedCapacity int64
	Examples         []string
}

// CalculateCapacityPerPixel returns typical capacity per pixel for LSB.
func CalculateCapacityPerPixel(technique stego.StegoTechnique) float64 {
	switch technique {
	case stego.LSB:
		return 0.125 // 1 bit per pixel for single-channel LSB
	case stego.DCT:
		return 0.072 // ~7.2% capacity for DCT (varies by quality)
	case stego.Palette:
		return 0.145 // ~14.5% for palette-based techniques
	default:
		return 0.10 // Conservative default
	}
}

// CalculateCapacityPerSample returns typical capacity per audio sample.
func CalculateCapacityPerSample(technique stego.StegoTechnique) float64 {
	switch technique {
	case stego.PhaseEncoding:
		// Phase encoding uses FFT segments
		return 0.001 // ~0.1% capacity (conservative)
	case stego.EchoHiding:
		// Echo hiding capacity depends on delay settings
		return 0.0005 // ~0.05% capacity
	default:
		// LSB-Audio: 1 bit per sample per channel
		return 0.125 // 1 bit / 8 bits = 0.125 bytes per sample
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
