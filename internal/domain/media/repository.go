// Package media repository interfaces.
package media

import "context"

// Repository defines the persistence interface for media assets.
type Repository interface {
	// SaveAsset persists a media asset.
	SaveAsset(ctx context.Context, asset *MediaAsset) error

	// GetAsset retrieves a media asset by ID.
	GetAsset(ctx context.Context, id AssetID) (*MediaAsset, error)

	// DeleteAsset removes a media asset.
	DeleteAsset(ctx context.Context, id AssetID) error

	// ListAssets returns a list of media assets filtered by type.
	ListAssets(ctx context.Context, mediaType MediaType, limit, offset int) ([]*MediaAsset, error)

	// SaveCapacityInfo persists capacity information.
	SaveCapacityInfo(ctx context.Context, info *CapacityInfo) error

	// GetCapacityInfo retrieves capacity information for an asset.
	GetCapacityInfo(ctx context.Context, assetID AssetID) (*CapacityInfo, error)
}
