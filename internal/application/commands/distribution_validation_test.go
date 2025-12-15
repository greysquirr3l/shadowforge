package commands

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDistributionValidator_ValidateManifest(t *testing.T) {
	validator := NewDistributionValidator()
	ctx := context.Background()

	tests := []struct {
		name           string
		manifest       *MatrixManifest
		expectValid    bool
		expectErrors   int
		expectWarnings int
		checksPassed   int
		description    string
	}{
		{
			name: "valid_manifest",
			manifest: &MatrixManifest{
				Version:      "1.0",
				Mode:         "round_robin",
				PayloadCount: 1,
				CoverCount:   2,
				TotalShards:  2,
				RSRedundancy: 0.5,
				CreatedAt:    time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
				AllocationMap: MatrixAllocationMap{
					Payloads: map[string]PayloadAllocation{
						"payload1": {
							Name:         "payload1",
							Size:         1024,
							DataShards:   1,
							ParityShards: 1,
							ShardPlacements: []ShardPlacement{
								{ShardIndex: 0, CoverPath: "/cover1.png", ShardHash: "hash1"},
								{ShardIndex: 1, CoverPath: "/cover2.png", ShardHash: "hash2"},
							},
						},
					},
					Covers: map[string]CoverAllocation{
						"/cover1.png": {Path: "/cover1.png", Capacity: 10000, UsedCapacity: 512, Utilization: 5.12, ShardCount: 1},
						"/cover2.png": {Path: "/cover2.png", Capacity: 10000, UsedCapacity: 512, Utilization: 5.12, ShardCount: 1},
					},
				},
				IntegrityHash: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			},
			expectValid:    true,
			expectErrors:   0,
			expectWarnings: 0,
			checksPassed:   6,
			description:    "Should pass all checks for valid manifest",
		},
		{
			name: "invalid_version",
			manifest: &MatrixManifest{
				Version:      "2.0",
				Mode:         "round_robin",
				PayloadCount: 1,
				CoverCount:   1,
				CreatedAt:    time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				AllocationMap: MatrixAllocationMap{
					Payloads: map[string]PayloadAllocation{
						"payload1": {Name: "payload1", Size: 1024, DataShards: 1, ParityShards: 0},
					},
					Covers: map[string]CoverAllocation{
						"/cover1.png": {Path: "/cover1.png", Capacity: 10000, ShardCount: 1},
					},
				},
			},
			expectValid:  false,
			expectErrors: 1,
			description:  "Should reject unsupported version",
		},
		{
			name: "future_timestamp",
			manifest: &MatrixManifest{
				Version:      "1.0",
				Mode:         "round_robin",
				PayloadCount: 1,
				CoverCount:   1,
				CreatedAt:    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				AllocationMap: MatrixAllocationMap{
					Payloads: map[string]PayloadAllocation{
						"payload1": {Name: "payload1", Size: 1024, DataShards: 1, ParityShards: 0},
					},
					Covers: map[string]CoverAllocation{
						"/cover1.png": {Path: "/cover1.png", Capacity: 10000, ShardCount: 1},
					},
				},
			},
			expectValid:  false,
			expectErrors: 1,
			description:  "Should reject future timestamps",
		},
		{
			name: "empty_allocation",
			manifest: &MatrixManifest{
				Version:      "1.0",
				Mode:         "round_robin",
				PayloadCount: 0,
				CoverCount:   0,
				CreatedAt:    time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				AllocationMap: MatrixAllocationMap{
					Payloads: map[string]PayloadAllocation{},
					Covers:   map[string]CoverAllocation{},
				},
			},
			expectValid:  false,
			expectErrors: 2, // Empty payload and cover maps
			description:  "Should reject empty allocations",
		},
		{
			name: "inconsistent_mapping",
			manifest: &MatrixManifest{
				Version:      "1.0",
				Mode:         "round_robin",
				PayloadCount: 1,
				CoverCount:   1,
				TotalShards:  2,
				CreatedAt:    time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				AllocationMap: MatrixAllocationMap{
					Payloads: map[string]PayloadAllocation{
						"payload1": {
							Name:         "payload1",
							Size:         1024,
							DataShards:   1,
							ParityShards: 1,
							ShardPlacements: []ShardPlacement{
								{ShardIndex: 0, CoverPath: "/cover1.png", ShardHash: "hash1"},
								{ShardIndex: 1, CoverPath: "/cover2.png", ShardHash: "hash2"},
							},
						},
					},
					Covers: map[string]CoverAllocation{
						"/cover1.png": {Path: "/cover1.png", Capacity: 10000, UsedCapacity: 512, ShardCount: 1},
						// Missing /cover2.png - creates inconsistency
					},
				},
			},
			expectValid:  false,
			expectErrors: 1,
			description:  "Should detect inconsistent shard mappings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateManifest(ctx, tt.manifest)

			assert.Equal(t, tt.expectValid, result.Valid,
				"%s: validity mismatch", tt.description)

			if tt.expectErrors > 0 {
				assert.GreaterOrEqual(t, len(result.Errors), tt.expectErrors,
					"%s: expected at least %d errors", tt.description, tt.expectErrors)
			}

			if tt.expectWarnings > 0 {
				assert.GreaterOrEqual(t, len(result.Warnings), tt.expectWarnings,
					"%s: expected at least %d warnings", tt.description, tt.expectWarnings)
			}

			if tt.checksPassed > 0 {
				assert.Equal(t, tt.checksPassed, result.ChecksPassed,
					"%s: checks passed mismatch", tt.description)
			}

			t.Logf("%s: Valid=%v, Errors=%d, Warnings=%d, Checks=%d/%d",
				tt.name, result.Valid, len(result.Errors), len(result.Warnings),
				result.ChecksPassed, result.ChecksTotal)
		})
	}
}

