// Package media provides infrastructure implementations for media processing.
package media

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
)

// ⚠️ SECURITY WARNING: TEST-ONLY IMPLEMENTATION
//
// MemoryRepository is designed EXCLUSIVELY for testing and development.
// ❌ DO NOT USE IN PRODUCTION ❌
//
// Security concerns:
//  1. NO ENCRYPTION: All data stored in plain text in RAM
//  2. NO PERSISTENCE: Data lost on restart (good for security, bad for production)
//  3. NO AUDIT TRAIL: No logging of access patterns
//  4. MEMORY LEAKS: Data not securely zeroed on deletion
//  5. PROCESS MEMORY DUMPS: Sensitive data visible in core dumps
//
// Production alternatives:
//  - CLI (Phase 5): Use direct file I/O with NO repository pattern
//  - API (Phase 6): Use EncryptedTempRepository with auto-shred (see docs/security/ephemeral_storage_architecture.md)
//
// This implementation exists ONLY to:
//  ✅ Enable fast unit/integration testing
//  ✅ Demonstrate the Repository interface pattern
//  ✅ Avoid database dependencies in tests
//
// See: docs/security/ephemeral_storage_architecture.md for production security architecture.

// MemoryRepository is a thread-safe in-memory implementation of media.Repository.
// Designed for testing and development environments.
// For production, use a persistent storage implementation (SQL, NoSQL, etc.).
type MemoryRepository struct {
	mu sync.RWMutex

	// assets stores MediaAsset objects keyed by AssetID
	assets map[string]*media.MediaAsset

	// capacityInfo stores CapacityInfo objects keyed by AssetID
	capacityInfo map[string]*media.CapacityInfo

	logger *slog.Logger
}

// NewMemoryRepository creates a new thread-safe in-memory repository.
func NewMemoryRepository(logger *slog.Logger) *MemoryRepository {
	if logger == nil {
		logger = slog.Default()
	}

	return &MemoryRepository{
		assets:       make(map[string]*media.MediaAsset),
		capacityInfo: make(map[string]*media.CapacityInfo),
		logger:       logger,
	}
}

// SaveAsset persists a media asset to memory.
// Thread-safe: uses write lock for modification.
func (r *MemoryRepository) SaveAsset(ctx context.Context, asset *media.MediaAsset) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return err
	}

	// Validate input
	if asset == nil {
		return fmt.Errorf("cannot save nil asset")
	}

	if asset.ID.IsZero() {
		return fmt.Errorf("cannot save asset with zero ID")
	}

	// Acquire write lock
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check context again after acquiring lock
	if err := ctx.Err(); err != nil {
		return err
	}

	assetID := asset.ID.String()

	// Check if asset exists (for logging purposes)
	_, exists := r.assets[assetID]

	// Store asset (creates or updates)
	r.assets[assetID] = asset

	if exists {
		r.logger.Debug("Updated asset in repository",
			slog.String("asset_id", assetID),
			slog.String("type", string(asset.Type)))
	} else {
		r.logger.Info("Saved new asset to repository",
			slog.String("asset_id", assetID),
			slog.String("type", string(asset.Type)),
			slog.String("format", string(asset.Format)))
	}

	return nil
}

// GetAsset retrieves a media asset by ID.
// Thread-safe: uses read lock for concurrent access.
func (r *MemoryRepository) GetAsset(ctx context.Context, id media.AssetID) (*media.MediaAsset, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Validate input
	if id.IsZero() {
		return nil, fmt.Errorf("cannot get asset with zero ID")
	}

	// Acquire read lock
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check context again after acquiring lock
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	assetID := id.String()

	// Retrieve asset
	asset, exists := r.assets[assetID]
	if !exists {
		return nil, media.ErrAssetNotFound
	}

	r.logger.Debug("Retrieved asset from repository",
		slog.String("asset_id", assetID),
		slog.String("type", string(asset.Type)))

	return asset, nil
}

// DeleteAsset removes a media asset from memory.
// Thread-safe: uses write lock for modification.
func (r *MemoryRepository) DeleteAsset(ctx context.Context, id media.AssetID) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return err
	}

	// Validate input
	if id.IsZero() {
		return fmt.Errorf("cannot delete asset with zero ID")
	}

	// Acquire write lock
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check context again after acquiring lock
	if err := ctx.Err(); err != nil {
		return err
	}

	assetID := id.String()

	// Check if asset exists
	asset, exists := r.assets[assetID]
	if !exists {
		return media.ErrAssetNotFound
	}

	// Delete asset and associated capacity info
	delete(r.assets, assetID)
	delete(r.capacityInfo, assetID)

	r.logger.Info("Deleted asset from repository",
		slog.String("asset_id", assetID),
		slog.String("type", string(asset.Type)))

	return nil
}

