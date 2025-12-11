package errorcorrection_test

import (
	"context"
	"crypto/rand"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
	infra "github.com/greysquirr3l/shadowforge/internal/infrastructure/errorcorrection"
)

func TestNewRSService(t *testing.T) {
	t.Run("with_logger", func(t *testing.T) {
		logger := slog.Default()
		service := infra.NewRSService(logger)

		assert.NotNil(t, service)
	})

	t.Run("with_nil_logger", func(t *testing.T) {
		service := infra.NewRSService(nil)

		assert.NotNil(t, service)
	})
}

func TestRSService_Encode(t *testing.T) {
	service := infra.NewRSService(slog.Default())
	ctx := context.Background()

	t.Run("success_balanced_config", func(t *testing.T) {
		// Arrange
		data := make([]byte, 10240) // 10KB
		_, err := rand.Read(data)
		require.NoError(t, err)

		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		// Act
		message, err := service.Encode(ctx, data, config)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, message)
		assert.NotEmpty(t, message.ID.String())
		assert.Equal(t, 15, len(message.Shards))
		assert.Equal(t, config.DataShards, message.DataShards)
		assert.Equal(t, config.ParityShards, message.ParityShards)

		// Verify data shards (first 10)
		dataShardCount := 0
		parityShardCount := 0
		for _, shard := range message.Shards {
			if shard.IsParity {
				parityShardCount++
			} else {
				dataShardCount++
			}
			assert.NotNil(t, shard.Checksum)
			assert.Len(t, shard.Checksum, 32) // SHA-256 size
		}
		assert.Equal(t, 10, dataShardCount)
		assert.Equal(t, 5, parityShardCount)
	})

	t.Run("success_conservative_config", func(t *testing.T) {
		data := []byte("test data for conservative encoding")
		config, err := errorcorrection.NewShardConfiguration(5, 5)
		require.NoError(t, err)

		message, err := service.Encode(ctx, data, config)

		require.NoError(t, err)
		assert.Equal(t, 10, len(message.Shards))
	})

	t.Run("success_aggressive_config", func(t *testing.T) {
		data := make([]byte, 102400) // 100KB
		_, err := rand.Read(data)
		require.NoError(t, err)

		config, err := errorcorrection.NewShardConfiguration(15, 3)
		require.NoError(t, err)

		message, err := service.Encode(ctx, data, config)

		require.NoError(t, err)
		assert.Equal(t, 18, len(message.Shards))
	})

	t.Run("error_empty_data", func(t *testing.T) {
		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		message, err := service.Encode(ctx, []byte{}, config)

		assert.Error(t, err)
		assert.Nil(t, message)
		assert.ErrorIs(t, err, errorcorrection.ErrEmptyData)
	})

	t.Run("error_nil_config", func(t *testing.T) {
		data := []byte("test data")

		message, err := service.Encode(ctx, data, nil)

		assert.Error(t, err)
		assert.Nil(t, message)
		assert.ErrorIs(t, err, errorcorrection.ErrInvalidShardCount)
	})

	t.Run("error_cancelled_context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		data := []byte("test data")
		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		message, err := service.Encode(ctx, data, config)

		assert.Error(t, err)
		assert.Nil(t, message)
	})
}

