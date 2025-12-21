package commands

import (
	"context"
	"fmt"
	"time"
)

// DistributionValidator provides validation services for distribution operations
type DistributionValidator struct {
	capacityPlanner *CapacityPlanner
}

// NewDistributionValidator creates a new validator instance
func NewDistributionValidator() *DistributionValidator {
	return &DistributionValidator{
		capacityPlanner: NewCapacityPlanner(),
	}
}

// ValidationResult contains the outcome of validation checks
type ValidationResult struct {
	Valid        bool
	Errors       []ValidationError
	Warnings     []string
	ChecksPassed int
	ChecksTotal  int
	ValidatedAt  time.Time
}

// ValidationError describes a specific validation failure
type ValidationError struct {
	Code     string
	Message  string
	Field    string
	Severity string // "error", "warning", "info"
}

// ValidateManifest verifies manifest integrity and consistency
func (v *DistributionValidator) ValidateManifest(ctx context.Context, manifest *MatrixManifest) ValidationResult {
	result := ValidationResult{
		Valid:       true,
		ValidatedAt: time.Now(),
	}

	// Check 1: Version compatibility
	result.ChecksTotal++
	if manifest.Version != "1.0" {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "INVALID_VERSION",
			Message:  fmt.Sprintf("unsupported manifest version: %s", manifest.Version),
			Field:    "version",
			Severity: "error",
		})
		result.Valid = false
	} else {
		result.ChecksPassed++
	}

	// Check 2: Created timestamp
	result.ChecksTotal++
	if manifest.CreatedAt == "" {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "MISSING_TIMESTAMP",
			Message:  "manifest missing creation timestamp",
			Field:    "created_at",
			Severity: "warning",
		})
		result.Warnings = append(result.Warnings, "manifest timestamp is missing")
	} else {
		// Parse and validate timestamp
		createdTime, err := time.Parse(time.RFC3339, manifest.CreatedAt)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "INVALID_TIMESTAMP",
				Message:  fmt.Sprintf("invalid timestamp format: %v", err),
				Field:    "created_at",
				Severity: "error",
			})
			result.Valid = false
		} else if createdTime.After(time.Now()) {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "FUTURE_TIMESTAMP",
				Message:  "manifest timestamp is in the future",
				Field:    "created_at",
				Severity: "error",
			})
			result.Valid = false
		} else {
			result.ChecksPassed++
		}
	}

	// Check 3: Payload allocations
	result.ChecksTotal++
	if len(manifest.AllocationMap.Payloads) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "EMPTY_ALLOCATION",
			Message:  "manifest has no payload allocations",
			Field:    "allocation_map.payloads",
			Severity: "error",
		})
		result.Valid = false
	} else {
		result.ChecksPassed++
	}

	// Check 4: Cover allocations
	result.ChecksTotal++
	if len(manifest.AllocationMap.Covers) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "EMPTY_COVER_MAP",
			Message:  "manifest has no cover allocations",
			Field:    "allocation_map.covers",
			Severity: "error",
		})
		result.Valid = false
	} else {
		result.ChecksPassed++
	}

	// Check 5: Shard placement consistency
	result.ChecksTotal++
	allShardPlacements := make(map[string]int) // cover_path -> shard count
	totalShardsExpected := 0

	// Count shards from payload allocations
	for _, payloadAlloc := range manifest.AllocationMap.Payloads {
		totalShardsExpected += len(payloadAlloc.ShardPlacements)
		for _, placement := range payloadAlloc.ShardPlacements {
			allShardPlacements[placement.CoverPath]++
		}
	}

	// Verify all referenced covers exist and have matching shard counts
	coverMismatch := false

	// First, check if all shard placements reference existing covers
	for coverPath := range allShardPlacements {
		if _, exists := manifest.AllocationMap.Covers[coverPath]; !exists {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "MISSING_COVER",
				Message:  fmt.Sprintf("shard placement references non-existent cover: %s", coverPath),
				Field:    "allocation_map",
				Severity: "error",
			})
			result.Valid = false
			coverMismatch = true
		}
	}

	// Second, verify covers have matching shard counts
	for coverPath, coverAlloc := range manifest.AllocationMap.Covers {
		if allShardPlacements[coverPath] != coverAlloc.ShardCount {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "INCONSISTENT_MAPPING",
				Message:  fmt.Sprintf("cover %s: expected %d shards from placements, cover reports %d", coverPath, allShardPlacements[coverPath], coverAlloc.ShardCount),
				Field:    "allocation_map",
				Severity: "error",
			})
			result.Valid = false
			coverMismatch = true
		}
	}

	if !coverMismatch {
		result.ChecksPassed++
	}

	// Check 6: Integrity hash (if present)
	result.ChecksTotal++
	if manifest.IntegrityHash == "" {
		result.Warnings = append(result.Warnings, "manifest has no integrity hash")
		result.ChecksPassed++ // Not critical
	} else {
		// Note: Actual hash verification would require knowing the HMAC key
		// For now, just check that it's present and formatted correctly
		if len(manifest.IntegrityHash) < 32 { // Minimum for SHA256 hex
			result.Errors = append(result.Errors, ValidationError{
				Code:     "INVALID_HASH",
				Message:  "integrity hash appears too short",
				Field:    "integrity_hash",
				Severity: "warning",
			})
			result.Warnings = append(result.Warnings, "integrity hash may be invalid")
		} else {
			result.ChecksPassed++
		}
	}

	return result
}

