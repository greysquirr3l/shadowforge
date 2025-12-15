package commands

import (
	"context"
	"fmt"
	"math"
)

// CapacityPlanner provides utilities for planning and analyzing distribution capacity.
type CapacityPlanner struct{}

// NewCapacityPlanner creates a new capacity planner.
func NewCapacityPlanner() *CapacityPlanner {
	return &CapacityPlanner{}
}

// CapacityAnalysis represents the result of capacity analysis.
type CapacityAnalysis struct {
	TotalCapacity       int64                  `json:"total_capacity"`       // Total available capacity (bytes)
	RequiredCapacity    int64                  `json:"required_capacity"`    // Required capacity for payload(s)
	UsableCapacity      int64                  `json:"usable_capacity"`      // Capacity after safety margin (90%)
	OverheadCapacity    int64                  `json:"overhead_capacity"`    // Capacity needed for RS + metadata
	NetCapacity         int64                  `json:"net_capacity"`         // Usable - Overhead
	IsFeasible          bool                   `json:"is_feasible"`          // Can payloads fit?
	UtilizationPercent  float64                `json:"utilization_percent"`  // Expected utilization %
	RecommendedSharding ShardingRecommendation `json:"recommended_sharding"` // Optimal shard configuration
	Warnings            []string               `json:"warnings,omitempty"`   // Capacity warnings
	CoverBreakdown      []CoverCapacity        `json:"cover_breakdown"`      // Per-cover capacity details
}

// ShardingRecommendation suggests optimal Reed-Solomon configuration.
type ShardingRecommendation struct {
	DataShards       int     `json:"data_shards"`        // Recommended data shards
	ParityShards     int     `json:"parity_shards"`      // Recommended parity shards
	TotalShards      int     `json:"total_shards"`       // Data + Parity
	RedundancyLevel  float64 `json:"redundancy_level"`   // Parity/Data ratio
	ShardSize        int64   `json:"shard_size"`         // Approx bytes per shard
	MinShardsNeeded  int     `json:"min_shards_needed"`  // K for K-of-N recovery
	MaxLossableCover int     `json:"max_lossable_cover"` // How many covers can be lost
}

// CoverCapacity describes capacity details for a single cover.
type CoverCapacity struct {
	Path            string  `json:"path"`
	RawCapacity     int64   `json:"raw_capacity"`    // Theoretical maximum
	SafeCapacity    int64   `json:"safe_capacity"`   // 90% of raw (recommended)
	Technique       string  `json:"technique"`       // Stego technique
	QualityScore    float64 `json:"quality_score"`   // Media quality (0-1)
	RecommendedUse  bool    `json:"recommended_use"` // Should this cover be used?
	CapacityWarning string  `json:"capacity_warning,omitempty"`
}

