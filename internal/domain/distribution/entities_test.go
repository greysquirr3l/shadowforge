package distribution

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== DistributionStrategy Tests ====================

func TestDistributionStrategy_NewDistributionStrategy_Success(t *testing.T) {
	tests := []struct {
		name           string
		pattern        PatternType
		dataShards     int
		parityShards   int
		targetMediaIDs []string
	}{
		{
			name:           "one_to_many_balanced",
			pattern:        PatternOneToMany,
			dataShards:     10,
			parityShards:   5,
			targetMediaIDs: []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8", "m9", "m10", "m11", "m12", "m13", "m14", "m15"},
		},
		{
			name:           "one_to_one",
			pattern:        PatternOneToOne,
			dataShards:     1,
			parityShards:   0,
			targetMediaIDs: []string{"media1"},
		},
		{
			name:           "many_to_many",
			pattern:        PatternManyToMany,
			dataShards:     8,
			parityShards:   4,
			targetMediaIDs: []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8", "m9", "m10", "m11", "m12"},
		},
		{
			name:           "high_redundancy",
			pattern:        PatternOneToMany,
			dataShards:     5,
			parityShards:   5,
			targetMediaIDs: []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8", "m9", "m10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, err := NewDistributionStrategy(tt.pattern, tt.dataShards, tt.parityShards, tt.targetMediaIDs)

			require.NoError(t, err)
			assert.NotNil(t, strategy)
			assert.False(t, strategy.ID.IsZero())
			assert.Equal(t, tt.pattern, strategy.Pattern)
			assert.Equal(t, tt.dataShards, strategy.DataShards)
			assert.Equal(t, tt.parityShards, strategy.ParityShards)
			assert.Equal(t, tt.dataShards+tt.parityShards, strategy.TotalShards)
			assert.Equal(t, tt.dataShards, strategy.Threshold)
			assert.Equal(t, tt.targetMediaIDs, strategy.TargetMediaIDs)
			assert.Equal(t, StatusPending, strategy.Status)
			assert.WithinDuration(t, time.Now(), strategy.CreatedAt, time.Second)
			assert.Nil(t, strategy.CompletedAt)
		})
	}
}

func TestDistributionStrategy_NewDistributionStrategy_InvalidPattern(t *testing.T) {
	strategy, err := NewDistributionStrategy("invalid", 5, 3, []string{"m1", "m2"})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPattern)
	assert.Nil(t, strategy)
}

func TestDistributionStrategy_NewDistributionStrategy_InvalidDataShards(t *testing.T) {
	tests := []struct {
		name       string
		dataShards int
	}{
		{"zero", 0},
		{"negative", -1},
		{"large_negative", -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, err := NewDistributionStrategy(PatternOneToMany, tt.dataShards, 3, []string{"m1", "m2"})

			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidDataShards)
			assert.Nil(t, strategy)
		})
	}
}

func TestDistributionStrategy_NewDistributionStrategy_InvalidParityShards(t *testing.T) {
	strategy, err := NewDistributionStrategy(PatternOneToMany, 5, -1, []string{"m1", "m2"})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidParityShards)
	assert.Nil(t, strategy)
}

func TestDistributionStrategy_NewDistributionStrategy_NoTargetMedia(t *testing.T) {
	strategy, err := NewDistributionStrategy(PatternOneToMany, 5, 3, []string{})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoTargetMedia)
	assert.Nil(t, strategy)
}

func TestDistributionStrategy_NewDistributionStrategy_InsufficientMedia(t *testing.T) {
	// Need 8 shards (5 data + 3 parity) but only have 5 media
	strategy, err := NewDistributionStrategy(PatternOneToMany, 5, 3, []string{"m1", "m2", "m3", "m4", "m5"})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInsufficientMedia)
	assert.Nil(t, strategy)
}