func TestDistributionValidator_ValidateAllocation(t *testing.T) {
	validator := NewDistributionValidator()
	ctx := context.Background()

	tests := []struct {
		name         string
		allocation   *MatrixAllocationMap
		pattern      string
		expectValid  bool
		expectErrors int
		description  string
	}{
		{
			name: "valid_one_to_many",
			allocation: &MatrixAllocationMap{
				Payloads: map[string]PayloadAllocation{
					"payload1": {
						Name:         "payload1",
						Size:         1024,
						DataShards:   2,
						ParityShards: 1,
						ShardPlacements: []ShardPlacement{
							{ShardIndex: 0, CoverPath: "/cover1.png", ShardHash: "hash1"},
							{ShardIndex: 1, CoverPath: "/cover2.png", ShardHash: "hash2"},
							{ShardIndex: 2, CoverPath: "/cover3.png", ShardHash: "hash3"},
						},
					},
				},
				Covers: map[string]CoverAllocation{
					"/cover1.png": {Path: "/cover1.png", Capacity: 10000, UsedCapacity: 512, ShardCount: 1},
					"/cover2.png": {Path: "/cover2.png", Capacity: 10000, UsedCapacity: 512, ShardCount: 1},
					"/cover3.png": {Path: "/cover3.png", Capacity: 10000, UsedCapacity: 512, ShardCount: 1},
				},
			},
			pattern:     "one-to-many",
			expectValid: true,
			description: "Should validate correct one-to-many allocation",
		},
		{
			name: "invalid_one_to_many_too_many_payloads",
			allocation: &MatrixAllocationMap{
				Payloads: map[string]PayloadAllocation{
					"payload1": {
						Name:         "payload1",
						Size:         1024,
						DataShards:   1,
						ParityShards: 0,
						ShardPlacements: []ShardPlacement{
							{ShardIndex: 0, CoverPath: "/cover1.png", ShardHash: "hash1"},
						},
					},
					"payload2": {
						Name:         "payload2",
						Size:         2048,
						DataShards:   1,
						ParityShards: 0,
						ShardPlacements: []ShardPlacement{
							{ShardIndex: 0, CoverPath: "/cover2.png", ShardHash: "hash2"},
						},
					},
				},
				Covers: map[string]CoverAllocation{
					"/cover1.png": {Path: "/cover1.png", Capacity: 10000, UsedCapacity: 512, ShardCount: 1},
					"/cover2.png": {Path: "/cover2.png", Capacity: 10000, UsedCapacity: 1024, ShardCount: 1},
				},
			},
			pattern:      "one-to-many",
			expectValid:  false,
			expectErrors: 1,
			description:  "Should reject one-to-many with multiple payloads",
		},
		{
			name: "invalid_one_to_many_too_few_covers",
			allocation: &MatrixAllocationMap{
				Payloads: map[string]PayloadAllocation{
					"payload1": {
						Name:         "payload1",
						Size:         1024,
						DataShards:   1,
						ParityShards: 0,
						ShardPlacements: []ShardPlacement{
							{ShardIndex: 0, CoverPath: "/cover1.png", ShardHash: "hash1"},
						},
					},
				},
				Covers: map[string]CoverAllocation{
					"/cover1.png": {Path: "/cover1.png", Capacity: 10000, UsedCapacity: 512, ShardCount: 1},
				},
			},
			pattern:      "one-to-many",
			expectValid:  false,
			expectErrors: 1,
			description:  "Should reject one-to-many with too few covers",
		},
		{
			name: "valid_many_to_one",
			allocation: &MatrixAllocationMap{
				Payloads: map[string]PayloadAllocation{
					"payload1": {
						Name:         "payload1",
						Size:         1024,
						DataShards:   1,
						ParityShards: 0,
						ShardPlacements: []ShardPlacement{
							{ShardIndex: 0, CoverPath: "/cover1.png", ShardHash: "hash1"},
						},
					},
					"payload2": {
						Name:         "payload2",
						Size:         2048,
						DataShards:   1,
						ParityShards: 0,
						ShardPlacements: []ShardPlacement{
							{ShardIndex: 0, CoverPath: "/cover1.png", ShardHash: "hash2"},
						},
					},
				},
				Covers: map[string]CoverAllocation{
					"/cover1.png": {Path: "/cover1.png", Capacity: 20000, UsedCapacity: 1536, ShardCount: 2},
				},
			},
			pattern:     "many-to-one",
			expectValid: true,
			description: "Should validate correct many-to-one allocation",
		},
		{
			name: "invalid_many_to_one_too_many_covers",
			allocation: &MatrixAllocationMap{
				Payloads: map[string]PayloadAllocation{
					"payload1": {
						Name:         "payload1",
						Size:         1024,
						DataShards:   1,
						ParityShards: 0,
						ShardPlacements: []ShardPlacement{
							{ShardIndex: 0, CoverPath: "/cover1.png", ShardHash: "hash1"},
						},
					},
					"payload2": {
						Name:         "payload2",
						Size:         2048,
						DataShards:   1,
						ParityShards: 0,
						ShardPlacements: []ShardPlacement{
							{ShardIndex: 0, CoverPath: "/cover2.png", ShardHash: "hash2"},
						},
					},
				},
				Covers: map[string]CoverAllocation{
					"/cover1.png": {Path: "/cover1.png", Capacity: 10000, UsedCapacity: 512, ShardCount: 1},
					"/cover2.png": {Path: "/cover2.png", Capacity: 10000, UsedCapacity: 1024, ShardCount: 1},
				},
			},
			pattern:      "many-to-one",
			expectValid:  false,
			expectErrors: 1,
			description:  "Should reject many-to-one with multiple covers",
		},
		{
			name: "valid_many_to_many",
			allocation: &MatrixAllocationMap{
				Payloads: map[string]PayloadAllocation{
					"payload1": {
						Name:         "payload1",
						Size:         1024,
						DataShards:   1,
						ParityShards: 1,
						ShardPlacements: []ShardPlacement{
							{ShardIndex: 0, CoverPath: "/cover1.png", ShardHash: "hash1"},
							{ShardIndex: 1, CoverPath: "/cover2.png", ShardHash: "hash2"},
						},
					},
					"payload2": {
						Name:         "payload2",
						Size:         2048,
						DataShards:   1,
						ParityShards: 1,
						ShardPlacements: []ShardPlacement{
							{ShardIndex: 0, CoverPath: "/cover1.png", ShardHash: "hash3"},
							{ShardIndex: 1, CoverPath: "/cover2.png", ShardHash: "hash4"},
						},
					},
				},
				Covers: map[string]CoverAllocation{
					"/cover1.png": {Path: "/cover1.png", Capacity: 10000, UsedCapacity: 1536, ShardCount: 2},
					"/cover2.png": {Path: "/cover2.png", Capacity: 10000, UsedCapacity: 1536, ShardCount: 2},
				},
			},
			pattern:     "many-to-many",
			expectValid: true,
			description: "Should validate correct many-to-many allocation",
		},
		{
			name: "invalid_many_to_many_insufficient_payloads",
			allocation: &MatrixAllocationMap{
				Payloads: map[string]PayloadAllocation{
					"payload1": {
						Name:         "payload1",
						Size:         1024,
						DataShards:   1,
						ParityShards: 0,
						ShardPlacements: []ShardPlacement{
							{ShardIndex: 0, CoverPath: "/cover1.png", ShardHash: "hash1"},
						},
					},
				},
				Covers: map[string]CoverAllocation{
					"/cover1.png": {Path: "/cover1.png", Capacity: 10000, UsedCapacity: 512, ShardCount: 1},
					"/cover2.png": {Path: "/cover2.png", Capacity: 10000, UsedCapacity: 0, ShardCount: 0},
				},
			},
			pattern:      "many-to-many",
			expectValid:  false,
			expectErrors: 1,
			description:  "Should reject many-to-many with only 1 payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateAllocation(ctx, tt.allocation, tt.pattern)

			assert.Equal(t, tt.expectValid, result.Valid,
				"%s: validity mismatch", tt.description)

			if tt.expectErrors > 0 {
				assert.GreaterOrEqual(t, len(result.Errors), tt.expectErrors,
					"%s: expected at least %d errors", tt.description, tt.expectErrors)
			}

			t.Logf("%s: Valid=%v, Errors=%d, Checks=%d/%d",
				tt.name, result.Valid, len(result.Errors),
				result.ChecksPassed, result.ChecksTotal)
		})
	}
}

