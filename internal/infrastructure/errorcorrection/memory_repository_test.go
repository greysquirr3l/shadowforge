package errorcorrection_test

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
	infra "github.com/greysquirr3l/shadowforge/internal/infrastructure/errorcorrection"
)

func TestNewMemoryRepository(t *testing.T) {
	t.Run("with_logger", func(t *testing.T) {
		logger := slog.Default()
		repo := infra.NewMemoryRepository(logger)
		assert.NotNil(t, repo)
		assert.Equal(t, 0, repo.MessageCount())
		assert.Equal(t, 0, repo.ShardCount())
	})

	t.Run("with_nil_logger", func(t *testing.T) {
		repo := infra.NewMemoryRepository(nil)
		assert.NotNil(t, repo)
	})
}

func TestMemoryRepository_SaveAndGetMessage(t *testing.T) {
	ctx := context.Background()
	repo := infra.NewMemoryRepository(slog.Default())

	t.Run("success", func(t *testing.T) {
		// Create test message
		messageID := errorcorrection.GenerateMessageID()
		message, err := errorcorrection.NewProtectedMessage(messageID, []byte("test data"), 10, 5)
		require.NoError(t, err)

		// Save
		err = repo.SaveMessage(ctx, message)
		require.NoError(t, err)

		// Get
		retrieved, err := repo.GetMessage(ctx, messageID)
		require.NoError(t, err)
		assert.Equal(t, message.ID, retrieved.ID)
		assert.Equal(t, message.DataShards, retrieved.DataShards)
		assert.Equal(t, message.ParityShards, retrieved.ParityShards)
	})

	t.Run("error_nil_message", func(t *testing.T) {
		err := repo.SaveMessage(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("error_message_not_found", func(t *testing.T) {
		nonExistentID := errorcorrection.GenerateMessageID()
		message, err := repo.GetMessage(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, message)
		assert.ErrorIs(t, err, infra.ErrMessageNotFound)
	})
}

func TestMemoryRepository_DeleteMessage(t *testing.T) {
	ctx := context.Background()
	repo := infra.NewMemoryRepository(slog.Default())

	t.Run("success", func(t *testing.T) {
		// Create and save message
		messageID := errorcorrection.GenerateMessageID()
		message, err := errorcorrection.NewProtectedMessage(messageID, []byte("test data"), 10, 5)
		require.NoError(t, err)
		err = repo.SaveMessage(ctx, message)
		require.NoError(t, err)

		// Verify it exists
		assert.Equal(t, 1, repo.MessageCount())

		// Delete
		err = repo.DeleteMessage(ctx, messageID)
		require.NoError(t, err)

		// Verify it's gone
		assert.Equal(t, 0, repo.MessageCount())

		// Try to get it
		_, err = repo.GetMessage(ctx, messageID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, infra.ErrMessageNotFound)
	})

	t.Run("error_not_found", func(t *testing.T) {
		nonExistentID := errorcorrection.GenerateMessageID()
		err := repo.DeleteMessage(ctx, nonExistentID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, infra.ErrMessageNotFound)
	})
}

func TestMemoryRepository_SaveAndGetShard(t *testing.T) {
	ctx := context.Background()
	repo := infra.NewMemoryRepository(slog.Default())

	t.Run("success", func(t *testing.T) {
		// Create test shard
		shardID := errorcorrection.GenerateShardID()
		messageID := errorcorrection.GenerateMessageID()
		shard, err := errorcorrection.NewShard(shardID, messageID, 0, []byte("shard data"), false)
		require.NoError(t, err)

		// Save
		err = repo.SaveShard(ctx, shard)
		require.NoError(t, err)

		// Get
		retrieved, err := repo.GetShard(ctx, shardID)
		require.NoError(t, err)
		assert.Equal(t, shard.ID, retrieved.ID)
		assert.Equal(t, shard.MessageID, retrieved.MessageID)
		assert.Equal(t, shard.Index, retrieved.Index)
		assert.Equal(t, shard.Data, retrieved.Data)
	})

	t.Run("error_nil_shard", func(t *testing.T) {
		err := repo.SaveShard(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("error_shard_not_found", func(t *testing.T) {
		nonExistentID := errorcorrection.GenerateShardID()
		shard, err := repo.GetShard(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, shard)
		assert.ErrorIs(t, err, infra.ErrShardNotFound)
	})
}

func TestMemoryRepository_GetShardsByMessage(t *testing.T) {
	ctx := context.Background()
	repo := infra.NewMemoryRepository(slog.Default())

	t.Run("success_multiple_shards", func(t *testing.T) {
		messageID := errorcorrection.GenerateMessageID()

		// Create and save multiple shards for the same message
		for i := 0; i < 5; i++ {
			shardID := errorcorrection.GenerateShardID()
			shard, err := errorcorrection.NewShard(shardID, messageID, i, []byte("data"), i >= 3)
			require.NoError(t, err)
			err = repo.SaveShard(ctx, shard)
			require.NoError(t, err)
		}

		// Get shards by message ID
		shards, err := repo.GetShardsByMessage(ctx, messageID)
		require.NoError(t, err)
		assert.Len(t, shards, 5)
	})

	t.Run("success_no_shards", func(t *testing.T) {
		nonExistentMessageID := errorcorrection.GenerateMessageID()
		shards, err := repo.GetShardsByMessage(ctx, nonExistentMessageID)
		require.NoError(t, err)
		assert.Empty(t, shards)
	})

	t.Run("success_filters_by_message_id", func(t *testing.T) {
		messageID1 := errorcorrection.GenerateMessageID()
		messageID2 := errorcorrection.GenerateMessageID()

		// Save shards for message 1
		for i := 0; i < 3; i++ {
			shardID := errorcorrection.GenerateShardID()
			shard, err := errorcorrection.NewShard(shardID, messageID1, i, []byte("data1"), false)
			require.NoError(t, err)
			err = repo.SaveShard(ctx, shard)
			require.NoError(t, err)
		}

		// Save shards for message 2
		for i := 0; i < 2; i++ {
			shardID := errorcorrection.GenerateShardID()
			shard, err := errorcorrection.NewShard(shardID, messageID2, i, []byte("data2"), false)
			require.NoError(t, err)
			err = repo.SaveShard(ctx, shard)
			require.NoError(t, err)
		}

		// Get shards for message 1
		shards1, err := repo.GetShardsByMessage(ctx, messageID1)
		require.NoError(t, err)
		assert.Len(t, shards1, 3)

		// Get shards for message 2
		shards2, err := repo.GetShardsByMessage(ctx, messageID2)
		require.NoError(t, err)
		assert.Len(t, shards2, 2)
	})
}

func TestMemoryRepository_DeleteShard(t *testing.T) {
	ctx := context.Background()
	repo := infra.NewMemoryRepository(slog.Default())

	t.Run("success", func(t *testing.T) {
		// Create and save shard
		shardID := errorcorrection.GenerateShardID()
		messageID := errorcorrection.GenerateMessageID()
		shard, err := errorcorrection.NewShard(shardID, messageID, 0, []byte("data"), false)
		require.NoError(t, err)
		err = repo.SaveShard(ctx, shard)
		require.NoError(t, err)

		// Verify it exists
		assert.Equal(t, 1, repo.ShardCount())

		// Delete
		err = repo.DeleteShard(ctx, shardID)
		require.NoError(t, err)

		// Verify it's gone
		assert.Equal(t, 0, repo.ShardCount())

		// Try to get it
		_, err = repo.GetShard(ctx, shardID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, infra.ErrShardNotFound)
	})

	t.Run("error_not_found", func(t *testing.T) {
		nonExistentID := errorcorrection.GenerateShardID()
		err := repo.DeleteShard(ctx, nonExistentID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, infra.ErrShardNotFound)
	})
}

func TestMemoryRepository_Clear(t *testing.T) {
	ctx := context.Background()
	repo := infra.NewMemoryRepository(slog.Default())

	// Add some messages and shards
	for i := 0; i < 3; i++ {
		messageID := errorcorrection.GenerateMessageID()
		message, err := errorcorrection.NewProtectedMessage(messageID, []byte("data"), 10, 5)
		require.NoError(t, err)
		err = repo.SaveMessage(ctx, message)
		require.NoError(t, err)

		shardID := errorcorrection.GenerateShardID()
		shard, err := errorcorrection.NewShard(shardID, messageID, i, []byte("shard"), false)
		require.NoError(t, err)
		err = repo.SaveShard(ctx, shard)
		require.NoError(t, err)
	}

	// Verify data exists
	assert.Equal(t, 3, repo.MessageCount())
	assert.Equal(t, 3, repo.ShardCount())

	// Clear
	repo.Clear()

	// Verify all data is gone
	assert.Equal(t, 0, repo.MessageCount())
	assert.Equal(t, 0, repo.ShardCount())
}

func TestMemoryRepository_ConcurrentMessageOperations(t *testing.T) {
	ctx := context.Background()
	repo := infra.NewMemoryRepository(slog.Default())

	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*3) // save, get, delete

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Create message
			messageID := errorcorrection.GenerateMessageID()
			message, err := errorcorrection.NewProtectedMessage(messageID, []byte("data"), 10, 5)
			if err != nil {
				errors <- err
				return
			}

			// Save
			if err := repo.SaveMessage(ctx, message); err != nil {
				errors <- err
				return
			}

			// Get
			if _, err := repo.GetMessage(ctx, messageID); err != nil {
				errors <- err
				return
			}

			// Delete
			if err := repo.DeleteMessage(ctx, messageID); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check no errors occurred
	for err := range errors {
		t.Errorf("Concurrent message operation error: %v", err)
	}

	// Verify repository is empty
	assert.Equal(t, 0, repo.MessageCount())
}

func TestMemoryRepository_ConcurrentShardOperations(t *testing.T) {
	ctx := context.Background()
	repo := infra.NewMemoryRepository(slog.Default())

	messageID := errorcorrection.GenerateMessageID()
	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*3)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Create shard
			shardID := errorcorrection.GenerateShardID()
			shard, err := errorcorrection.NewShard(shardID, messageID, index, []byte("data"), false)
			if err != nil {
				errors <- err
				return
			}

			// Save
			if err := repo.SaveShard(ctx, shard); err != nil {
				errors <- err
				return
			}

			// Get
			if _, err := repo.GetShard(ctx, shardID); err != nil {
				errors <- err
				return
			}

			// Delete
			if err := repo.DeleteShard(ctx, shardID); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check no errors occurred
	for err := range errors {
		t.Errorf("Concurrent shard operation error: %v", err)
	}

	// Verify repository is empty
	assert.Equal(t, 0, repo.ShardCount())
}

func TestMemoryRepository_ConcurrentGetShardsByMessage(t *testing.T) {
	ctx := context.Background()
	repo := infra.NewMemoryRepository(slog.Default())

	messageID := errorcorrection.GenerateMessageID()

	// Pre-populate with shards
	const numShards = 15
	for i := 0; i < numShards; i++ {
		shardID := errorcorrection.GenerateShardID()
		shard, err := errorcorrection.NewShard(shardID, messageID, i, []byte("data"), i >= 10)
		require.NoError(t, err)
		err = repo.SaveShard(ctx, shard)
		require.NoError(t, err)
	}

	// Concurrent reads
	const numGoroutines = 20
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			shards, err := repo.GetShardsByMessage(ctx, messageID)
			if err != nil {
				errors <- err
				return
			}

			if len(shards) != numShards {
				errors <- assert.AnError
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check no errors occurred
	for err := range errors {
		t.Errorf("Concurrent GetShardsByMessage error: %v", err)
	}
}