func TestDistributionStrategy_Complete_Success(t *testing.T) {
	strategy, _ := NewDistributionStrategy(PatternOneToMany, 5, 3, []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8"})
	manifest, _ := NewShardManifest(strategy.ID, []ShardMetadata{
		{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"},
		{Index: 1, Size: 1024, Checksum: "def", MediaID: "m2"},
		{Index: 2, Size: 1024, Checksum: "ghi", MediaID: "m3"},
		{Index: 3, Size: 1024, Checksum: "jkl", MediaID: "m4"},
		{Index: 4, Size: 1024, Checksum: "mno", MediaID: "m5"},
	}, 5)

	err := strategy.Complete(manifest)

	require.NoError(t, err)
	assert.Equal(t, StatusComplete, strategy.Status)
	assert.Equal(t, manifest, strategy.Manifest)
	assert.NotNil(t, strategy.CompletedAt)
	assert.WithinDuration(t, time.Now(), *strategy.CompletedAt, time.Second)
}

func TestDistributionStrategy_Complete_AlreadyComplete(t *testing.T) {
	strategy, _ := NewDistributionStrategy(PatternOneToMany, 5, 3, []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8"})
	manifest1, _ := NewShardManifest(strategy.ID, []ShardMetadata{
		{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"},
		{Index: 1, Size: 1024, Checksum: "def", MediaID: "m2"},
		{Index: 2, Size: 1024, Checksum: "ghi", MediaID: "m3"},
		{Index: 3, Size: 1024, Checksum: "jkl", MediaID: "m4"},
		{Index: 4, Size: 1024, Checksum: "mno", MediaID: "m5"},
	}, 5)
	require.NoError(t, strategy.Complete(manifest1))

	manifest2, _ := NewShardManifest(strategy.ID, []ShardMetadata{
		{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"},
		{Index: 1, Size: 1024, Checksum: "def", MediaID: "m2"},
		{Index: 2, Size: 1024, Checksum: "ghi", MediaID: "m3"},
		{Index: 3, Size: 1024, Checksum: "jkl", MediaID: "m4"},
		{Index: 4, Size: 1024, Checksum: "mno", MediaID: "m5"},
	}, 5)
	err := strategy.Complete(manifest2)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAlreadyComplete)
}

func TestDistributionStrategy_Complete_AfterFailure(t *testing.T) {
	strategy, _ := NewDistributionStrategy(PatternOneToMany, 5, 3, []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8"})
	manifest, _ := NewShardManifest(strategy.ID, []ShardMetadata{
		{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"},
	}, 5)
	require.NoError(t, strategy.MarkFailed())

	err := strategy.Complete(manifest)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCannotCompleteAfterFailure)
}

func TestDistributionStrategy_Complete_NilManifest(t *testing.T) {
	strategy, _ := NewDistributionStrategy(PatternOneToMany, 5, 3, []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8"})

	err := strategy.Complete(nil)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrManifestRequired)
}

func TestDistributionStrategy_MarkFailed_Success(t *testing.T) {
	strategy, _ := NewDistributionStrategy(PatternOneToMany, 5, 3, []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8"})

	err := strategy.MarkFailed()

	require.NoError(t, err)
	assert.Equal(t, StatusFailed, strategy.Status)
}

func TestDistributionStrategy_MarkFailed_AfterComplete(t *testing.T) {
	strategy, _ := NewDistributionStrategy(PatternOneToMany, 5, 3, []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8"})
	manifest, _ := NewShardManifest(strategy.ID, []ShardMetadata{
		{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"},
		{Index: 1, Size: 1024, Checksum: "def", MediaID: "m2"},
		{Index: 2, Size: 1024, Checksum: "ghi", MediaID: "m3"},
		{Index: 3, Size: 1024, Checksum: "jkl", MediaID: "m4"},
		{Index: 4, Size: 1024, Checksum: "mno", MediaID: "m5"},
	}, 5)
	require.NoError(t, strategy.Complete(manifest))

	err := strategy.MarkFailed()

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCannotFailAfterComplete)
}

func TestDistributionStrategy_CanRecover(t *testing.T) {
	strategy, _ := NewDistributionStrategy(PatternOneToMany, 10, 5, []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8", "m9", "m10", "m11", "m12", "m13", "m14", "m15"})

	tests := []struct {
		name             string
		availableShards  int
		expectedRecovery bool
	}{
		{"exact_threshold", 10, true},
		{"above_threshold", 11, true},
		{"all_shards", 15, true},
		{"below_threshold", 9, false},
		{"minimal_fail", 5, false},
		{"zero", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.CanRecover(tt.availableShards)
			assert.Equal(t, tt.expectedRecovery, result)
		})
	}
}

func TestDistributionStrategy_RedundancyLevel(t *testing.T) {
	tests := []struct {
		name               string
		dataShards         int
		parityShards       int
		expectedRedundancy float64
	}{
		{"50_percent", 10, 5, 0.5},
		{"100_percent", 5, 5, 1.0},
		{"33_percent", 9, 3, 0.333333333},
		{"no_redundancy", 10, 0, 0.0},
		{"200_percent", 3, 6, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, _ := NewDistributionStrategy(
				PatternOneToMany,
				tt.dataShards,
				tt.parityShards,
				make([]string, tt.dataShards+tt.parityShards),
			)

			result := strategy.RedundancyLevel()
			assert.InDelta(t, tt.expectedRedundancy, result, 0.00001)
		})
	}
}

// ==================== ShardManifest Tests ====================

func TestShardManifest_NewShardManifest_Success(t *testing.T) {
	strategyID, _ := NewStrategyID()
	metadata := []ShardMetadata{
		{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"},
		{Index: 1, Size: 2048, Checksum: "def", MediaID: "m2"},
		{Index: 2, Size: 1536, Checksum: "ghi", MediaID: "m3"},
	}

	manifest, err := NewShardManifest(strategyID, metadata, 2)

	require.NoError(t, err)
	assert.NotNil(t, manifest)
	assert.False(t, manifest.ID.IsZero())
	assert.Equal(t, strategyID, manifest.StrategyID)
	assert.Equal(t, metadata, manifest.ShardMetadata)
	assert.Equal(t, 3, manifest.TotalShards)
	assert.Equal(t, 2, manifest.RequiredShards)
	assert.WithinDuration(t, time.Now(), manifest.CreatedAt, time.Second)
	assert.Nil(t, manifest.ExpiresAt)
}

func TestShardManifest_NewShardManifest_ZeroStrategyID(t *testing.T) {
	metadata := []ShardMetadata{
		{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"},
	}

	manifest, err := NewShardManifest(StrategyID{}, metadata, 1)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidStrategyID)
	assert.Nil(t, manifest)
}

func TestShardManifest_NewShardManifest_NoMetadata(t *testing.T) {
	strategyID, _ := NewStrategyID()

	manifest, err := NewShardManifest(strategyID, []ShardMetadata{}, 1)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoShardMetadata)
	assert.Nil(t, manifest)
}

func TestShardManifest_NewShardManifest_InvalidThreshold(t *testing.T) {
	strategyID, _ := NewStrategyID()
	metadata := []ShardMetadata{
		{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"},
		{Index: 1, Size: 2048, Checksum: "def", MediaID: "m2"},
	}

	tests := []struct {
		name      string
		threshold int
	}{
		{"zero", 0},
		{"negative", -1},
		{"exceeds_total", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := NewShardManifest(strategyID, metadata, tt.threshold)

			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidThreshold)
			assert.Nil(t, manifest)
		})
	}
}

func TestShardManifest_IsExpired_NotExpired(t *testing.T) {
	strategyID, _ := NewStrategyID()
	metadata := []ShardMetadata{{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"}}
	manifest, _ := NewShardManifest(strategyID, metadata, 1)

	// No expiry set
	assert.False(t, manifest.IsExpired())

	// Future expiry
	future := time.Now().Add(24 * time.Hour)
	manifest.ExpiresAt = &future
	assert.False(t, manifest.IsExpired())
}

func TestShardManifest_IsExpired_Expired(t *testing.T) {
	strategyID, _ := NewStrategyID()
	metadata := []ShardMetadata{{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"}}
	manifest, _ := NewShardManifest(strategyID, metadata, 1)

	// Past expiry
	past := time.Now().Add(-24 * time.Hour)
	manifest.ExpiresAt = &past

	assert.True(t, manifest.IsExpired())
}

func TestShardManifest_GetShardByIndex_Success(t *testing.T) {
	strategyID, _ := NewStrategyID()
	metadata := []ShardMetadata{
		{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"},
		{Index: 1, Size: 2048, Checksum: "def", MediaID: "m2"},
		{Index: 2, Size: 1536, Checksum: "ghi", MediaID: "m3"},
	}
	manifest, _ := NewShardManifest(strategyID, metadata, 2)

	shard, err := manifest.GetShardByIndex(1)

	require.NoError(t, err)
	assert.Equal(t, metadata[1], shard)
}

func TestShardManifest_GetShardByIndex_InvalidIndex(t *testing.T) {
	strategyID, _ := NewStrategyID()
	metadata := []ShardMetadata{
		{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"},
	}
	manifest, _ := NewShardManifest(strategyID, metadata, 1)

	tests := []struct {
		name  string
		index int
	}{
		{"negative", -1},
		{"out_of_bounds", 1},
		{"far_out_of_bounds", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shard, err := manifest.GetShardByIndex(tt.index)

			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidShardIndex)
			assert.Equal(t, ShardMetadata{}, shard)
		})
	}
}

func TestShardManifest_ValidateCompleteness_Sufficient(t *testing.T) {
	strategyID, _ := NewStrategyID()
	metadata := []ShardMetadata{
		{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"},
		{Index: 1, Size: 2048, Checksum: "def", MediaID: "m2"},
		{Index: 2, Size: 1536, Checksum: "ghi", MediaID: "m3"},
	}
	manifest, _ := NewShardManifest(strategyID, metadata, 2)

	tests := []struct {
		name             string
		availableIndices []int
		expectError      bool
	}{
		{"exact_required", []int{0, 1}, false},
		{"more_than_required", []int{0, 1, 2}, false},
		{"all_shards", []int{0, 1, 2}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manifest.ValidateCompleteness(tt.availableIndices)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShardManifest_ValidateCompleteness_Insufficient(t *testing.T) {
	strategyID, _ := NewStrategyID()
	metadata := []ShardMetadata{
		{Index: 0, Size: 1024, Checksum: "abc", MediaID: "m1"},
		{Index: 1, Size: 2048, Checksum: "def", MediaID: "m2"},
		{Index: 2, Size: 1536, Checksum: "ghi", MediaID: "m3"},
	}
	manifest, _ := NewShardManifest(strategyID, metadata, 2)

	tests := []struct {
		name             string
		availableIndices []int
	}{
		{"one_below", []int{0}},
		{"empty", []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manifest.ValidateCompleteness(tt.availableIndices)

			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrInsufficientShards)
		})
	}
}
