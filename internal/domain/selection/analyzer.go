package selection

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// MediaAnalysis represents the analysis result for a single media file.
type MediaAnalysis struct {
	Path                 string
	Type                 media.MediaType
	Format               media.MediaFormat
	FileSize             int64
	RawCapacity          int64
	SafeCapacity         int64
	DetectabilityScore   float64
	RecommendedTechnique stego.StegoTechnique
	Dimensions           string // e.g., "1920x1080" for images, "3:45" for audio
}

// AnalysisResult represents the complete analysis of multiple media files.
type AnalysisResult struct {
	Files              []*MediaAnalysis
	TotalRawCapacity   int64
	TotalSafeCapacity  int64
	AverageDetectScore float64
	MediaTypeCounts    map[media.MediaType]int
}

// RecommendTechniqueForMedia determines the best steganography technique for a media type.
func RecommendTechniqueForMedia(mediaType media.MediaType, format media.MediaFormat) stego.StegoTechnique {
	switch mediaType {
	case media.MediaTypeImage:
		switch format {
		case media.FormatJPEG:
			return stego.DCT
		case media.FormatGIF:
			return stego.Palette
		default: // PNG, BMP
			return stego.LSB
		}
	case media.MediaTypeAudio:
		// Phase encoding is generally more robust
		return stego.PhaseEncoding
	case media.MediaTypeText:
		return stego.ZeroWidth
	default:
		return stego.LSB
	}
}

// CalculateDetectabilityScore estimates how detectable steganography would be.
// Lower scores are better (less detectable). Range: 0.0 to 1.0
func CalculateDetectabilityScore(capacity, usage int64, technique stego.StegoTechnique) float64 {
	if capacity == 0 {
		return 1.0 // Maximum risk if no capacity
	}

	utilizationRatio := float64(usage) / float64(capacity)

	// Base score from utilization (higher usage = more detectable)
	baseScore := utilizationRatio * 0.6

	// Technique-specific adjustments
	switch technique {
	case stego.LSB:
		baseScore += 0.2 // LSB is more detectable
	case stego.DCT:
		baseScore += 0.15 // DCT is fairly detectable
	case stego.PhaseEncoding, stego.EchoHiding:
		baseScore += 0.05 // Audio techniques are less detectable
	case stego.ZeroWidth:
		baseScore += 0.0 // Zero-width is very hard to detect
	case stego.Palette:
		baseScore += 0.1 // Palette manipulation is moderately detectable
	}

	// Cap at 1.0
	if baseScore > 1.0 {
		baseScore = 1.0
	}

	return baseScore
}

// FormatDimensions formats dimension information based on media type.
func FormatDimensions(mediaType media.MediaType, width, height int, duration float64) string {
	switch mediaType {
	case media.MediaTypeImage:
		return fmt.Sprintf("%dx%d", width, height)
	case media.MediaTypeAudio:
		minutes := int(duration / 60)
		seconds := int(duration) % 60
		return fmt.Sprintf("%d:%02d", minutes, seconds)
	default:
		return "N/A"
	}
}

// DetermineMediaType determines media type from file extension.
func DetermineMediaType(filePath string) media.MediaType {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".png", ".jpg", ".jpeg", ".bmp", ".gif":
		return media.MediaTypeImage
	case ".wav", ".flac", ".mp3":
		return media.MediaTypeAudio
	case ".txt", ".md":
		return media.MediaTypeText
	default:
		return "" // Empty string for unknown type
	}
}

// DetermineMediaFormat determines specific format from file extension.
func DetermineMediaFormat(filePath string) media.MediaFormat {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".png":
		return media.FormatPNG
	case ".jpg", ".jpeg":
		return media.FormatJPEG
	case ".bmp":
		return media.FormatBMP
	case ".gif":
		return media.FormatGIF
	case ".wav":
		return media.FormatWAV
	case ".flac":
		return media.FormatFLAC
	case ".mp3":
		return media.FormatMP3
	case ".txt":
		return media.FormatTXT
	case ".md":
		return media.FormatMarkdown
	default:
		return "" // Empty string for unknown format
	}
}