// ValidateAllocation checks if shard allocation is valid for a given pattern
func (v *DistributionValidator) ValidateAllocation(ctx context.Context, allocation *MatrixAllocationMap, pattern string) ValidationResult {
	result := ValidationResult{
		Valid:       true,
		ValidatedAt: time.Now(),
	}

	// Check 1: Non-empty allocations
	result.ChecksTotal++
	if len(allocation.Payloads) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "NO_PAYLOADS",
			Message:  "allocation has no payloads",
			Severity: "error",
		})
		result.Valid = false
	} else {
		result.ChecksPassed++
	}

	result.ChecksTotal++
	if len(allocation.Covers) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "NO_COVERS",
			Message:  "allocation has no cover mappings",
			Severity: "error",
		})
		result.Valid = false
	} else {
		result.ChecksPassed++
	}

	// Check 2: Pattern-specific validation
	result.ChecksTotal++
	payloadCount := len(allocation.Payloads)
	coverCount := len(allocation.Covers)

	switch pattern {
	case "one-to-many":
		if payloadCount != 1 {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "INVALID_PATTERN",
				Message:  fmt.Sprintf("one-to-many requires exactly 1 payload, got %d", payloadCount),
				Severity: "error",
			})
			result.Valid = false
		} else if coverCount < 3 {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "INSUFFICIENT_COVERS",
				Message:  fmt.Sprintf("one-to-many requires at least 3 covers, got %d", coverCount),
				Severity: "error",
			})
			result.Valid = false
		} else {
			result.ChecksPassed++
		}

	case "many-to-one":
		if payloadCount < 2 {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "INSUFFICIENT_PAYLOADS",
				Message:  fmt.Sprintf("many-to-one requires at least 2 payloads, got %d", payloadCount),
				Severity: "error",
			})
			result.Valid = false
		} else if coverCount != 1 {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "INVALID_COVER_COUNT",
				Message:  fmt.Sprintf("many-to-one requires exactly 1 cover, got %d", coverCount),
				Severity: "error",
			})
			result.Valid = false
		} else {
			result.ChecksPassed++
		}

	case "many-to-many":
		if payloadCount < 2 {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "INSUFFICIENT_PAYLOADS",
				Message:  fmt.Sprintf("many-to-many requires at least 2 payloads, got %d", payloadCount),
				Severity: "error",
			})
			result.Valid = false
		} else if coverCount < 2 {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "INSUFFICIENT_COVERS",
				Message:  fmt.Sprintf("many-to-many requires at least 2 covers, got %d", coverCount),
				Severity: "error",
			})
			result.Valid = false
		} else {
			result.ChecksPassed++
		}

	default:
		result.ChecksPassed++
	}

	// Check 3: Shard distribution balance
	result.ChecksTotal++
	var maxShards, minShards int
	first := true
	for _, coverAlloc := range allocation.Covers {
		count := coverAlloc.ShardCount
		if first {
			maxShards, minShards = count, count
			first = false
		} else {
			if count > maxShards {
				maxShards = count
			}
			if count < minShards {
				minShards = count
			}
		}
	}

	if maxShards > 0 && minShards > 0 {
		imbalance := float64(maxShards-minShards) / float64(maxShards)
		if imbalance > 0.5 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("significant load imbalance: max=%d, min=%d shards per cover", maxShards, minShards))
		}
	}
	result.ChecksPassed++

	return result
}

