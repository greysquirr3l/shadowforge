package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
)

// TestAllocateRoundRobin verifies round-robin allocation distributes shards evenly
func TestAllocateRoundRobin(t *testing.T) {
	handler := &EmbedMatrixHandler{}

	// Create 3 payloads with 10 shards each = 30 total shards
	payloadShards := make(map[string]*PayloadShardSet)
	for i := 0; i < 3; i++ {
		name := []string{"payload1", "payload2", "payload3"}[i]
		shards := make([]*errorcorrection.Shard, 10)
		for j := 0; j < 10; j++ {
			shards[j] = &errorcorrection.Shard{
				Index: j,
				Data:  make([]byte, 100),
			}
		}
		payloadShards[name] = &PayloadShardSet{
			PayloadName: name,
			Shards:      shards,
		}
	}

	// Create 5 covers with large capacity
	coverCapacities := []CoverCapacityInfo{
		{Path: "/cover1.png", Capacity: 10000, Technique: "lsb"},
		{Path: "/cover2.png", Capacity: 10000, Technique: "lsb"},
		{Path: "/cover3.png", Capacity: 10000, Technique: "lsb"},
		{Path: "/cover4.png", Capacity: 10000, Technique: "lsb"},
		{Path: "/cover5.png", Capacity: 10000, Technique: "lsb"},
	}

	// Execute allocation
	allocation, err := handler.allocateRoundRobin(payloadShards, coverCapacities)

	// Verify success
	require.NoError(t, err)
	require.NotNil(t, allocation)

	// Count total allocated shards
	totalAllocated := 0
	for _, coverAlloc := range allocation.CoverAllocations {
		totalAllocated += len(coverAlloc.Shards)
	}
	assert.Equal(t, 30, totalAllocated, "All 30 shards should be allocated")

	// Verify even distribution (should be 6 shards per cover)
	for coverPath, coverAlloc := range allocation.CoverAllocations {
		assert.InDelta(t, 6, len(coverAlloc.Shards), 1,
			"Cover %s should have ~6 shards (round-robin)", coverPath)
	}
}

// TestAllocateRandom verifies random allocation respects capacity constraints
func TestAllocateRandom(t *testing.T) {
	handler := &EmbedMatrixHandler{}

	// Create 2 payloads with 8 shards each = 16 total shards
	payloadShards := make(map[string]*PayloadShardSet)
	for i := 0; i < 2; i++ {
		name := []string{"payload1", "payload2"}[i]
		shards := make([]*errorcorrection.Shard, 8)
		for j := 0; j < 8; j++ {
			shards[j] = &errorcorrection.Shard{
				Index: j,
				Data:  make([]byte, 100),
			}
		}
		payloadShards[name] = &PayloadShardSet{
			PayloadName: name,
			Shards:      shards,
		}
	}

	// Create 4 covers with sufficient capacity
	coverCapacities := []CoverCapacityInfo{
		{Path: "/cover1.png", Capacity: 10000, Technique: "lsb"},
		{Path: "/cover2.png", Capacity: 10000, Technique: "lsb"},
		{Path: "/cover3.png", Capacity: 10000, Technique: "lsb"},
		{Path: "/cover4.png", Capacity: 10000, Technique: "lsb"},
	}

	// Execute allocation
	allocation, err := handler.allocateRandom(payloadShards, coverCapacities)

	// Verify success
	require.NoError(t, err)
	require.NotNil(t, allocation)

	// Count total allocated shards
	totalAllocated := 0
	for _, coverAlloc := range allocation.CoverAllocations {
		totalAllocated += len(coverAlloc.Shards)
		// Verify capacity not exceeded
		assert.LessOrEqual(t, coverAlloc.TotalSize, coverAlloc.Capacity,
			"Capacity should not be exceeded")
	}
	assert.Equal(t, 16, totalAllocated, "All 16 shards should be allocated")
}

