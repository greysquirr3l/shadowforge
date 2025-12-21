package media

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
)

// TestNewMemoryRepository tests repository creation.
func TestNewMemoryRepository(t *testing.T) {
	t.Run("WithLogger", func(t *testing.T) {
		logger := slog.Default()
		repo := NewMemoryRepository(logger)

		assert.NotNil(t, repo)
		assert.NotNil(t, repo.assets)
		assert.NotNil(t, repo.capacityInfo)
		assert.Equal(t, logger, repo.logger)
	})

	t.Run("WithNilLogger", func(t *testing.T) {
		repo := NewMemoryRepository(nil)

		assert.NotNil(t, repo)
		assert.NotNil(t, repo.logger) // Should use default
	})
}

// TestSaveAsset tests asset saving functionality.
func TestSaveAsset(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository(slog.Default())

	t.Run("Success", func(t *testing.T) {
		asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)

		err := repo.SaveAsset(ctx, asset)
		require.NoError(t, err)

		// Verify saved
		retrieved, err := repo.GetAsset(ctx, asset.ID)
		require.NoError(t, err)
		assert.Equal(t, asset.ID, retrieved.ID)
	})

	t.Run("Update", func(t *testing.T) {
		asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)

		// Save initially
		err := repo.SaveAsset(ctx, asset)
		require.NoError(t, err)

		// Update
		asset.ModifiedAt = time.Now()
		err = repo.SaveAsset(ctx, asset)
		require.NoError(t, err)

		// Verify updated
		retrieved, err := repo.GetAsset(ctx, asset.ID)
		require.NoError(t, err)
		assert.Equal(t, asset.ModifiedAt, retrieved.ModifiedAt)
	})

	t.Run("NilAsset", func(t *testing.T) {
		err := repo.SaveAsset(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil asset")
	})

	t.Run("ZeroID", func(t *testing.T) {
		asset := &media.MediaAsset{} // Zero ID
		err := repo.SaveAsset(ctx, asset)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "zero ID")
	})

	t.Run("CancelledContext", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
		err := repo.SaveAsset(cancelledCtx, asset)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestGetAsset tests asset retrieval functionality.
func TestGetAsset(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository(slog.Default())

	t.Run("Success", func(t *testing.T) {
		asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)

		err := repo.SaveAsset(ctx, asset)
		require.NoError(t, err)

		retrieved, err := repo.GetAsset(ctx, asset.ID)
		require.NoError(t, err)
		assert.Equal(t, asset.ID, retrieved.ID)
		assert.Equal(t, asset.Type, retrieved.Type)
		assert.Equal(t, asset.Format, retrieved.Format)
	})

	t.Run("NotFound", func(t *testing.T) {
		id := media.NewAssetID()

		asset, err := repo.GetAsset(ctx, id)
		assert.Nil(t, asset)
		assert.ErrorIs(t, err, media.ErrAssetNotFound)
	})

	t.Run("ZeroID", func(t *testing.T) {
		asset, err := repo.GetAsset(ctx, media.AssetID{})
		assert.Nil(t, asset)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "zero ID")
	})

	t.Run("CancelledContext", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		id := media.NewAssetID()
		asset, err := repo.GetAsset(cancelledCtx, id)
		assert.Nil(t, asset)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestDeleteAsset tests asset deletion functionality.
func TestDeleteAsset(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository(slog.Default())

	t.Run("Success", func(t *testing.T) {
		asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)

		// Save first
		err := repo.SaveAsset(ctx, asset)
		require.NoError(t, err)

		// Delete
		err = repo.DeleteAsset(ctx, asset.ID)
		require.NoError(t, err)

		// Verify deleted
		retrieved, err := repo.GetAsset(ctx, asset.ID)
		assert.Nil(t, retrieved)
		assert.ErrorIs(t, err, media.ErrAssetNotFound)
	})

	t.Run("NotFound", func(t *testing.T) {
		id := media.NewAssetID()

		err := repo.DeleteAsset(ctx, id)
		assert.ErrorIs(t, err, media.ErrAssetNotFound)
	})

	t.Run("DeletesCapacityInfo", func(t *testing.T) {
		asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)

		// Save asset and capacity info
		err := repo.SaveAsset(ctx, asset)
		require.NoError(t, err)

		info := createTestCapacityInfo(asset.ID)
		err = repo.SaveCapacityInfo(ctx, info)
		require.NoError(t, err)

		// Delete asset
		err = repo.DeleteAsset(ctx, asset.ID)
		require.NoError(t, err)

		// Verify capacity info also deleted
		retrieved, err := repo.GetCapacityInfo(ctx, asset.ID)
		assert.Nil(t, retrieved)
		assert.ErrorIs(t, err, media.ErrAssetNotFound)
	})

	t.Run("ZeroID", func(t *testing.T) {
		err := repo.DeleteAsset(ctx, media.AssetID{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "zero ID")
	})

	t.Run("CancelledContext", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		id := media.NewAssetID()
		err := repo.DeleteAsset(cancelledCtx, id)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestListAssets tests asset listing functionality.
func TestListAssets(t *testing.T) {
	ctx := context.Background()

	t.Run("AllAssets", func(t *testing.T) {
		repo := NewMemoryRepository(slog.Default())

		// Create test assets
		assets := []*media.MediaAsset{
			createTestAsset(t, media.MediaTypeImage, media.FormatPNG),
			createTestAsset(t, media.MediaTypeImage, media.FormatJPEG),
			createTestAsset(t, media.MediaTypeAudio, media.FormatWAV),
		}

		// Save all
		for _, asset := range assets {
			err := repo.SaveAsset(ctx, asset)
			require.NoError(t, err)
		}

		// List all (no filter)
		results, err := repo.ListAssets(ctx, "", 0, 0)
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("FilterByType", func(t *testing.T) {
		repo := NewMemoryRepository(slog.Default())

		// Create mixed types
		imageAsset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
		audioAsset := createTestAsset(t, media.MediaTypeAudio, media.FormatWAV)
		textAsset := createTestAsset(t, media.MediaTypeText, media.FormatTXT)

		require.NoError(t, repo.SaveAsset(ctx, imageAsset))
		require.NoError(t, repo.SaveAsset(ctx, audioAsset))
		require.NoError(t, repo.SaveAsset(ctx, textAsset))

		// Filter by image
		results, err := repo.ListAssets(ctx, media.MediaTypeImage, 0, 0)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, media.MediaTypeImage, results[0].Type)

		// Filter by audio
		results, err = repo.ListAssets(ctx, media.MediaTypeAudio, 0, 0)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, media.MediaTypeAudio, results[0].Type)
	})

	t.Run("Pagination", func(t *testing.T) {
		repo := NewMemoryRepository(slog.Default())

		// Create 10 assets
		for i := 0; i < 10; i++ {
			asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
			require.NoError(t, repo.SaveAsset(ctx, asset))
		}

		// First page (limit 3)
		results, err := repo.ListAssets(ctx, "", 3, 0)
		require.NoError(t, err)
		assert.Len(t, results, 3)

		// Second page (offset 3, limit 3)
		results, err = repo.ListAssets(ctx, "", 3, 3)
		require.NoError(t, err)
		assert.Len(t, results, 3)

		// Third page (offset 6, limit 3)
		results, err = repo.ListAssets(ctx, "", 3, 6)
		require.NoError(t, err)
		assert.Len(t, results, 3)

		// Fourth page (offset 9, limit 3) - only 1 result
		results, err = repo.ListAssets(ctx, "", 3, 9)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})

	t.Run("OffsetExceedsTotal", func(t *testing.T) {
		repo := NewMemoryRepository(slog.Default())

		asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
		require.NoError(t, repo.SaveAsset(ctx, asset))

		// Offset beyond available
		results, err := repo.ListAssets(ctx, "", 10, 10)
		require.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("EmptyRepository", func(t *testing.T) {
		repo := NewMemoryRepository(slog.Default())

		results, err := repo.ListAssets(ctx, "", 0, 0)
		require.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("NegativeLimit", func(t *testing.T) {
		repo := NewMemoryRepository(slog.Default())

		results, err := repo.ListAssets(ctx, "", -1, 0)
		assert.Nil(t, results)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limit cannot be negative")
	})

	t.Run("NegativeOffset", func(t *testing.T) {
		repo := NewMemoryRepository(slog.Default())

		results, err := repo.ListAssets(ctx, "", 10, -1)
		assert.Nil(t, results)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "offset cannot be negative")
	})

	t.Run("CancelledContext", func(t *testing.T) {
		repo := NewMemoryRepository(slog.Default())
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		results, err := repo.ListAssets(cancelledCtx, "", 0, 0)
		assert.Nil(t, results)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestSaveCapacityInfo tests capacity info saving.
func TestSaveCapacityInfo(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository(slog.Default())

	t.Run("Success", func(t *testing.T) {
		asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
		require.NoError(t, repo.SaveAsset(ctx, asset))

		info := createTestCapacityInfo(asset.ID)
		err := repo.SaveCapacityInfo(ctx, info)
		require.NoError(t, err)

		// Verify saved
		retrieved, err := repo.GetCapacityInfo(ctx, asset.ID)
		require.NoError(t, err)
		assert.Equal(t, info.AssetID, retrieved.AssetID)
		assert.Equal(t, info.TotalCapacity, retrieved.TotalCapacity)
	})

	t.Run("NilInfo", func(t *testing.T) {
		err := repo.SaveCapacityInfo(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil capacity info")
	})

	t.Run("ZeroAssetID", func(t *testing.T) {
		info := &media.CapacityInfo{} // Zero AssetID
		err := repo.SaveCapacityInfo(ctx, info)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "zero asset ID")
	})

	t.Run("AssetNotFound", func(t *testing.T) {
		id := media.NewAssetID() // Asset doesn't exist
		info := createTestCapacityInfo(id)

		err := repo.SaveCapacityInfo(ctx, info)
		assert.ErrorIs(t, err, media.ErrAssetNotFound)
	})

	t.Run("CancelledContext", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		info := createTestCapacityInfo(media.NewAssetID())
		err := repo.SaveCapacityInfo(cancelledCtx, info)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestGetCapacityInfo tests capacity info retrieval.
func TestGetCapacityInfo(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository(slog.Default())

	t.Run("Success", func(t *testing.T) {
		asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
		require.NoError(t, repo.SaveAsset(ctx, asset))

		info := createTestCapacityInfo(asset.ID)
		require.NoError(t, repo.SaveCapacityInfo(ctx, info))

		// Retrieve
		retrieved, err := repo.GetCapacityInfo(ctx, asset.ID)
		require.NoError(t, err)
		assert.Equal(t, info.TotalCapacity, retrieved.TotalCapacity)
		assert.Equal(t, info.UsableCapacity, retrieved.UsableCapacity)
	})

	t.Run("AssetNotFound", func(t *testing.T) {
		id := media.NewAssetID()

		info, err := repo.GetCapacityInfo(ctx, id)
		assert.Nil(t, info)
		assert.ErrorIs(t, err, media.ErrAssetNotFound)
	})

	t.Run("CapacityNotFound", func(t *testing.T) {
		// Save asset but not capacity info
		asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
		require.NoError(t, repo.SaveAsset(ctx, asset))

		info, err := repo.GetCapacityInfo(ctx, asset.ID)
		assert.Nil(t, info)
		assert.ErrorIs(t, err, media.ErrCapacityNotFound)
	})

	t.Run("ZeroID", func(t *testing.T) {
		info, err := repo.GetCapacityInfo(ctx, media.AssetID{})
		assert.Nil(t, info)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "zero asset ID")
	})

	t.Run("CancelledContext", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		info, err := repo.GetCapacityInfo(cancelledCtx, media.NewAssetID())
		assert.Nil(t, info)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestClear tests repository clearing.
func TestClear(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository(slog.Default())

	// Add some data
	asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
	require.NoError(t, repo.SaveAsset(ctx, asset))

	info := createTestCapacityInfo(asset.ID)
	require.NoError(t, repo.SaveCapacityInfo(ctx, info))

	// Clear
	err := repo.Clear(ctx)
	require.NoError(t, err)

	// Verify empty
	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Verify asset gone
	retrieved, err := repo.GetAsset(ctx, asset.ID)
	assert.Nil(t, retrieved)
	assert.ErrorIs(t, err, media.ErrAssetNotFound)
}

// TestCount tests asset counting.
func TestCount(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository(slog.Default())

	// Initially zero
	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Add assets
	for i := 0; i < 5; i++ {
		asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
		require.NoError(t, repo.SaveAsset(ctx, asset))
	}

	// Verify count
	count, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, count)

	// Delete one
	asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
	require.NoError(t, repo.SaveAsset(ctx, asset))
	require.NoError(t, repo.DeleteAsset(ctx, asset.ID))

	// Count unchanged (we added then deleted)
	count, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

// TestConcurrentAccess tests thread-safety with concurrent operations.
func TestConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository(slog.Default())

	t.Run("ConcurrentSaves", func(t *testing.T) {
		require.NoError(t, repo.Clear(ctx))

		const numGoroutines = 10
		const assetsPerGoroutine = 10

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Concurrent saves
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < assetsPerGoroutine; j++ {
					asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
					err := repo.SaveAsset(ctx, asset)
					assert.NoError(t, err)
				}
			}()
		}

		wg.Wait()

		// Verify all saved
		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, numGoroutines*assetsPerGoroutine, count)
	})

	t.Run("ConcurrentReadsAndWrites", func(t *testing.T) {
		require.NoError(t, repo.Clear(ctx))

		// Pre-populate
		assets := make([]*media.MediaAsset, 20)
		for i := 0; i < 20; i++ {
			assets[i] = createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
			require.NoError(t, repo.SaveAsset(ctx, assets[i]))
		}

		var wg sync.WaitGroup
		const numReaders = 5
		const numWriters = 5

		// Start readers
		wg.Add(numReaders)
		for i := 0; i < numReaders; i++ {
			go func(idx int) {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					assetIdx := (idx*100 + j) % len(assets)
					_, err := repo.GetAsset(ctx, assets[assetIdx].ID)
					assert.NoError(t, err)
				}
			}(i)
		}

		// Start writers
		wg.Add(numWriters)
		for i := 0; i < numWriters; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
					err := repo.SaveAsset(ctx, asset)
					assert.NoError(t, err)
				}
			}()
		}

		wg.Wait()

		// Verify integrity
		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 20) // At least original 20
	})

	t.Run("ConcurrentListAndModify", func(t *testing.T) {
		require.NoError(t, repo.Clear(ctx))

		var wg sync.WaitGroup
		const duration = 100 * time.Millisecond

		// Continuous listing
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			for time.Since(start) < duration {
				_, err := repo.ListAssets(ctx, "", 0, 0)
				assert.NoError(t, err)
			}
		}()

		// Continuous modifications
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			for time.Since(start) < duration {
				asset := createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
				err := repo.SaveAsset(ctx, asset)
				assert.NoError(t, err)
			}
		}()

		wg.Wait()
	})

	t.Run("ConcurrentDeleteAndRead", func(t *testing.T) {
		require.NoError(t, repo.Clear(ctx))

		// Pre-populate
		assets := make([]*media.MediaAsset, 50)
		for i := 0; i < 50; i++ {
			assets[i] = createTestAsset(t, media.MediaTypeImage, media.FormatPNG)
			require.NoError(t, repo.SaveAsset(ctx, assets[i]))
		}

		var wg sync.WaitGroup
		const numDeleters = 3

		// Start deleters
		wg.Add(numDeleters)
		for i := 0; i < numDeleters; i++ {
			go func(idx int) {
				defer wg.Done()
				start := idx * (len(assets) / numDeleters)
				end := start + (len(assets) / numDeleters)
				for j := start; j < end; j++ {
					_ = repo.DeleteAsset(ctx, assets[j].ID)
				}
			}(i)
		}

		// Start readers (may get not found errors)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, asset := range assets {
				_, _ = repo.GetAsset(ctx, asset.ID)
			}
		}()

		wg.Wait()

		// Verify some deleted
		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Less(t, count, 50)
	})
}

// Helper functions

func createTestAsset(t *testing.T, mediaType media.MediaType, format media.MediaFormat) *media.MediaAsset {
	t.Helper()

	id := media.NewAssetID()
	data := []byte("test data")

	asset, err := media.NewMediaAsset(id, mediaType, format, data)
	require.NoError(t, err)

	// Set dimensions for image assets
	if mediaType == media.MediaTypeImage {
		dims, err := media.NewDimensions(800, 600)
		require.NoError(t, err)
		err = asset.SetDimensions(dims)
		require.NoError(t, err)
	}

	// Set sample rate for audio assets
	if mediaType == media.MediaTypeAudio {
		sampleRate, err := media.NewSampleRate(44100, 16, 2)
		require.NoError(t, err)
		err = asset.SetSampleRate(sampleRate)
		require.NoError(t, err)
	}

	return asset
}

func createTestCapacityInfo(assetID media.AssetID) *media.CapacityInfo {
	info, err := media.NewCapacityInfo(
		assetID,
		media.MediaTypeImage,
		10000, // total
		7000,  // usable
		5000,  // recommended
	)
	if err != nil {
		panic(err) // Should never happen in tests
	}
	return info
}
