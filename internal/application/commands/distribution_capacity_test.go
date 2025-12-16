package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapacityPlanner_AnalyzeCapacity(t *testing.T) {
	planner := NewCapacityPlanner()

	tests := []struct {
		name            string
		payloadSizes    []int64
		coverCapacities []CoverCapacityInfo
		rsRedundancy    float64
		expectFeasible  bool
		expectWarnings  bool
		description     string
	}{
		{
			name:         "sufficient_capacity_single_payload",
			payloadSizes: []int64{10240}, // 10KB
			coverCapacities: []CoverCapacityInfo{
				{Path: "/cover1.png", Capacity: 50000},
				{Path: "/cover2.png", Capacity: 50000},
				{Path: "/cover3.png", Capacity: 50000},
			},
			rsRedundancy:   0.3,
			expectFeasible: true,
			expectWarnings: false,
			description:    "Should be feasible with 150KB total for 10KB payload",
		},
		{
			name:         "insufficient_capacity",
			payloadSizes: []int64{100000}, // 100KB
			coverCapacities: []CoverCapacityInfo{
				{Path: "/cover1.png", Capacity: 10000},
				{Path: "/cover2.png", Capacity: 10000},
			},
			rsRedundancy:   0.3,
			expectFeasible: false,
			expectWarnings: true,
			description:    "Should fail with only 20KB for 100KB payload",
		},
		{
			name:         "multiple_payloads",
			payloadSizes: []int64{5120, 5120, 5120}, // 3x 5KB = 15KB
			coverCapacities: []CoverCapacityInfo{
				{Path: "/cover1.png", Capacity: 30000},
				{Path: "/cover2.png", Capacity: 30000},
				{Path: "/cover3.png", Capacity: 30000},
				{Path: "/cover4.png", Capacity: 30000},
			},
			rsRedundancy:   0.3,
			expectFeasible: true,
			expectWarnings: false,
			description:    "Should handle multiple small payloads",
		},
		{
			name:         "high_utilization",
			payloadSizes: []int64{15000}, // 15KB
			coverCapacities: []CoverCapacityInfo{
				{Path: "/cover1.png", Capacity: 30000}, // Will use ~65% (15KB * 1.3 / 30KB)
			},
			rsRedundancy:   0.3,
			expectFeasible: true,
			expectWarnings: false, // Adjust: feasible but somewhat tight
			description:    "Should be feasible with moderate utilization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			analysis, err := planner.AnalyzeCapacity(ctx, tt.payloadSizes, tt.coverCapacities, tt.rsRedundancy)

			require.NoError(t, err, tt.description)
			require.NotNil(t, analysis)

			assert.Equal(t, tt.expectFeasible, analysis.IsFeasible,
				"%s: feasibility mismatch", tt.description)

			if tt.expectWarnings {
				assert.NotEmpty(t, analysis.Warnings,
					"%s: expected warnings", tt.description)
			}

			// Verify basic calculations
			assert.Greater(t, analysis.TotalCapacity, int64(0))
			assert.Greater(t, analysis.RequiredCapacity, int64(0))
			assert.LessOrEqual(t, analysis.UsableCapacity, analysis.TotalCapacity)
			assert.Equal(t, len(tt.coverCapacities), len(analysis.CoverBreakdown))

			t.Logf("%s: Utilization=%.1f%%, Feasible=%v, Warnings=%d",
				tt.name, analysis.UtilizationPercent, analysis.IsFeasible, len(analysis.Warnings))
		})
	}
}