// TestAllocateOptimized verifies optimized allocation maximizes space efficiency
func TestAllocateOptimized(t *testing.T) {
	handler := &EmbedMatrixHandler{}

	// Create 1 payload with 15 shards
	payloadShards := make(map[string]*PayloadShardSet)
	shards := make([]*errorcorrection.Shard, 15)
	for j := 0; j < 15; j++ {
		shards[j] = &errorcorrection.Shard{
			Index: j,
			Data:  make([]byte, 100),
		}
	}
	payloadShards["payload1"] = &PayloadShardSet{
		PayloadName: "payload1",
		Shards:      shards,
	}

	// Create 3 covers with varied capacities (large → small)
	coverCapacities := []CoverCapacityInfo{
		{Path: "/cover1.png", Capacity: 10000, Technique: "lsb"}, // Largest
		{Path: "/cover2.png", Capacity: 5000, Technique: "lsb"},  // Medium
		{Path: "/cover3.png", Capacity: 2000, Technique: "lsb"},  // Smallest
	}

	// Execute allocation
	allocation, err := handler.allocateOptimized(payloadShards, coverCapacities)

	// Verify success
	require.NoError(t, err)
	require.NotNil(t, allocation)

	// Count total allocated shards
	totalAllocated := 0
	for _, coverAlloc := range allocation.CoverAllocations {
		totalAllocated += len(coverAlloc.Shards)
	}
	assert.Equal(t, 15, totalAllocated, "All 15 shards should be allocated")

	// Verify first-fit decreasing (largest cover should have most shards)
	cover1Shards := len(allocation.CoverAllocations["/cover1.png"].Shards)
	cover3Shards := len(allocation.CoverAllocations["/cover3.png"].Shards)
	assert.GreaterOrEqual(t, cover1Shards, cover3Shards,
		"Largest cover should have at least as many shards as smallest")
}

// TestAllocateBalanced verifies balanced allocation distributes load evenly
func TestAllocateBalanced(t *testing.T) {
	handler := &EmbedMatrixHandler{}

	// Create 2 payloads with 10 shards each = 20 total shards
	payloadShards := make(map[string]*PayloadShardSet)
	for i := 0; i < 2; i++ {
		name := []string{"payload1", "payload2"}[i]
		shards := make([]*errorcorrection.Shard, 10)
		for j := 0; j < 10; j++ {
			shards[j] = &errorcorrection.Shard{
				Index: j,
				Data:  make([]byte, 100),
			}
		}
		payloadShards[name] = &PayloadShardSet{
			PayloadName: name,
			Shards:      shards,
		}
	}

	// Create 4 covers with equal capacities
	coverCapacities := []CoverCapacityInfo{
		{Path: "/cover1.png", Capacity: 5000, Technique: "lsb"},
		{Path: "/cover2.png", Capacity: 5000, Technique: "lsb"},
		{Path: "/cover3.png", Capacity: 5000, Technique: "lsb"},
		{Path: "/cover4.png", Capacity: 5000, Technique: "lsb"},
	}

	// Execute allocation
	allocation, err := handler.allocateBalanced(payloadShards, coverCapacities)

	// Verify success
	require.NoError(t, err)
	require.NotNil(t, allocation)

	// Count total allocated shards
	totalAllocated := 0
	for _, coverAlloc := range allocation.CoverAllocations {
		totalAllocated += len(coverAlloc.Shards)
	}
	assert.Equal(t, 20, totalAllocated, "All 20 shards should be allocated")

	// Verify balanced distribution (each cover should have 5 shards)
	for coverPath, coverAlloc := range allocation.CoverAllocations {
		assert.InDelta(t, 5, len(coverAlloc.Shards), 1,
			"Cover %s should have ~5 shards (balanced)", coverPath)
	}
}

