package distribution_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainDist "github.com/greysquirr3l/shadowforge/internal/domain/distribution"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
	infraDist "github.com/greysquirr3l/shadowforge/internal/infrastructure/distribution"
)

// MockRepository implements distribution.Repository for testing.
type MockRepository struct {
	strategies map[string]*domainDist.DistributionStrategy
	manifests  map[string]*domainDist.ShardManifest
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		strategies: make(map[string]*domainDist.DistributionStrategy),
		manifests:  make(map[string]*domainDist.ShardManifest),
	}
}

func (m *MockRepository) SaveStrategy(ctx context.Context, strategy *domainDist.DistributionStrategy) error {
	m.strategies[strategy.ID.String()] = strategy
	return nil
}

func (m *MockRepository) GetStrategy(ctx context.Context, id domainDist.StrategyID) (*domainDist.DistributionStrategy, error) {
	strategy, ok := m.strategies[id.String()]
	if !ok {
		return nil, fmt.Errorf("strategy not found")
	}
	return strategy, nil
}

func (m *MockRepository) SaveManifest(ctx context.Context, manifest *domainDist.ShardManifest) error {
	m.manifests[manifest.ID.String()] = manifest
	return nil
}

func (m *MockRepository) GetManifest(ctx context.Context, id domainDist.ManifestID) (*domainDist.ShardManifest, error) {
	manifest, ok := m.manifests[id.String()]
	if !ok {
		return nil, fmt.Errorf("manifest not found")
	}
	return manifest, nil
}

func (m *MockRepository) DeleteStrategy(ctx context.Context, id domainDist.StrategyID) error {
	delete(m.strategies, id.String())
	return nil
}

func (m *MockRepository) ListStrategies(ctx context.Context, limit, offset int) ([]*domainDist.DistributionStrategy, error) {
	strategies := make([]*domainDist.DistributionStrategy, 0, len(m.strategies))
	for _, strategy := range m.strategies {
		strategies = append(strategies, strategy)
	}
	return strategies, nil
}

func (m *MockRepository) GetManifestsByStrategy(ctx context.Context, id domainDist.StrategyID) ([]*domainDist.ShardManifest, error) {
	manifests := make([]*domainDist.ShardManifest, 0)
	for _, manifest := range m.manifests {
		if manifest.StrategyID.String() == id.String() {
			manifests = append(manifests, manifest)
		}
	}
	return manifests, nil
}

// MockServices for testing distribution service.

type MockECService struct{}

func (m *MockECService) Encode(ctx context.Context, data []byte, config interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockECService) Decode(ctx context.Context, shards interface{}, config interface{}) ([]byte, error) {
	return nil, nil
}

func (m *MockECService) VerifyShards(ctx context.Context, shards interface{}) error {
	return nil
}

func (m *MockECService) CalculateCapacity(config interface{}, shardSize int) int {
	return 0
}

func (m *MockECService) OptimizeConfiguration(ctx context.Context, dataSize int, redundancy interface{}) (interface{}, error) {
	return nil, nil
}

type MockMediaService struct{}

func (m *MockMediaService) LoadMedia(ctx context.Context, data []byte, format interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockMediaService) DetectFormat(ctx context.Context, data []byte) (interface{}, error) {
	return nil, nil
}

func (m *MockMediaService) CalculateCapacity(ctx context.Context, asset interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockMediaService) SanitizeMetadata(ctx context.Context, asset interface{}) error {
	return nil
}

func (m *MockMediaService) ValidateMedia(ctx context.Context, asset interface{}) error {
	return nil
}

func (m *MockMediaService) AnalyzeQuality(ctx context.Context, asset interface{}) (float64, error) {
	return 0, nil
}

type MockStegoService struct{}

func (m *MockStegoService) CalculateCapacity(ctx context.Context, media []byte, technique stego.StegoTechnique) (int64, error) {
	// Return different capacities based on media size
	capacity := int64(len(media)) / 8 // Assume 1 bit per byte (LSB)
	return capacity, nil
}

