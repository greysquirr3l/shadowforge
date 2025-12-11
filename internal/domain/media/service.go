// Package media service interfaces.
package media

import "context"

// Service defines the domain service interface for media processing operations.
type Service interface {
	// LoadMedia loads media data and creates a MediaAsset.
	LoadMedia(ctx context.Context, data []byte, format MediaFormat) (*MediaAsset, error)

	// DetectFormat automatically detects the media format from data.
	DetectFormat(ctx context.Context, data []byte) (MediaFormat, error)

	// CalculateCapacity calculates the embedding capacity for a media asset.
	// Note: This will need stego.StegoTechnique when steganography domain exists.
	CalculateCapacity(ctx context.Context, asset *MediaAsset) (*CapacityInfo, error)

	// SanitizeMetadata removes all metadata from a media asset.
	SanitizeMetadata(ctx context.Context, asset *MediaAsset) error

	// ValidateMedia validates the integrity of a media asset.
	ValidateMedia(ctx context.Context, asset *MediaAsset) error

	// AnalyzeQuality analyzes the quality score of a media asset.
	AnalyzeQuality(ctx context.Context, asset *MediaAsset) (float64, error)
}