func TestDistributionValidator_ValidateConfiguration(t *testing.T) {
	validator := NewDistributionValidator()
	ctx := context.Background()

	tests := []struct {
		name         string
		config       *DistributionConfig
		expectValid  bool
		expectErrors int
		description  string
	}{
		{
			name: "valid_config",
			config: &DistributionConfig{
				DataShards:     10,
				ParityShards:   3,
				AllocationMode: MatrixModeBalanced,
				MaxUtilization: 0.8,
			},
			expectValid: true,
			description: "Should validate correct configuration",
		},
		{
			name: "invalid_zero_data_shards",
			config: &DistributionConfig{
				DataShards:     0,
				ParityShards:   3,
				AllocationMode: MatrixModeBalanced,
			},
			expectValid:  false,
			expectErrors: 1,
			description:  "Should reject zero data shards",
		},
		{
			name: "invalid_zero_parity_shards",
			config: &DistributionConfig{
				DataShards:     10,
				ParityShards:   0,
				AllocationMode: MatrixModeBalanced,
			},
			expectValid:  false,
			expectErrors: 1,
			description:  "Should reject zero parity shards",
		},
		{
			name: "invalid_mode",
			config: &DistributionConfig{
				DataShards:     10,
				ParityShards:   3,
				AllocationMode: "invalid_mode", // Invalid mode string
			},
			expectValid:  false,
			expectErrors: 1,
			description:  "Should reject invalid allocation mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateConfiguration(ctx, tt.config)

			assert.Equal(t, tt.expectValid, result.Valid,
				"%s: validity mismatch", tt.description)

			if tt.expectErrors > 0 {
				assert.GreaterOrEqual(t, len(result.Errors), tt.expectErrors,
					"%s: expected at least %d errors", tt.description, tt.expectErrors)
			}

			t.Logf("%s: Valid=%v, Errors=%d, Warnings=%d",
				tt.name, result.Valid, len(result.Errors), len(result.Warnings))
		})
	}
}

