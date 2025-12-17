package selection

import (
	"errors"
	"fmt"
	"sort"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

var (
	ErrInsufficientCapacity = errors.New("insufficient total capacity for payload")
	ErrNoValidFiles         = errors.New("no valid media files found")
)

// SelectionConstraints defines constraints for optimal cover selection.
type SelectionConstraints struct {
	MaxFiles          int     // Maximum number of files to use (0 = unlimited)
	MaxPerType        int     // Maximum files per media type (0 = unlimited)
	TargetUtilization float64 // Target capacity utilization (0.0-1.0, default 0.7)
	PreferDiversity   bool    // Prefer mixed media types
	MinimizeDetect    bool    // Minimize detectability score
}

// SelectionPlan represents the planned distribution of data across covers.
type SelectionPlan struct {
	SelectedFiles    []*CoverAllocation
	TotalCapacity    int64
	UsedCapacity     int64
	Utilization      float64
	AvgDetectability float64
	DiversityScore   float64
	Warnings         []string
}

// CoverAllocation represents how a specific cover file will be used.
type CoverAllocation struct {
	File             *MediaAnalysis
	Technique        stego.StegoTechnique
	AllocatedBytes   int64
	UtilizationRatio float64
	ShardIndices     []int // Which shards will be embedded (for distributed mode)
}

// OptimalSelector performs optimal cover selection based on constraints.
type OptimalSelector struct {
	constraints SelectionConstraints
}

// NewOptimalSelector creates a new optimal selector with given constraints.
func NewOptimalSelector(constraints SelectionConstraints) *OptimalSelector {
	// Set defaults
	if constraints.TargetUtilization == 0 {
		constraints.TargetUtilization = 0.7
	}
	return &OptimalSelector{constraints: constraints}
}

// SelectCovers performs optimal cover selection for a given payload size.
func (s *OptimalSelector) SelectCovers(
	available []*MediaAnalysis,
	payloadSize int64,
) (*SelectionPlan, error) {
	if len(available) == 0 {
		return nil, ErrNoValidFiles
	}

	// Sort files by capacity (descending) for greedy selection
	sorted := make([]*MediaAnalysis, len(available))
	copy(sorted, available)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SafeCapacity > sorted[j].SafeCapacity
	})

	// Calculate total available capacity
	totalCapacity := int64(0)
	for _, file := range sorted {
		totalCapacity += file.SafeCapacity
	}

	if totalCapacity < payloadSize {
		return nil, fmt.Errorf("%w: need %d bytes, have %d bytes",
			ErrInsufficientCapacity, payloadSize, totalCapacity)
	}

	// Greedy selection algorithm
	selected := make([]*CoverAllocation, 0)
	remaining := payloadSize
	typeCount := make(map[media.MediaType]int)

	for _, file := range sorted {
		// Check constraints
		if s.constraints.MaxFiles > 0 && len(selected) >= s.constraints.MaxFiles {
			break
		}

		if s.constraints.MaxPerType > 0 && typeCount[file.Type] >= s.constraints.MaxPerType {
			continue
		}

		// Calculate allocation
		allocSize := remaining
		if allocSize > file.SafeCapacity {
			allocSize = file.SafeCapacity
		}

		// Add to selection
		utilization := float64(allocSize) / float64(file.SafeCapacity)
		selected = append(selected, &CoverAllocation{
			File:             file,
			Technique:        file.RecommendedTechnique,
			AllocatedBytes:   allocSize,
			UtilizationRatio: utilization,
			ShardIndices:     nil, // Will be assigned later for distributed mode
		})

		typeCount[file.Type]++
		remaining -= allocSize

		if remaining <= 0 {
			break
		}
	}

	// Calculate metrics
	plan := &SelectionPlan{
		SelectedFiles: selected,
		TotalCapacity: calculateTotalCapacity(selected),
		UsedCapacity:  payloadSize,
		Utilization:   0, // Will be calculated below
		Warnings:      make([]string, 0),
	}

	plan.Utilization = float64(plan.UsedCapacity) / float64(plan.TotalCapacity)
	plan.AvgDetectability = calculateAvgDetectability(selected)
	plan.DiversityScore = calculateDiversityScore(selected)

	// Generate warnings
	if plan.Utilization > 0.9 {
		plan.Warnings = append(plan.Warnings,
			fmt.Sprintf("High capacity utilization (%.1f%%) may increase detectability", plan.Utilization*100))
	}

	if plan.AvgDetectability > 0.6 {
		plan.Warnings = append(plan.Warnings,
			fmt.Sprintf("High average detectability score (%.2f)", plan.AvgDetectability))
	}

	if plan.DiversityScore < 0.3 && len(selected) > 1 {
		plan.Warnings = append(plan.Warnings,
			"Low media type diversity may create detectable patterns")
	}

	return plan, nil
}

func calculateTotalCapacity(allocations []*CoverAllocation) int64 {
	total := int64(0)
	for _, alloc := range allocations {
		total += alloc.File.SafeCapacity
	}
	return total
}

func calculateAvgDetectability(allocations []*CoverAllocation) float64 {
	if len(allocations) == 0 {
		return 0
	}

	sum := 0.0
	for _, alloc := range allocations {
		score := CalculateDetectabilityScore(
			alloc.File.SafeCapacity,
			alloc.AllocatedBytes,
			alloc.Technique,
		)
		sum += score
	}

	return sum / float64(len(allocations))
}

func calculateDiversityScore(allocations []*CoverAllocation) float64 {
	if len(allocations) <= 1 {
		return 0
	}

	typeCount := make(map[media.MediaType]int)
	for _, alloc := range allocations {
		typeCount[alloc.File.Type]++
	}

	// Diversity score is the ratio of unique types to total files
	return float64(len(typeCount)) / float64(len(allocations))
}
