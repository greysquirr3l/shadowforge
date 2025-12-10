// Package stego implements the Steganography bounded context.
package stego

import "context"

// StegoService defines the domain service interface for steganographic operations.
// Implementations coordinate between domain aggregates and repositories.
type StegoService interface {
	// Embed embeds data into cover media using the specified technique.
	Embed(ctx context.Context, coverMedia, payload []byte, technique StegoTechnique) (*StegoContainer, error)

	// Extract extracts hidden data from steganographic media.
	Extract(ctx context.Context, stegoMedia []byte, technique StegoTechnique) ([]byte, error)

	// CalculateCapacity calculates the maximum embedding capacity for given media.
	CalculateCapacity(ctx context.Context, coverMedia []byte, technique StegoTechnique) (int64, error)

	// AnalyzeQuality analyzes the quality and detectability of an embedding.
	AnalyzeQuality(ctx context.Context, stegoMedia []byte) (*Quality, error)

	// ValidateContainer validates a steganographic container for correctness.
	ValidateContainer(ctx context.Context, container *StegoContainer) error

	// OptimizeTechnique selects the best technique for given media and requirements.
	OptimizeTechnique(ctx context.Context, coverMedia []byte, payloadSize int64) (StegoTechnique, error)
}