func TestDistributionValidator_ValidateCapacity(t *testing.T) {
	validator := NewDistributionValidator()
	ctx := context.Background()

	tests := []struct {
		name         string
		payloadSizes []int64
		covers       []CoverCapacityInfo
		rsRedundancy float64
		expectValid  bool
		description  string
	}{
		{
			name:         "sufficient_capacity",
			payloadSizes: []int64{10240}, // 10KB
			covers: []CoverCapacityInfo{
				{Path: "/cover1.png", Capacity: 50000},
				{Path: "/cover2.png", Capacity: 50000},
			},
			rsRedundancy: 0.3,
			expectValid:  true,
			description:  "Should validate sufficient capacity",
		},
		{
			name:         "insufficient_capacity",
			payloadSizes: []int64{100000}, // 100KB
			covers: []CoverCapacityInfo{
				{Path: "/cover1.png", Capacity: 10000},
			},
			rsRedundancy: 0.3,
			expectValid:  false,
			description:  "Should detect insufficient capacity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateCapacity(ctx, tt.payloadSizes, tt.covers, tt.rsRedundancy)

			assert.Equal(t, tt.expectValid, result.Valid,
				"%s: validity mismatch", tt.description)

			t.Logf("%s: Valid=%v, Errors=%d, Warnings=%d",
				tt.name, result.Valid, len(result.Errors), len(result.Warnings))
		})
	}
}

func TestValidationResult_Summary(t *testing.T) {
	result := ValidationResult{
		Valid:        false,
		ChecksPassed: 5,
		ChecksTotal:  7,
		Errors: []ValidationError{
			{Code: "ERR1", Message: "error 1", Severity: "error"},
			{Code: "ERR2", Message: "error 2", Severity: "error"},
		},
		Warnings:    []string{"warning 1", "warning 2"},
		ValidatedAt: time.Now(),
	}

	assert.False(t, result.Valid)
	assert.Equal(t, 2, len(result.Errors))
	assert.Equal(t, 2, len(result.Warnings))
	assert.Equal(t, 5, result.ChecksPassed)
	assert.Equal(t, 7, result.ChecksTotal)
	assert.False(t, result.ValidatedAt.IsZero())
}