func (m *MockStegoService) Embed(ctx context.Context, payload []byte, cover []byte, technique stego.StegoTechnique) ([]byte, error) {
	return nil, nil
}

func (m *MockStegoService) Extract(ctx context.Context, stego []byte, technique stego.StegoTechnique) ([]byte, error) {
	return nil, nil
}

func (m *MockStegoService) AnalyzeQuality(ctx context.Context, stego []byte) (interface{}, error) {
	return nil, nil
}

func TestCalculateOptimalSharding(t *testing.T) {
	tests := []struct {
		name             string
		dataSize         int64
		targetCount      int
		wantDataShards   int
		wantParityShards int
		wantErr          bool
	}{
		{
			name:             "one_target",
			dataSize:         1024,
			targetCount:      1,
			wantDataShards:   1,
			wantParityShards: 0,
			wantErr:          false,
		},
		{
			name:             "small_data_3_targets",
			dataSize:         500, // < 1KB
			targetCount:      3,
			wantDataShards:   1, // 50/50 split for small data
			wantParityShards: 2,
			wantErr:          false,
		},
		{
			name:             "medium_data_10_targets",
			dataSize:         50 * 1024, // 50KB
			targetCount:      10,
			wantDataShards:   6, // 60/40 split (2/3 for data)
			wantParityShards: 4,
			wantErr:          false,
		},
		{
			name:             "large_data_15_targets",
			dataSize:         150 * 1024 * 1024, // 150MB
			targetCount:      15,
			wantDataShards:   11, // 75/25 split for large data
			wantParityShards: 4,
			wantErr:          false,
		},
		{
			name:             "two_targets",
			dataSize:         10 * 1024,
			targetCount:      2,
			wantDataShards:   1,
			wantParityShards: 1,
			wantErr:          false,
		},
		{
			name:        "invalid_data_size",
			dataSize:    0,
			targetCount: 5,
			wantErr:     true,
		},
		{
			name:        "invalid_target_count",
			dataSize:    1024,
			targetCount: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := infraDist.NewDistributionService(
				&MockECService{},
				&MockMediaService{},
				&MockStegoService{},
				NewMockRepository(),
				nil,
			)

			dataShards, parityShards, err := service.CalculateOptimalSharding(
				context.Background(),
				tt.dataSize,
				tt.targetCount,
			)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantDataShards, dataShards, "data shards mismatch")
			assert.Equal(t, tt.wantParityShards, parityShards, "parity shards mismatch")
			assert.Equal(t, tt.targetCount, dataShards+parityShards, "total shards should equal target count")
		})
	}
}

func TestCreateStrategy(t *testing.T) {
	repo := NewMockRepository()
	service := infraDist.NewDistributionService(
		&MockECService{},
		&MockMediaService{},
		&MockStegoService{},
		repo,
		nil,
	)
	ctx := context.Background()

	tests := []struct {
		name           string
		pattern        domainDist.PatternType
		dataShards     int
		parityShards   int
		targetMediaIDs []string
		wantErr        bool
	}{
		{
			name:           "valid_one_to_many",
			pattern:        domainDist.PatternOneToMany,
			dataShards:     10,
			parityShards:   5,
			targetMediaIDs: []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8", "m9", "m10", "m11", "m12", "m13", "m14", "m15"},
			wantErr:        false,
		},
		{
			name:           "valid_one_to_one",
			pattern:        domainDist.PatternOneToOne,
			dataShards:     1,
			parityShards:   0,
			targetMediaIDs: []string{"m1"},
			wantErr:        false,
		},
		{
			name:           "insufficient_media",
			pattern:        domainDist.PatternOneToMany,
			dataShards:     10,
			parityShards:   5,
			targetMediaIDs: []string{"m1", "m2"}, // Need 15, have 2
			wantErr:        true,
		},
		{
			name:           "invalid_pattern",
			pattern:        domainDist.PatternType("invalid"),
			dataShards:     5,
			parityShards:   2,
			targetMediaIDs: []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7"},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, err := service.CreateStrategy(
				ctx,
				tt.pattern,
				tt.dataShards,
				tt.parityShards,
				tt.targetMediaIDs,
			)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, strategy)
			assert.Equal(t, tt.pattern, strategy.Pattern)
			assert.Equal(t, tt.dataShards, strategy.DataShards)
			assert.Equal(t, tt.parityShards, strategy.ParityShards)
			assert.Equal(t, tt.dataShards+tt.parityShards, strategy.TotalShards)
			assert.Equal(t, domainDist.StatusPending, strategy.Status)

			// Verify it was saved to repository
			retrieved, err := repo.GetStrategy(ctx, strategy.ID)
			require.NoError(t, err)
			assert.Equal(t, strategy.ID, retrieved.ID)
		})
	}
}

