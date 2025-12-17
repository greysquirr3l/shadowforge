package commands

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// stringToTechnique converts a string to a StegoTechnique.
func stringToTechnique(technique string) (stego.StegoTechnique, error) {
	switch strings.ToLower(technique) {
	case "lsb":
		return stego.LSB, nil
	case "dct":
		return stego.DCT, nil
	case "phase":
		return stego.PhaseEncoding, nil
	case "echo":
		return stego.EchoHiding, nil
	case "zerowidth":
		return stego.ZeroWidth, nil
	case "palette":
		return stego.Palette, nil
	default:
		return "", fmt.Errorf("unsupported technique: %s", technique)
	}
}

// detectTechniqueFromFile detects the best steganography technique for a file.
func detectTechniqueFromFile(filename string) stego.StegoTechnique {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".png", ".bmp":
		return stego.LSB
	case ".jpg", ".jpeg":
		return stego.DCT
	case ".wav":
		return stego.PhaseEncoding // Default to phase for audio
	case ".txt", ".md":
		return stego.ZeroWidth
	case ".gif":
		return stego.Palette
	default:
		return stego.LSB // Default fallback
	}
}

// generateEmbedOutputFilename creates an innocuous output filename for embedded data.
// Uses subtle variations that are common in everyday file naming to avoid suspicion.
func generateEmbedOutputFilename(coverFile string) string {
	dir := filepath.Dir(coverFile)
	base := filepath.Base(coverFile)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Generate 4-byte random hex suffix for uniqueness
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomSuffix := hex.EncodeToString(randomBytes)

	// Use common, innocuous naming patterns
	patterns := []string{
		fmt.Sprintf("%s_edited%s", nameWithoutExt, ext),               // "photo_edited.png"
		fmt.Sprintf("%s_final%s", nameWithoutExt, ext),                // "document_final.jpg"
		fmt.Sprintf("%s_copy%s", nameWithoutExt, ext),                 // "image_copy.png"
		fmt.Sprintf("%s_%s%s", nameWithoutExt, randomSuffix[:6], ext), // "photo_a3f8d2.png"
	}

	// Use the random suffix pattern (most subtle and unique)
	outputFilename := patterns[3]

	// If original has directory, use it; otherwise use current directory
	if dir != "." && dir != "" {
		return filepath.Join(dir, outputFilename)
	}
	return outputFilename
}

// generateExtractOutputFilename creates a generic output filename for extracted data.
// Since we don't know the original file type, use a neutral name.
func generateExtractOutputFilename(stegoFile string) string {
	dir := filepath.Dir(stegoFile)

	// Generate random filename for extracted data
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomName := hex.EncodeToString(randomBytes)

	// Use generic names that don't suggest hidden data
	outputFilename := fmt.Sprintf("output_%s.bin", randomName[:6])

	// If original has directory, use it; otherwise use current directory
	if dir != "." && dir != "" {
		return filepath.Join(dir, outputFilename)
	}
	return outputFilename
}

// TechniqueScore represents the suitability of a technique for a given file.
type TechniqueScore struct {
	Technique     stego.StegoTechnique
	CapacityScore float64 // 0-100: higher = more capacity
	StealthScore  float64 // 0-100: higher = more stealthy/obfuscated
	CombinedScore float64 // Weighted combination
	Reason        string
}