// ListAssets returns a list of media assets filtered by type.
// Thread-safe: uses read lock for concurrent access.
// Parameters:
//   - mediaType: Filter by media type (empty for all types)
//   - limit: Maximum number of results (0 for no limit)
//   - offset: Number of results to skip (0 for start)
func (r *MemoryRepository) ListAssets(ctx context.Context, mediaType media.MediaType, limit, offset int) ([]*media.MediaAsset, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Validate pagination parameters
	if limit < 0 {
		return nil, fmt.Errorf("limit cannot be negative: %d", limit)
	}
	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative: %d", offset)
	}

	// Acquire read lock
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check context again after acquiring lock
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Collect all matching assets
	var matches []*media.MediaAsset
	for _, asset := range r.assets {
		// Filter by media type if specified
		if mediaType != "" && asset.Type != mediaType {
			continue
		}
		matches = append(matches, asset)
	}

	// Apply pagination
	total := len(matches)

	// Handle offset
	if offset >= total {
		r.logger.Debug("ListAssets offset exceeds total",
			slog.Int("offset", offset),
			slog.Int("total", total))
		return []*media.MediaAsset{}, nil
	}

	// Apply offset
	matches = matches[offset:]

	// Apply limit
	if limit > 0 && limit < len(matches) {
		matches = matches[:limit]
	}

	r.logger.Debug("Listed assets from repository",
		slog.String("type_filter", string(mediaType)),
		slog.Int("total_matched", total),
		slog.Int("returned", len(matches)),
		slog.Int("limit", limit),
		slog.Int("offset", offset))

	return matches, nil
}

// SaveCapacityInfo persists capacity information.
// Thread-safe: uses write lock for modification.
func (r *MemoryRepository) SaveCapacityInfo(ctx context.Context, info *media.CapacityInfo) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return err
	}

	// Validate input
	if info == nil {
		return fmt.Errorf("cannot save nil capacity info")
	}

	if info.AssetID.IsZero() {
		return fmt.Errorf("cannot save capacity info with zero asset ID")
	}

	// Acquire write lock
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check context again after acquiring lock
	if err := ctx.Err(); err != nil {
		return err
	}

	assetID := info.AssetID.String()

	// Verify asset exists
	if _, exists := r.assets[assetID]; !exists {
		return media.ErrAssetNotFound
	}

	// Store capacity info
	r.capacityInfo[assetID] = info

	r.logger.Debug("Saved capacity info to repository",
		slog.String("asset_id", assetID),
		slog.Int64("total_capacity", info.TotalCapacity),
		slog.Int64("usable_capacity", info.UsableCapacity))

	return nil
}

// GetCapacityInfo retrieves capacity information for an asset.
// Thread-safe: uses read lock for concurrent access.
func (r *MemoryRepository) GetCapacityInfo(ctx context.Context, assetID media.AssetID) (*media.CapacityInfo, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Validate input
	if assetID.IsZero() {
		return nil, fmt.Errorf("cannot get capacity info with zero asset ID")
	}

	// Acquire read lock
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check context again after acquiring lock
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	id := assetID.String()

	// Verify asset exists
	if _, exists := r.assets[id]; !exists {
		return nil, media.ErrAssetNotFound
	}

	// Retrieve capacity info
	info, exists := r.capacityInfo[id]
	if !exists {
		return nil, media.ErrCapacityNotFound
	}

	r.logger.Debug("Retrieved capacity info from repository",
		slog.String("asset_id", id))

	return info, nil
}

// Clear removes all assets and capacity info from the repository.
// Useful for testing. Thread-safe: uses write lock.
func (r *MemoryRepository) Clear(ctx context.Context) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return err
	}

	// Acquire write lock
	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear all data
	r.assets = make(map[string]*media.MediaAsset)
	r.capacityInfo = make(map[string]*media.CapacityInfo)

	r.logger.Info("Cleared all data from repository")

	return nil
}

// Count returns the total number of assets in the repository.
// Thread-safe: uses read lock.
func (r *MemoryRepository) Count(ctx context.Context) (int, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	// Acquire read lock
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.assets), nil
}