// AnalyzeCapacity analyzes total capacity across all covers and provides recommendations.
func (cp *CapacityPlanner) AnalyzeCapacity(
	ctx context.Context,
	payloadSizes []int64,
	coverCapacities []CoverCapacityInfo,
	rsRedundancy float64,
) (*CapacityAnalysis, error) {
	if len(payloadSizes) == 0 {
		return nil, fmt.Errorf("no payloads provided for analysis")
	}
	if len(coverCapacities) == 0 {
		return nil, fmt.Errorf("no cover media provided for analysis")
	}
	if rsRedundancy < 0.1 || rsRedundancy > 0.9 {
		return nil, fmt.Errorf("rs_redundancy must be between 0.1 and 0.9, got: %.2f", rsRedundancy)
	}

	// Calculate total payload size
	var totalPayloadSize int64
	for _, size := range payloadSizes {
		totalPayloadSize += size
	}

	// Calculate total raw capacity
	var totalRawCapacity int64
	coverBreakdown := make([]CoverCapacity, len(coverCapacities))
	for i, cover := range coverCapacities {
		totalRawCapacity += cover.Capacity
		safeCapacity := int64(float64(cover.Capacity) * 0.9) // 90% safety margin

		coverBreakdown[i] = CoverCapacity{
			Path:           cover.Path,
			RawCapacity:    cover.Capacity,
			SafeCapacity:   safeCapacity,
			Technique:      cover.Technique,
			QualityScore:   1.0,                  // Default - would come from analysis
			RecommendedUse: safeCapacity >= 1024, // Minimum 1KB useful capacity
		}

		if cover.Capacity < 1024 {
			coverBreakdown[i].CapacityWarning = "Capacity too small for practical use"
			coverBreakdown[i].RecommendedUse = false
		}
	}

	usableCapacity := int64(float64(totalRawCapacity) * 0.9) // 90% safety margin

	// Calculate Reed-Solomon overhead
	// Formula: overhead = payloadSize * (parityShards / dataShards)
	// For rsRedundancy = 0.3, we need 30% extra capacity for parity
	overheadCapacity := int64(float64(totalPayloadSize) * rsRedundancy)

	// Add metadata overhead (manifest, shard headers, etc.)
	// Estimate: 256 bytes per shard + 1KB manifest per payload
	estimatedShardCount := cp.estimateShardCount(totalPayloadSize, len(coverCapacities))
	metadataOverhead := int64(estimatedShardCount*256 + len(payloadSizes)*1024)
	overheadCapacity += metadataOverhead

	// Calculate net capacity (what's actually available for data)
	netCapacity := usableCapacity - overheadCapacity
	requiredCapacity := totalPayloadSize

	// Check feasibility
	isFeasible := netCapacity >= requiredCapacity
	utilizationPercent := 0.0
	if usableCapacity > 0 {
		utilizationPercent = float64(requiredCapacity+overheadCapacity) / float64(usableCapacity) * 100
	}

	// Generate sharding recommendation
	shardingRec := cp.recommendSharding(totalPayloadSize, len(coverCapacities), rsRedundancy)

	// Generate warnings
	var warnings []string
	if !isFeasible {
		deficit := requiredCapacity - netCapacity
		warnings = append(warnings, fmt.Sprintf("Insufficient capacity: need %d more bytes", deficit))
	}
	if utilizationPercent > 80 {
		warnings = append(warnings, fmt.Sprintf("High utilization (%.1f%%) - consider adding more covers", utilizationPercent))
	}
	if len(coverCapacities) < 3 {
		warnings = append(warnings, "Low cover count - recommend at least 3-5 covers for distributed patterns")
	}
	if rsRedundancy < 0.2 {
		warnings = append(warnings, fmt.Sprintf("Low redundancy (%.0f%%) - minimal fault tolerance", rsRedundancy*100))
	}

	return &CapacityAnalysis{
		TotalCapacity:       totalRawCapacity,
		RequiredCapacity:    requiredCapacity,
		UsableCapacity:      usableCapacity,
		OverheadCapacity:    overheadCapacity,
		NetCapacity:         netCapacity,
		IsFeasible:          isFeasible,
		UtilizationPercent:  utilizationPercent,
		RecommendedSharding: shardingRec,
		Warnings:            warnings,
		CoverBreakdown:      coverBreakdown,
	}, nil
}

// recommendSharding suggests optimal Reed-Solomon configuration.
func (cp *CapacityPlanner) recommendSharding(
	payloadSize int64,
	coverCount int,
	rsRedundancy float64,
) ShardingRecommendation {
	// Target: distribute data across 60-80% of available covers for good balance
	targetCoverUsage := int(math.Ceil(float64(coverCount) * 0.7))
	if targetCoverUsage < 3 {
		targetCoverUsage = min(3, coverCount)
	}

	// Calculate data shards (use slightly fewer than covers for flexibility)
	dataShards := max(targetCoverUsage-2, 1)
	if dataShards > 15 {
		dataShards = 15 // Cap at 15 for reasonable performance
	}

	// Calculate parity shards based on redundancy level
	parityShards := int(math.Ceil(float64(dataShards) * rsRedundancy))
	if parityShards < 1 {
		parityShards = 1 // Minimum 1 parity shard
	}

	totalShards := dataShards + parityShards

	// Estimate shard size
	shardSize := payloadSize / int64(dataShards)
	if shardSize < 1024 {
		shardSize = 1024 // Minimum 1KB per shard
	}

	// Maximum lossable covers = parity shards (can lose up to P shards in K-of-N)
	maxLossable := parityShards

	return ShardingRecommendation{
		DataShards:       dataShards,
		ParityShards:     parityShards,
		TotalShards:      totalShards,
		RedundancyLevel:  float64(parityShards) / float64(dataShards),
		ShardSize:        shardSize,
		MinShardsNeeded:  dataShards,
		MaxLossableCover: maxLossable,
	}
}

