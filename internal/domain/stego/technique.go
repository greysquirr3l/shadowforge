// Package stego defines the steganography domain model and interfaces.
//
// This package is part of the Steganography bounded context, providing
// interfaces and types for embedding and extracting hidden data in various
// media formats using different steganographic techniques.
package stego

import (
	"context"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
)

// Technique defines the interface for steganographic embedding and extraction.
//
// All steganography techniques (LSB, DCT, Phase, Echo, etc.) must implement
// this interface to ensure consistent behavior across the system.
type Technique interface {
	// Embed hides payload data within the carrier media.
	// Returns the modified carrier with embedded data.
	Embed(ctx context.Context, carrier, payload []byte) ([]byte, error)

	// Extract retrieves hidden payload data from the carrier media.
	// Returns the extracted payload data.
	Extract(ctx context.Context, carrier []byte) ([]byte, error)

	// CalculateCapacity determines the maximum payload size that can be
	// embedded in the given carrier media.
	CalculateCapacity(ctx context.Context, carrier []byte) (int, error)

	// Name returns the technique identifier (e.g., "lsb", "dct", "phase").
	Name() string

	// SupportsFormat checks if this technique can be applied to the given media format.
	SupportsFormat(format media.MediaFormat) bool
}

// Embedder is the interface for embedding operations only.
// Use this for write-only steganography operations.
type Embedder interface {
	Embed(ctx context.Context, carrier, payload []byte) ([]byte, error)
	CalculateCapacity(ctx context.Context, carrier []byte) (int, error)
}

// Extractor is the interface for extraction operations only.
// Use this for read-only steganography operations.
type Extractor interface {
	Extract(ctx context.Context, carrier []byte) ([]byte, error)
}

// CapacityAnalyzer determines embedding capacity for media.
type CapacityAnalyzer interface {
	CalculateCapacity(ctx context.Context, carrier []byte) (int, error)
	SupportsFormat(format media.MediaFormat) bool
}
