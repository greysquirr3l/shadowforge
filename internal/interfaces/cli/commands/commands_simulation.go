// Package commands provides CLI command implementations for the Shadowforge steganography tool.
// This file contains simulation logic for demonstrating CLI functionality without full backend integration.
package commands

import (
	"fmt"

	"github.com/greysquirr3l/shadowforge/internal/application/commands"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// simulateCapacityAnalysis creates mock capacity analysis results for demonstration.
func simulateCapacityAnalysis(coverFile string, fileSize int64, techniques []stego.StegoTechnique) *commands.CapacityAnalysisResult {
	result := &commands.CapacityAnalysisResult{
		CoverFile:        coverFile,
		CoverSize:        fileSize,
		MediaType:        "image", // Simplified for simulation
		Format:           "png",   // Simplified for simulation
		TechniqueResults: make([]commands.TechniqueCapacityResult, 0, len(techniques)),
		Recommendations:  make([]commands.CapacityRecommendation, 0),
	}

	for _, tech := range techniques {
		techniqueResult := simulateTechniqueCapacity(tech, fileSize)
		result.TechniqueResults = append(result.TechniqueResults, techniqueResult)
	}

	// Add a best recommendation
	if len(result.TechniqueResults) > 0 {
		best := result.TechniqueResults[0]
		for _, tr := range result.TechniqueResults {
			score := float64(tr.SafeCapacity) * (1.0 - tr.DetectabilityRisk)
			bestScore := float64(best.SafeCapacity) * (1.0 - best.DetectabilityRisk)
			if score > bestScore {
				best = tr
			}
		}

		result.Recommendations = append(result.Recommendations, commands.CapacityRecommendation{
			Type:       "best",
			Technique:  best.Technique,
			Reason:     fmt.Sprintf("Optimal balance of capacity (%d bytes) and low detectability (%.1f%%)", best.SafeCapacity, best.DetectabilityRisk*100),
			MaxPayload: best.SafeCapacity,
			Confidence: 0.85,
		})
	}

	return result
}

// simulateTechniqueCapacity creates mock capacity metrics for a specific technique.
func simulateTechniqueCapacity(tech stego.StegoTechnique, fileSize int64) commands.TechniqueCapacityResult {
	var maxCapacity, safeCapacity int64
	var quality, detectability, performance float64

	switch tech {
	case stego.LSB:
		maxCapacity = fileSize / 8
		safeCapacity = maxCapacity / 4
		quality = 0.85
		detectability = 0.30
		performance = 0.95
	case stego.DCT:
		maxCapacity = fileSize / 16
		safeCapacity = maxCapacity / 3
		quality = 0.90
		detectability = 0.20
		performance = 0.70
	case stego.PhaseEncoding:
		maxCapacity = fileSize / 32
		safeCapacity = maxCapacity / 2
		quality = 0.92
		detectability = 0.10
		performance = 0.60
	case stego.EchoHiding:
		maxCapacity = fileSize / 64
		safeCapacity = maxCapacity / 2
		quality = 0.88
		detectability = 0.15
		performance = 0.65
	case stego.ZeroWidth:
		maxCapacity = fileSize / 4
		safeCapacity = maxCapacity / 8
		quality = 0.80
		detectability = 0.05
		performance = 0.98
	case stego.Palette:
		maxCapacity = fileSize / 64
		safeCapacity = maxCapacity / 4
		quality = 0.75
		detectability = 0.25
		performance = 0.85
	default:
		maxCapacity = fileSize / 10
		safeCapacity = maxCapacity / 5
		quality = 0.70
		detectability = 0.40
		performance = 0.80
	}

	return commands.TechniqueCapacityResult{
		Technique:         tech,
		MaxCapacity:       maxCapacity,
		SafeCapacity:      safeCapacity,
		QualityScore:      quality,
		DetectabilityRisk: detectability,
		PerformanceScore:  performance,
		Supported:         true,
		ErrorReason:       "",
	}
}