func TestRSService_Decode(t *testing.T) {
	service := infra.NewRSService(slog.Default())
	ctx := context.Background()

	t.Run("success_all_shards_present", func(t *testing.T) {
		// Arrange - Encode data first
		originalData := []byte("This is test data for Reed-Solomon encoding and decoding verification")
		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		message, err := service.Encode(ctx, originalData, config)
		require.NoError(t, err)

		// Act - Decode with all shards
		recoveredData, err := service.Decode(ctx, message.Shards, config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, originalData, recoveredData)
	})

	t.Run("success_with_missing_shards_reconstruction", func(t *testing.T) {
		// Arrange
		originalData := make([]byte, 10240)
		_, err := rand.Read(originalData)
		require.NoError(t, err)

		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		message, err := service.Encode(ctx, originalData, config)
		require.NoError(t, err)

		// Simulate losing 5 shards (maximum allowed with this config)
		availableShards := message.Shards[5:] // Keep last 10 shards (lose first 5)
		require.Len(t, availableShards, 10)

		// Act - Decode with missing shards
		recoveredData, err := service.Decode(ctx, availableShards, config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, originalData, recoveredData)
	})

	t.Run("success_with_parity_shards_only", func(t *testing.T) {
		// Test reconstruction with mix of data and parity shards
		originalData := []byte("Testing reconstruction with mixed shard types")
		config, err := errorcorrection.NewShardConfiguration(5, 5)
		require.NoError(t, err)

		message, err := service.Encode(ctx, originalData, config)
		require.NoError(t, err)

		// Keep some data shards (indices 0, 1, 2) and some parity shards (indices 7, 8, 9)
		// This gives us 6 shards total when we only need 5
		availableShards := []*errorcorrection.Shard{
			message.Shards[0],
			message.Shards[1],
			message.Shards[2],
			message.Shards[7],
			message.Shards[8],
			message.Shards[9],
		}

		recoveredData, err := service.Decode(ctx, availableShards, config)

		require.NoError(t, err)
		assert.Equal(t, originalData, recoveredData)
	})

	t.Run("error_insufficient_shards", func(t *testing.T) {
		// Need 10 shards but only provide 9
		originalData := []byte("test data")
		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		message, err := service.Encode(ctx, originalData, config)
		require.NoError(t, err)

		// Only provide 9 shards (insufficient)
		insufficientShards := message.Shards[:9]

		recoveredData, err := service.Decode(ctx, insufficientShards, config)

		assert.Error(t, err)
		assert.Nil(t, recoveredData)
		assert.ErrorIs(t, err, errorcorrection.ErrInsufficientShards)
	})

	t.Run("error_empty_shards", func(t *testing.T) {
		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		recoveredData, err := service.Decode(ctx, []*errorcorrection.Shard{}, config)

		assert.Error(t, err)
		assert.Nil(t, recoveredData)
	})

	t.Run("error_nil_config", func(t *testing.T) {
		originalData := []byte("test data")
		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		message, err := service.Encode(ctx, originalData, config)
		require.NoError(t, err)

		recoveredData, err := service.Decode(ctx, message.Shards, nil)

		assert.Error(t, err)
		assert.Nil(t, recoveredData)
	})

	t.Run("error_corrupted_shard_checksum", func(t *testing.T) {
		// Encode data
		originalData := []byte("test data for checksum verification")
		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		message, err := service.Encode(ctx, originalData, config)
		require.NoError(t, err)

		// Corrupt a shard's data (without updating checksum)
		message.Shards[0].Data[0] ^= 0xFF

		recoveredData, err := service.Decode(ctx, message.Shards, config)

		assert.Error(t, err)
		assert.Nil(t, recoveredData)
		assert.ErrorIs(t, err, errorcorrection.ErrChecksumMismatch)
	})

	t.Run("error_cancelled_context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		originalData := []byte("test data")
		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		message, err := service.Encode(context.Background(), originalData, config)
		require.NoError(t, err)

		recoveredData, err := service.Decode(ctx, message.Shards, config)

		assert.Error(t, err)
		assert.Nil(t, recoveredData)
	})
}

func TestRSService_VerifyShards(t *testing.T) {
	service := infra.NewRSService(slog.Default())
	ctx := context.Background()

	t.Run("success_all_valid", func(t *testing.T) {
		// Encode data to get valid shards
		data := []byte("test data for verification")
		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		message, err := service.Encode(ctx, data, config)
		require.NoError(t, err)

		// Verify all shards
		err = service.VerifyShards(ctx, message.Shards)

		assert.NoError(t, err)
	})

	t.Run("error_corrupted_shard", func(t *testing.T) {
		// Encode data
		data := []byte("test data")
		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		message, err := service.Encode(ctx, data, config)
		require.NoError(t, err)

		// Corrupt a shard
		message.Shards[0].Data[0] ^= 0xFF

		err = service.VerifyShards(ctx, message.Shards)

		assert.Error(t, err)
		assert.ErrorIs(t, err, errorcorrection.ErrCorruptedShard)
	})

	t.Run("error_empty_shards", func(t *testing.T) {
		err := service.VerifyShards(ctx, []*errorcorrection.Shard{})

		assert.Error(t, err)
	})

	t.Run("error_cancelled_context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		data := []byte("test data")
		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		message, err := service.Encode(context.Background(), data, config)
		require.NoError(t, err)

		err = service.VerifyShards(ctx, message.Shards)

		assert.Error(t, err)
	})
}