func TestCapacityAwareDistribution(t *testing.T) {
	service := infraDist.NewDistributionService(
		&MockECService{},
		&MockMediaService{},
		&MockStegoService{},
		NewMockRepository(),
		nil,
	)
	ctx := context.Background()

	tests := []struct {
		name          string
		targetMedia   [][]byte
		techniques    []stego.StegoTechnique
		totalDataSize int64
		dataShards    int
		wantErr       bool
		checkFunc     func(*testing.T, []infraDist.ShardAllocation)
	}{
		{
			name: "equal_capacity_3_media",
			targetMedia: [][]byte{
				make([]byte, 10000), // 10KB -> capacity ~1250 bytes
				make([]byte, 10000),
				make([]byte, 10000),
			},
			techniques: []stego.StegoTechnique{
				stego.LSB,
				stego.LSB,
				stego.LSB,
			},
			totalDataSize: 3000, // 3KB total
			dataShards:    3,
			wantErr:       false,
			checkFunc: func(t *testing.T, allocs []infraDist.ShardAllocation) {
				assert.Len(t, allocs, 3)
				// Should distribute evenly
				totalAllocated := 0
				for _, alloc := range allocs {
					assert.True(t, alloc.ShardCount >= 0)
					totalAllocated += alloc.ShardCount
				}
				assert.Equal(t, 3, totalAllocated, "all shards should be allocated")
			},
		},
		{
			name: "varying_capacity_3_media",
			targetMedia: [][]byte{
				make([]byte, 20000), // Large: ~2500 bytes capacity
				make([]byte, 10000), // Medium: ~1250 bytes capacity
				make([]byte, 5000),  // Small: ~625 bytes capacity
			},
			techniques: []stego.StegoTechnique{
				stego.LSB,
				stego.LSB,
				stego.LSB,
			},
			totalDataSize: 3000,
			dataShards:    6,
			wantErr:       false,
			checkFunc: func(t *testing.T, allocs []infraDist.ShardAllocation) {
				assert.Len(t, allocs, 3)
				// Larger capacity media should get more shards
				totalAllocated := 0
				for i, alloc := range allocs {
					assert.True(t, alloc.Utilization <= 0.9, "media %d utilization should not exceed 90%%", i)
					totalAllocated += alloc.ShardCount
				}
				assert.Equal(t, 6, totalAllocated, "all 6 shards should be allocated")
				// First media (largest) should get most shards
				assert.True(t, allocs[0].ShardCount >= allocs[1].ShardCount)
				assert.True(t, allocs[1].ShardCount >= allocs[2].ShardCount)
			},
		},
		{
			name: "insufficient_total_capacity",
			targetMedia: [][]byte{
				make([]byte, 1000), // Only 125 bytes capacity each
				make([]byte, 1000),
			},
			techniques: []stego.StegoTechnique{
				stego.LSB,
				stego.LSB,
			},
			totalDataSize: 5000, // Need 5KB, only have ~250 bytes
			dataShards:    5,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allocations, err := service.CalculateCapacityAwareDistribution(
				ctx,
				tt.targetMedia,
				tt.techniques,
				tt.totalDataSize,
				tt.dataShards,
			)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, allocations)
			if tt.checkFunc != nil {
				tt.checkFunc(t, allocations)
			}
		})
	}
}