// selectOptimalTechnique intelligently selects the best technique for a cover file
// considering both capacity and stealth/obfuscation characteristics.
func selectOptimalTechnique(coverFile string, payloadSize int64) (stego.StegoTechnique, error) {
	ext := strings.ToLower(filepath.Ext(coverFile))

	var candidates []TechniqueScore

	switch ext {
	case ".png", ".bmp":
		// PNG/BMP supports LSB (high capacity, medium stealth)
		candidates = append(candidates, TechniqueScore{
			Technique:     stego.LSB,
			CapacityScore: 90, // Very high capacity
			StealthScore:  65, // Medium stealth (detectable with analysis)
			Reason:        "LSB: High capacity, moderate stealth",
		})

	case ".jpg", ".jpeg":
		// JPEG supports DCT (medium capacity, high stealth)
		candidates = append(candidates, TechniqueScore{
			Technique:     stego.DCT,
			CapacityScore: 60, // Medium capacity
			StealthScore:  85, // High stealth (survives recompression)
			Reason:        "DCT: Robust to compression, high stealth",
		})

	case ".gif":
		// GIF supports Palette (low capacity, very high stealth)
		candidates = append(candidates, TechniqueScore{
			Technique:     stego.Palette,
			CapacityScore: 40, // Lower capacity
			StealthScore:  95, // Very high stealth (statistical analysis resistant)
			Reason:        "Palette: Extremely subtle, statistical stealth",
		})

	case ".wav":
		// WAV supports multiple techniques - choose based on payload size
		// Phase encoding: medium capacity, very high stealth
		candidates = append(candidates, TechniqueScore{
			Technique:     stego.PhaseEncoding,
			CapacityScore: 70,
			StealthScore:  90, // Very stealthy (perceptually transparent)
			Reason:        "Phase: Imperceptible audio changes, high stealth",
		})

		// Echo hiding: low capacity, high stealth
		candidates = append(candidates, TechniqueScore{
			Technique:     stego.EchoHiding,
			CapacityScore: 50,
			StealthScore:  85,
			Reason:        "Echo: Natural-sounding, moderate stealth",
		})

		// LSB audio: high capacity, lower stealth
		candidates = append(candidates, TechniqueScore{
			Technique:     stego.LSB, // Using LSB for audio
			CapacityScore: 85,
			StealthScore:  60, // More detectable in analysis
			Reason:        "LSB Audio: High capacity, moderate stealth",
		})

	case ".txt", ".md":
		// Text supports zero-width (very low capacity, maximum stealth)
		candidates = append(candidates, TechniqueScore{
			Technique:     stego.ZeroWidth,
			CapacityScore: 30,  // Very limited capacity
			StealthScore:  100, // Completely invisible
			Reason:        "Zero-Width: Invisible characters, maximum stealth",
		})

	default:
		return "", fmt.Errorf("unsupported file format for steganography: %s", ext)
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no suitable techniques found for file: %s", coverFile)
	}

	// Calculate combined scores with weights
	// For maximum obfuscation, prioritize stealth over capacity
	stealthWeight := 0.7  // 70% weight on stealth
	capacityWeight := 0.3 // 30% weight on capacity

	for i := range candidates {
		candidates[i].CombinedScore =
			(candidates[i].StealthScore * stealthWeight) +
				(candidates[i].CapacityScore * capacityWeight)
	}

	// Find the best technique
	bestIdx := 0
	bestScore := candidates[0].CombinedScore
	for i, candidate := range candidates {
		if candidate.CombinedScore > bestScore {
			bestScore = candidate.CombinedScore
			bestIdx = i
		}
	}

	return candidates[bestIdx].Technique, nil
}

// autoDetectEmbeddedTechnique attempts to detect which technique was used
// for embedding by examining the file type and metadata.
func autoDetectEmbeddedTechnique(stegoFile string) (stego.StegoTechnique, error) {
	ext := strings.ToLower(filepath.Ext(stegoFile))

	// For now, use file extension-based detection
	// In a production system, this would examine the file for embedded markers
	// or use statistical analysis to detect the technique
	switch ext {
	case ".png", ".bmp":
		return stego.LSB, nil
	case ".jpg", ".jpeg":
		return stego.DCT, nil
	case ".gif":
		return stego.Palette, nil
	case ".wav":
		// Default to Phase for WAV (most common/recommended)
		// TODO: Implement actual technique detection via markers or analysis
		return stego.PhaseEncoding, nil
	case ".txt", ".md":
		return stego.ZeroWidth, nil
	default:
		return "", fmt.Errorf("unable to auto-detect technique for file type: %s", ext)
	}
}

// getTechniqueScore returns the capacity and stealth scores for a given technique
func getTechniqueScore(technique stego.StegoTechnique) TechniqueScore {
	switch technique {
	case stego.LSB:
		return TechniqueScore{
			Technique:     stego.LSB,
			CapacityScore: 90,
			StealthScore:  65,
		}
	case stego.DCT:
		return TechniqueScore{
			Technique:     stego.DCT,
			CapacityScore: 60,
			StealthScore:  85,
		}
	case stego.Palette:
		return TechniqueScore{
			Technique:     stego.Palette,
			CapacityScore: 40,
			StealthScore:  95,
		}
	case stego.PhaseEncoding:
		return TechniqueScore{
			Technique:     stego.PhaseEncoding,
			CapacityScore: 70,
			StealthScore:  90,
		}
	case stego.EchoHiding:
		return TechniqueScore{
			Technique:     stego.EchoHiding,
			CapacityScore: 50,
			StealthScore:  85,
		}
	case stego.ZeroWidth:
		return TechniqueScore{
			Technique:     stego.ZeroWidth,
			CapacityScore: 30,
			StealthScore:  100,
		}
	default:
		// Default fallback
		return TechniqueScore{
			Technique:     technique,
			CapacityScore: 50,
			StealthScore:  50,
		}
	}
}