func TestCapacityPlanner_RecommendSharding(t *testing.T) {
	planner := NewCapacityPlanner()

	tests := []struct {
		name            string
		payloadSize     int64
		coverCount      int
		rsRedundancy    float64
		minDataShards   int
		minParityShards int
		description     string
	}{
		{
			name:            "small_payload_few_covers",
			payloadSize:     10240, // 10KB
			coverCount:      3,
			rsRedundancy:    0.3,
			minDataShards:   1,
			minParityShards: 1,
			description:     "Should handle small payload with minimal sharding",
		},
		{
			name:            "medium_payload_many_covers",
			payloadSize:     102400, // 100KB
			coverCount:      10,
			rsRedundancy:    0.3,
			minDataShards:   5,
			minParityShards: 1,
			description:     "Should use more shards with more covers",
		},
		{
			name:            "high_redundancy",
			payloadSize:     51200, // 50KB
			coverCount:      8,
			rsRedundancy:    0.5, // 50% redundancy
			minDataShards:   3,
			minParityShards: 1,
			description:     "Should increase parity shards with high redundancy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := planner.recommendSharding(tt.payloadSize, tt.coverCount, tt.rsRedundancy)

			assert.GreaterOrEqual(t, rec.DataShards, tt.minDataShards,
				"%s: too few data shards", tt.description)
			assert.GreaterOrEqual(t, rec.ParityShards, tt.minParityShards,
				"%s: too few parity shards", tt.description)
			assert.Equal(t, rec.DataShards+rec.ParityShards, rec.TotalShards,
				"%s: total shards mismatch", tt.description)
			assert.Equal(t, rec.DataShards, rec.MinShardsNeeded,
				"%s: min shards should equal data shards", tt.description)
			assert.Equal(t, rec.ParityShards, rec.MaxLossableCover,
				"%s: max lossable should equal parity shards", tt.description)

			t.Logf("%s: %dD+%dP=%dT shards, shard_size=%dKB",
				tt.name, rec.DataShards, rec.ParityShards, rec.TotalShards,
				rec.ShardSize/1024)
		})
	}
}

func TestCapacityPlanner_CalculateDistributionFeasibility(t *testing.T) {
	planner := NewCapacityPlanner()

	covers := []CoverCapacityInfo{
		{Path: "/cover1.png", Capacity: 50000},
		{Path: "/cover2.png", Capacity: 50000},
		{Path: "/cover3.png", Capacity: 50000},
	}

	tests := []struct {
		name           string
		payloadCount   int
		avgPayloadSize int64
		covers         []CoverCapacityInfo
		pattern        string
		expectFeasible bool
		description    string
	}{
		{
			name:           "one_to_many_feasible",
			payloadCount:   1,
			avgPayloadSize: 10000,
			covers:         covers,
			pattern:        "one-to-many",
			expectFeasible: true,
			description:    "Should be feasible with 3 covers for 10KB",
		},
		{
			name:           "one_to_many_insufficient_covers",
			payloadCount:   1,
			avgPayloadSize: 10000,
			covers:         covers[:1], // Only 1 cover
			pattern:        "one-to-many",
			expectFeasible: false,
			description:    "Should fail with insufficient covers",
		},
		{
			name:           "many_to_one_feasible",
			payloadCount:   3,
			avgPayloadSize: 10000,
			covers:         covers,
			pattern:        "many-to-one",
			expectFeasible: true,
			description:    "Should be feasible with 50KB cover for 3×10KB",
		},
		{
			name:           "many_to_many_feasible",
			payloadCount:   3,
			avgPayloadSize: 10000,
			covers:         covers,
			pattern:        "many-to-many",
			expectFeasible: true,
			description:    "Should be feasible for matrix pattern",
		},
		{
			name:           "many_to_many_insufficient_payloads",
			payloadCount:   1, // Need at least 2
			avgPayloadSize: 10000,
			covers:         covers,
			pattern:        "many-to-many",
			expectFeasible: false,
			description:    "Should fail with only 1 payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feasible, reason := planner.CalculateDistributionFeasibility(
				tt.payloadCount, tt.avgPayloadSize, tt.covers, tt.pattern)

			assert.Equal(t, tt.expectFeasible, feasible,
				"%s: %s", tt.description, reason)

			t.Logf("%s: feasible=%v, reason=%s", tt.name, feasible, reason)
		})
	}
}

func TestCapacityPlanner_PredictUtilization(t *testing.T) {
	planner := NewCapacityPlanner()

	payloadSizes := []int64{10240, 10240} // 2× 10KB = 20KB
	covers := []CoverCapacityInfo{
		{Path: "/cover1.png", Capacity: 50000},
		{Path: "/cover2.png", Capacity: 50000},
		{Path: "/cover3.png", Capacity: 50000},
	}

	modes := []MatrixMode{
		MatrixModeRoundRobin,
		MatrixModeBalanced,
		MatrixModeOptimized,
		MatrixModeRandom,
	}

	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			utilization := planner.PredictUtilization(payloadSizes, covers, mode)

			require.Len(t, utilization, len(covers),
				"Should predict utilization for all covers")

			for path, util := range utilization {
				assert.GreaterOrEqual(t, util, 0.0,
					"Utilization for %s should be non-negative", path)
				assert.LessOrEqual(t, util, 200.0,
					"Utilization for %s should be reasonable", path)

				t.Logf("Mode %s: %s utilization = %.1f%%", mode, path, util)
			}
		})
	}
}