func TestDistributeShards(t *testing.T) {
	repo := NewMockRepository()
	service := infraDist.NewDistributionService(
		&MockECService{},
		&MockMediaService{},
		&MockStegoService{},
		repo,
		nil,
	)
	ctx := context.Background()

	// First create a strategy
	strategy, err := service.CreateStrategy(
		ctx,
		domainDist.PatternOneToMany,
		3, // 3 data shards
		2, // 2 parity shards
		[]string{"m1", "m2", "m3", "m4", "m5"},
	)
	require.NoError(t, err)

	// Create shard data
	shardData := [][]byte{
		[]byte("shard-0-data"),
		[]byte("shard-1-data"),
		[]byte("shard-2-data"),
		[]byte("shard-3-parity"),
		[]byte("shard-4-parity"),
	}

	// Distribute shards
	manifest, err := service.DistributeShards(ctx, strategy.ID, shardData)
	require.NoError(t, err)
	assert.NotNil(t, manifest)
	assert.Equal(t, 5, manifest.TotalShards)
	assert.Equal(t, 3, manifest.RequiredShards) // Need 3 data shards to recover

	// Verify each shard metadata
	assert.Len(t, manifest.ShardMetadata, 5)
	for i, meta := range manifest.ShardMetadata {
		assert.Equal(t, i, meta.Index)
		assert.Greater(t, meta.Size, int64(0))
		assert.NotEmpty(t, meta.Checksum)
		assert.NotEmpty(t, meta.MediaID)
	}

	// Verify strategy was updated to complete
	updatedStrategy, err := repo.GetStrategy(ctx, strategy.ID)
	require.NoError(t, err)
	assert.Equal(t, domainDist.StatusComplete, updatedStrategy.Status)
	assert.NotNil(t, updatedStrategy.CompletedAt)
}

func TestCheckRecoverability(t *testing.T) {
	repo := NewMockRepository()
	service := infraDist.NewDistributionService(
		&MockECService{},
		&MockMediaService{},
		&MockStegoService{},
		repo,
		nil,
	)
	ctx := context.Background()

	// Create strategy and distribute shards
	strategy, _ := service.CreateStrategy(
		ctx,
		domainDist.PatternOneToMany,
		10, // 10 data shards
		5,  // 5 parity shards
		[]string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8", "m9", "m10", "m11", "m12", "m13", "m14", "m15"},
	)

	shardData := make([][]byte, 15)
	for i := 0; i < 15; i++ {
		shardData[i] = []byte(fmt.Sprintf("shard-%d", i))
	}
	manifest, _ := service.DistributeShards(ctx, strategy.ID, shardData)

	tests := []struct {
		name             string
		availableIndices []int
		wantRecoverable  bool
	}{
		{
			name:             "all_shards_available",
			availableIndices: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14},
			wantRecoverable:  true,
		},
		{
			name:             "exactly_threshold_shards",
			availableIndices: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, // Exactly 10
			wantRecoverable:  true,
		},
		{
			name:             "above_threshold",
			availableIndices: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, // 12 shards
			wantRecoverable:  true,
		},
		{
			name:             "below_threshold",
			availableIndices: []int{0, 1, 2, 3, 4, 5, 6, 7, 8}, // Only 9 shards
			wantRecoverable:  false,
		},
		{
			name:             "no_shards",
			availableIndices: []int{},
			wantRecoverable:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canRecover, err := service.CheckRecoverability(
				ctx,
				manifest.ID,
				tt.availableIndices,
			)
			require.NoError(t, err)
			assert.Equal(t, tt.wantRecoverable, canRecover)
		})
	}
}