// ValidateConfiguration checks if distribution configuration is valid
func (v *DistributionValidator) ValidateConfiguration(ctx context.Context, config *DistributionConfig) ValidationResult {
	result := ValidationResult{
		Valid:       true,
		ValidatedAt: time.Now(),
	}

	// Check 1: Reed-Solomon parameters
	result.ChecksTotal++
	if config.DataShards < 1 {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "INVALID_DATA_SHARDS",
			Message:  "data shards must be at least 1",
			Severity: "error",
		})
		result.Valid = false
	} else {
		result.ChecksPassed++
	}

	result.ChecksTotal++
	if config.ParityShards < 1 {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "INVALID_PARITY_SHARDS",
			Message:  "parity shards must be at least 1",
			Severity: "error",
		})
		result.Valid = false
	} else {
		result.ChecksPassed++
	}

	// Check 2: Redundancy ratio
	result.ChecksTotal++
	if config.DataShards > 0 {
		redundancy := float64(config.ParityShards) / float64(config.DataShards)
		if redundancy < 0.1 {
			result.Warnings = append(result.Warnings, "very low redundancy (<10%) - high data loss risk")
		} else if redundancy > 2.0 {
			result.Warnings = append(result.Warnings, "very high redundancy (>200%) - inefficient space usage")
		}
	}
	result.ChecksPassed++

	// Check 3: Mode validity
	result.ChecksTotal++
	validModes := map[MatrixMode]bool{
		MatrixModeRoundRobin: true,
		MatrixModeRandom:     true,
		MatrixModeOptimized:  true,
		MatrixModeBalanced:   true,
	}
	if !validModes[config.AllocationMode] {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "INVALID_MODE",
			Message:  fmt.Sprintf("invalid allocation mode: %s", config.AllocationMode),
			Severity: "error",
		})
		result.Valid = false
	} else {
		result.ChecksPassed++
	}

	return result
}

// DistributionConfig represents configuration for distribution
type DistributionConfig struct {
	DataShards     int
	ParityShards   int
	AllocationMode MatrixMode
	MaxUtilization float64
}

// ValidateCapacity checks if available capacity is sufficient
func (v *DistributionValidator) ValidateCapacity(ctx context.Context, payloadSizes []int64, covers []CoverCapacityInfo, rsRedundancy float64) ValidationResult {
	result := ValidationResult{
		Valid:       true,
		ValidatedAt: time.Now(),
	}

	// Use capacity planner for validation
	analysis, err := v.capacityPlanner.AnalyzeCapacity(ctx, payloadSizes, covers, rsRedundancy)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:     "CAPACITY_ANALYSIS_FAILED",
			Message:  err.Error(),
			Severity: "error",
		})
		return result
	}

	result.ChecksTotal = 3

	// Check 1: Sufficient capacity
	if !analysis.IsFeasible {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:     "INSUFFICIENT_CAPACITY",
			Message:  "total capacity insufficient for payload(s) with Reed-Solomon overhead",
			Severity: "error",
		})
	} else {
		result.ChecksPassed++
	}

	// Check 2: Utilization within limits
	if analysis.UtilizationPercent > 90.0 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("very high utilization: %.1f%%", analysis.UtilizationPercent))
	}
	result.ChecksPassed++

	// Check 3: Warnings from capacity analysis
	if len(analysis.Warnings) > 0 {
		result.Warnings = append(result.Warnings, analysis.Warnings...)
	}
	result.ChecksPassed++

	return result
}