// TestAllocationInsufficientCapacity verifies error handling for insufficient capacity
func TestAllocationInsufficientCapacity(t *testing.T) {
	handler := &EmbedMatrixHandler{}

	// Create 5 payloads with 10 shards each = 50 total shards
	payloadShards := make(map[string]*PayloadShardSet)
	for i := 0; i < 5; i++ {
		name := []string{"payload1", "payload2", "payload3", "payload4", "payload5"}[i]
		shards := make([]*errorcorrection.Shard, 10)
		for j := 0; j < 10; j++ {
			shards[j] = &errorcorrection.Shard{
				Index: j,
				Data:  make([]byte, 100), // 100 bytes per shard = 5000 bytes total
			}
		}
		payloadShards[name] = &PayloadShardSet{
			PayloadName: name,
			Shards:      shards,
		}
	}

	// Create 2 covers with insufficient capacity (only 200 bytes total)
	coverCapacities := []CoverCapacityInfo{
		{Path: "/cover1.png", Capacity: 100, Technique: "lsb"},
		{Path: "/cover2.png", Capacity: 100, Technique: "lsb"},
	}

	// Execute allocation - should fail for all modes
	modes := []struct {
		name     string
		allocate func() (*MatrixAllocation, error)
	}{
		{"round_robin", func() (*MatrixAllocation, error) {
			return handler.allocateRoundRobin(payloadShards, coverCapacities)
		}},
		{"random", func() (*MatrixAllocation, error) {
			return handler.allocateRandom(payloadShards, coverCapacities)
		}},
		{"optimized", func() (*MatrixAllocation, error) {
			return handler.allocateOptimized(payloadShards, coverCapacities)
		}},
		{"balanced", func() (*MatrixAllocation, error) {
			return handler.allocateBalanced(payloadShards, coverCapacities)
		}},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			_, err := mode.allocate()
			assert.Error(t, err, "Mode %s should fail with insufficient capacity", mode.name)
		})
	}
}

// TestAllocationModeComparison compares behavior of all allocation modes
func TestAllocationModeComparison(t *testing.T) {
	handler := &EmbedMatrixHandler{}

	// Create 3 payloads with 10 shards each = 30 total shards
	payloadShards := make(map[string]*PayloadShardSet)
	for i := 0; i < 3; i++ {
		name := []string{"payload1", "payload2", "payload3"}[i]
		shards := make([]*errorcorrection.Shard, 10)
		for j := 0; j < 10; j++ {
			shards[j] = &errorcorrection.Shard{
				Index: j,
				Data:  make([]byte, 100),
			}
		}
		payloadShards[name] = &PayloadShardSet{
			PayloadName: name,
			Shards:      shards,
		}
	}

	// Create 5 covers with large capacity
	coverCapacities := []CoverCapacityInfo{
		{Path: "/cover1.png", Capacity: 10000, Technique: "lsb"},
		{Path: "/cover2.png", Capacity: 10000, Technique: "lsb"},
		{Path: "/cover3.png", Capacity: 10000, Technique: "lsb"},
		{Path: "/cover4.png", Capacity: 10000, Technique: "lsb"},
		{Path: "/cover5.png", Capacity: 10000, Technique: "lsb"},
	}

	// Test all modes allocate all shards
	modes := []struct {
		name     string
		allocate func() (*MatrixAllocation, error)
	}{
		{"round_robin", func() (*MatrixAllocation, error) {
			return handler.allocateRoundRobin(payloadShards, coverCapacities)
		}},
		{"random", func() (*MatrixAllocation, error) {
			return handler.allocateRandom(payloadShards, coverCapacities)
		}},
		{"optimized", func() (*MatrixAllocation, error) {
			return handler.allocateOptimized(payloadShards, coverCapacities)
		}},
		{"balanced", func() (*MatrixAllocation, error) {
			return handler.allocateBalanced(payloadShards, coverCapacities)
		}},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			allocation, err := mode.allocate()

			require.NoError(t, err, "Mode %s should succeed", mode.name)
			require.NotNil(t, allocation)

			// Verify all 30 shards allocated
			totalAllocated := 0
			for _, coverAlloc := range allocation.CoverAllocations {
				totalAllocated += len(coverAlloc.Shards)
				// Verify capacity not exceeded
				assert.LessOrEqual(t, coverAlloc.TotalSize, coverAlloc.Capacity,
					"Mode %s should not exceed capacity", mode.name)
			}
			assert.Equal(t, 30, totalAllocated, "Mode %s should allocate all 30 shards", mode.name)

			t.Logf("Mode %s: Total allocated = %d bytes across %d covers",
				mode.name, allocation.TotalAllocatedSize, len(allocation.CoverAllocations))
		})
	}
}