func TestRSService_CalculateCapacity(t *testing.T) {
	service := infra.NewRSService(slog.Default())

	t.Run("success_balanced_config", func(t *testing.T) {
		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		capacity := service.CalculateCapacity(config, 1024)

		assert.Equal(t, 10*1024, capacity)
	})

	t.Run("success_conservative_config", func(t *testing.T) {
		config, err := errorcorrection.NewShardConfiguration(5, 5)
		require.NoError(t, err)

		capacity := service.CalculateCapacity(config, 2048)

		assert.Equal(t, 5*2048, capacity)
	})

	t.Run("nil_config", func(t *testing.T) {
		capacity := service.CalculateCapacity(nil, 1024)

		assert.Equal(t, 0, capacity)
	})

	t.Run("invalid_shard_size", func(t *testing.T) {
		config, err := errorcorrection.NewShardConfiguration(10, 5)
		require.NoError(t, err)

		capacity := service.CalculateCapacity(config, 0)

		assert.Equal(t, 0, capacity)
	})
}

func TestRSService_OptimizeConfiguration(t *testing.T) {
	service := infra.NewRSService(slog.Default())
	ctx := context.Background()

	t.Run("small_data_conservative", func(t *testing.T) {
		dataSize := 5 * 1024 // 5KB

		config, err := service.OptimizeConfiguration(ctx, dataSize, errorcorrection.RedundancyConservative)

		require.NoError(t, err)
		assert.Equal(t, 5, config.DataShards)
		assert.Equal(t, 2, config.ParityShards) // 50% redundancy: 5 * 0.5 = 2.5 → 2
	})

	t.Run("medium_data_balanced", func(t *testing.T) {
		dataSize := 50 * 1024 // 50KB

		config, err := service.OptimizeConfiguration(ctx, dataSize, errorcorrection.RedundancyBalanced)

		require.NoError(t, err)
		assert.Equal(t, 10, config.DataShards)
		assert.Equal(t, 3, config.ParityShards) // 33% redundancy: 10 * 0.33 = 3.3 → 3
	})

	t.Run("large_data_aggressive", func(t *testing.T) {
		dataSize := 200 * 1024 * 1024 // 200MB

		config, err := service.OptimizeConfiguration(ctx, dataSize, errorcorrection.RedundancyLow)

		require.NoError(t, err)
		assert.Equal(t, 20, config.DataShards)
		assert.Equal(t, 3, config.ParityShards) // 17% redundancy: 20 * 0.17 = 3.4 → 3 (approx)
	})

	t.Run("error_invalid_data_size", func(t *testing.T) {
		config, err := service.OptimizeConfiguration(ctx, 0, errorcorrection.RedundancyBalanced)

		assert.Error(t, err)
		assert.Nil(t, config)
	})

	t.Run("error_negative_data_size", func(t *testing.T) {
		config, err := service.OptimizeConfiguration(ctx, -100, errorcorrection.RedundancyBalanced)

		assert.Error(t, err)
		assert.Nil(t, config)
	})

	t.Run("error_cancelled_context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		config, err := service.OptimizeConfiguration(ctx, 10240, errorcorrection.RedundancyBalanced)

		assert.Error(t, err)
		assert.Nil(t, config)
	})
}

func TestRSService_ConcurrentEncode(t *testing.T) {
	service := infra.NewRSService(slog.Default())
	ctx := context.Background()

	// Test concurrent encoding operations
	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			data := make([]byte, 1024)
			_, err := rand.Read(data)
			if err != nil {
				errors <- err
				return
			}

			config, err := errorcorrection.NewShardConfiguration(10, 5)
			if err != nil {
				errors <- err
				return
			}

			_, err = service.Encode(ctx, data, config)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check no errors occurred
	for err := range errors {
		t.Errorf("Concurrent encode error: %v", err)
	}
}

func TestRSService_ConcurrentDecode(t *testing.T) {
	service := infra.NewRSService(slog.Default())
	ctx := context.Background()

	// Prepare test messages
	const numMessages = 10
	messages := make([]*errorcorrection.ProtectedMessage, numMessages)
	originalData := make([][]byte, numMessages)

	config, err := errorcorrection.NewShardConfiguration(10, 5)
	require.NoError(t, err)

	for i := 0; i < numMessages; i++ {
		data := make([]byte, 1024)
		_, err := rand.Read(data)
		require.NoError(t, err)

		originalData[i] = data
		messages[i], err = service.Encode(ctx, data, config)
		require.NoError(t, err)
	}

	// Test concurrent decoding
	var wg sync.WaitGroup
	errors := make(chan error, numMessages)

	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			recovered, err := service.Decode(ctx, messages[index].Shards, config)
			if err != nil {
				errors <- err
				return
			}

			if !assert.Equal(t, originalData[index], recovered) {
				errors <- assert.AnError
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check no errors occurred
	for err := range errors {
		t.Errorf("Concurrent decode error: %v", err)
	}
}