// estimateShardCount estimates total shards across all payloads.
func (cp *CapacityPlanner) estimateShardCount(totalPayloadSize int64, coverCount int) int {
	// Use recommendSharding to get typical shard count
	rec := cp.recommendSharding(totalPayloadSize, coverCount, 0.3)
	return rec.TotalShards
}

// CalculateDistributionFeasibility checks if a specific distribution is feasible.
func (cp *CapacityPlanner) CalculateDistributionFeasibility(
	payloadCount int,
	avgPayloadSize int64,
	coverCapacities []CoverCapacityInfo,
	pattern string, // "one-to-many", "many-to-one", "many-to-many"
) (bool, string) {
	switch pattern {
	case "one-to-many":
		// Need enough covers to distribute shards
		if len(coverCapacities) < 3 {
			return false, "one-to-many requires at least 3 covers"
		}
		// Check total capacity
		var totalCap int64
		for _, c := range coverCapacities {
			totalCap += c.Capacity
		}
		required := avgPayloadSize * 13 / 10 // 30% overhead
		if totalCap < required {
			return false, fmt.Sprintf("insufficient capacity: have %d bytes, need %d bytes", totalCap, required)
		}
		return true, "feasible"

	case "many-to-one":
		// Need one cover large enough for all payloads
		totalRequired := avgPayloadSize * int64(payloadCount) * 13 / 10
		for _, c := range coverCapacities {
			if c.Capacity >= totalRequired {
				return true, "feasible - cover has sufficient capacity"
			}
		}
		return false, fmt.Sprintf("no single cover large enough for %d payloads", payloadCount)

	case "many-to-many":
		// Need matrix dimensions N×M where both > 1
		if payloadCount < 2 {
			return false, "many-to-many requires at least 2 payloads"
		}
		if len(coverCapacities) < 2 {
			return false, "many-to-many requires at least 2 covers"
		}
		// Check total capacity
		var totalCap int64
		for _, c := range coverCapacities {
			totalCap += c.Capacity
		}
		totalRequired := avgPayloadSize * int64(payloadCount) * 13 / 10
		if totalCap < totalRequired {
			return false, fmt.Sprintf("insufficient total capacity")
		}
		return true, "feasible"

	default:
		return false, fmt.Sprintf("unknown pattern: %s", pattern)
	}
}

// PredictUtilization predicts how capacity will be used across covers.
func (cp *CapacityPlanner) PredictUtilization(
	payloadSizes []int64,
	coverCapacities []CoverCapacityInfo,
	mode MatrixMode,
) map[string]float64 {
	utilization := make(map[string]float64)

	var totalPayload int64
	for _, size := range payloadSizes {
		totalPayload += size
	}

	// Add 30% overhead for RS + metadata
	totalRequired := totalPayload * 13 / 10

	switch mode {
	case MatrixModeRoundRobin, MatrixModeBalanced:
		// Even distribution
		perCover := float64(totalRequired) / float64(len(coverCapacities))
		for _, cover := range coverCapacities {
			if cover.Capacity > 0 {
				utilization[cover.Path] = (perCover / float64(cover.Capacity)) * 100
			}
		}

	case MatrixModeOptimized:
		// Largest covers get more
		var totalCap int64
		for _, c := range coverCapacities {
			totalCap += c.Capacity
		}
		for _, cover := range coverCapacities {
			proportion := float64(cover.Capacity) / float64(totalCap)
			allocated := proportion * float64(totalRequired)
			if cover.Capacity > 0 {
				utilization[cover.Path] = (allocated / float64(cover.Capacity)) * 100
			}
		}

	case MatrixModeRandom:
		// Assume roughly even with variance
		avgPerCover := float64(totalRequired) / float64(len(coverCapacities))
		for _, cover := range coverCapacities {
			// Random varies ±20%
			variance := avgPerCover * 0.2
			allocated := avgPerCover + (variance * (0.5 - 0.5)) // Simplified
			if cover.Capacity > 0 {
				utilization[cover.Path] = (allocated / float64(cover.Capacity)) * 100
			}
		}
	}

	return utilization
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
