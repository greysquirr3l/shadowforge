package commands

import (
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
