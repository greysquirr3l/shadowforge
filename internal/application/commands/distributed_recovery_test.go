package commands

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/greysquirr3l/shadowforge/internal/domain/distribution"
	"github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
	errorcorrection_infra "github.com/greysquirr3l/shadowforge/internal/infrastructure/errorcorrection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create valid RS-encoded shards for testing
func createValidShards(t *testing.T, dataShards, parityShards int, dataSize int) []*errorcorrection.Shard {
	t.Helper()

	// Create logger and RSService
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Suppress logs in tests
	svc := errorcorrection_infra.NewRSService(logger)

	// Create test payload
	payload := make([]byte, dataSize)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	// Create configuration
	config, err := errorcorrection.NewShardConfiguration(dataShards, parityShards)
	require.NoError(t, err)

	// Use RSService.Encode to create valid RS-encoded shards
	protectedMsg, err := svc.Encode(context.Background(), payload, config)
	require.NoError(t, err)

	// Extract shards from ProtectedMessage
	return protectedMsg.Shards
}

func TestRecoverDistributedHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		cmd            *RecoverDistributedCommand
		setupShards    func() []*errorcorrection.Shard
		expectSuccess  bool
		expectMethod   string
		expectQuality  float64
		expectErrors   int
		expectWarnings int
		description    string
	}{
		{
			name: "full_recovery_all_shards",
			cmd: &RecoverDistributedCommand{
				Manifest: &distribution.ShardManifest{
					RequiredShards: 10,
					TotalShards:    15,
				},
				MinimumThreshold: 10,
				AllowPartial:     false,
			},
			setupShards: func() []*errorcorrection.Shard {
				return createValidShards(t, 10, 5, 1024)
			},
			expectSuccess: true,
			expectMethod:  "full",
			expectQuality: 1.0,
			expectErrors:  0,
			description:   "Should recover with all 15 shards available",
		},
		{
			name: "partial_recovery_minimum_shards",
			cmd: &RecoverDistributedCommand{
				Manifest: &distribution.ShardManifest{
					RequiredShards: 10,
					TotalShards:    15,
				},
				MinimumThreshold: 10,
				AllowPartial:     true,
			},
			setupShards: func() []*errorcorrection.Shard {
				allShards := createValidShards(t, 10, 5, 1024)
				return allShards[:10] // Only return minimum required
			},
			expectSuccess:  true,
			expectMethod:   "partial",
			expectQuality:  0.67, // 10/15 = 0.666...
			expectErrors:   0,
			expectWarnings: 1, // Warning about partial recovery
			description:    "Should recover with exactly minimum shards",
		},
		{
			name: "insufficient_shards",
			cmd: &RecoverDistributedCommand{
				Manifest: &distribution.ShardManifest{
					RequiredShards: 10,
					TotalShards:    15,
				},
				MinimumThreshold: 10,
			},
			setupShards: func() []*errorcorrection.Shard {
				shards := make([]*errorcorrection.Shard, 9)
				for i := 0; i < 9; i++ {
					shards[i] = &errorcorrection.Shard{
						Index: i,
						Data:  []byte{byte(i)},
					}
				}
				return shards
			},
			expectSuccess: false,
			expectMethod:  "failed",
			expectErrors:  1, // Insufficient shards error
			description:   "Should fail with insufficient shards",
		},
		{
			name: "no_shards_available",
			cmd: &RecoverDistributedCommand{
				Manifest: &distribution.ShardManifest{
					RequiredShards: 10,
					TotalShards:    15,
				},
				MinimumThreshold: 10,
			},
			setupShards: func() []*errorcorrection.Shard {
				return []*errorcorrection.Shard{}
			},
			expectSuccess: false,
			expectMethod:  "",
			expectErrors:  1, // No shards error
			description:   "Should fail with no shards",
		},
		{
			name: "good_availability_with_warnings",
			cmd: &RecoverDistributedCommand{
				Manifest: &distribution.ShardManifest{
					RequiredShards: 10,
					TotalShards:    15,
				},
				MinimumThreshold: 10,
			},
			setupShards: func() []*errorcorrection.Shard {
				allShards := createValidShards(t, 10, 5, 1024)
				return allShards[:13] // 13 of 15 available (good redundancy)
			},
			expectSuccess:  true,
			expectMethod:   "partial",
			expectQuality:  0.87, // 13/15 = 0.8666...
			expectWarnings: 1,
			description:    "Should recover with good availability and warnings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			handler := NewRecoverDistributedHandler(logger)
			ctx := context.Background()

			// Prepare shards
			tt.cmd.AvailableShards = tt.setupShards()

			// Execute
			result, err := handler.Handle(ctx, tt.cmd)

			// Assert based on test case
			if tt.expectSuccess {
				require.NoError(t, err)
				assert.Equal(t, tt.expectSuccess, result.Success, tt.description)
				assert.Equal(t, tt.expectMethod, result.RecoveryMethod)
				assert.InDelta(t, tt.expectQuality, result.QualityScore, 0.01)
			} else {
				// Failed recovery cases
				assert.Equal(t, tt.expectSuccess, result.Success, tt.description)
			}

			assert.Len(t, result.Errors, tt.expectErrors)
			if tt.expectWarnings > 0 {
				assert.GreaterOrEqual(t, len(result.Warnings), tt.expectWarnings)
			}

			// Common assertions
			if tt.expectSuccess || len(tt.setupShards()) > 0 {
				// Only assert non-zero time if we attempted recovery
				assert.NotZero(t, result.RecoveryTime)
			}
			t.Logf("%s: Success=%v, Method=%s, Quality=%.2f, Errors=%d, Warnings=%d, Time=%v",
				tt.name, result.Success, result.RecoveryMethod, result.QualityScore,
				len(result.Errors), len(result.Warnings), result.RecoveryTime)
		})
	}
}

func TestRecoverDistributedHandler_ValidateInputs(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	handler := NewRecoverDistributedHandler(logger)

	tests := []struct {
		name        string
		cmd         *RecoverDistributedCommand
		expectError bool
		description string
	}{
		{
			name: "valid_inputs",
			cmd: &RecoverDistributedCommand{
				Manifest: &distribution.ShardManifest{
					RequiredShards: 10,
					TotalShards:    15,
				},
				AvailableShards:  []*errorcorrection.Shard{{Index: 0, Data: []byte("test")}},
				MinimumThreshold: 10,
			},
			expectError: false,
			description: "Should accept valid inputs",
		},
		{
			name: "nil_manifest",
			cmd: &RecoverDistributedCommand{
				AvailableShards:  []*errorcorrection.Shard{{Index: 0, Data: []byte("test")}},
				MinimumThreshold: 10,
			},
			expectError: true,
			description: "Should reject nil manifest",
		},
		{
			name: "no_shards",
			cmd: &RecoverDistributedCommand{
				Manifest: &distribution.ShardManifest{
					RequiredShards: 10,
					TotalShards:    15,
				},
				AvailableShards:  []*errorcorrection.Shard{},
				MinimumThreshold: 10,
			},
			expectError: true,
			description: "Should reject empty shard list",
		},
		{
			name: "auto_threshold",
			cmd: &RecoverDistributedCommand{
				Manifest: &distribution.ShardManifest{
					RequiredShards: 10,
					TotalShards:    15,
				},
				AvailableShards:  []*errorcorrection.Shard{{Index: 0, Data: []byte("test")}},
				MinimumThreshold: 0, // Should auto-set to DataShards
			},
			expectError: false,
			description: "Should auto-set minimum threshold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &RecoverDistributedResult{}
			err := handler.validateInputs(tt.cmd, result)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				if tt.cmd.MinimumThreshold == 0 {
					// Should be auto-set to DataShards
					assert.Equal(t, tt.cmd.Manifest.RequiredShards, result.ShardsRequired)
				}
			}
		})
	}
}

func TestRecoverDistributedHandler_AnalyzeShardQuality(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	handler := NewRecoverDistributedHandler(logger)

	tests := []struct {
		name           string
		shards         []*errorcorrection.Shard
		totalShards    int
		expectMissing  int
		expectQuality  float64
		expectWarnings int
		description    string
	}{
		{
			name: "all_shards_present",
			shards: func() []*errorcorrection.Shard {
				s := make([]*errorcorrection.Shard, 15)
				for i := 0; i < 15; i++ {
					s[i] = &errorcorrection.Shard{Index: i, Data: []byte{byte(i)}}
				}
				return s
			}(),
			totalShards:    15,
			expectMissing:  0,
			expectQuality:  1.0,
			expectWarnings: 0,
			description:    "Should have perfect quality with all shards",
		},
		{
			name: "some_shards_missing",
			shards: func() []*errorcorrection.Shard {
				s := make([]*errorcorrection.Shard, 10)
				for i := 0; i < 10; i++ {
					s[i] = &errorcorrection.Shard{Index: i, Data: []byte{byte(i)}}
				}
				return s
			}(),
			totalShards:    15,
			expectMissing:  5,
			expectQuality:  0.67,
			expectWarnings: 1, // Low availability warning
			description:    "Should detect missing shards",
		},
		{
			name: "very_low_availability",
			shards: func() []*errorcorrection.Shard {
				s := make([]*errorcorrection.Shard, 8)
				for i := 0; i < 8; i++ {
					s[i] = &errorcorrection.Shard{Index: i, Data: []byte{byte(i)}}
				}
				return s
			}(),
			totalShards:    15,
			expectMissing:  7,
			expectQuality:  0.53,
			expectWarnings: 2, // Low availability + too many missing
			description:    "Should warn on very low availability",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &RecoverDistributedCommand{
				Manifest: &distribution.ShardManifest{
					RequiredShards: 10,
					TotalShards:    tt.totalShards,
				},
				AvailableShards: tt.shards,
			}

			result := &RecoverDistributedResult{
				ShardsAvailable: len(tt.shards),
			}

			handler.analyzeShardQuality(cmd, result)

			assert.Len(t, result.MissingShards, tt.expectMissing, tt.description)
			assert.InDelta(t, tt.expectQuality, result.QualityScore, 0.01)
			assert.GreaterOrEqual(t, len(result.Warnings), tt.expectWarnings)

			t.Logf("%s: Missing=%d, Quality=%.2f, Warnings=%d",
				tt.name, len(result.MissingShards), result.QualityScore, len(result.Warnings))
		})
	}
}

func TestOptimizeRecovery(t *testing.T) {
	tests := []struct {
		name             string
		strategy         *OptimizeRecoveryStrategy
		expectStrategy   string
		expectConfidence float64
		expectOptimal    int
		description      string
	}{
		{
			name: "high_availability",
			strategy: &OptimizeRecoveryStrategy{
				AvailableShards: 14,
				TotalShards:     15,
				DataShards:      10,
				ParityShards:    5,
				NetworkLatency:  10 * time.Millisecond,
			},
			expectStrategy:   "reliable",
			expectConfidence: 0.99,
			expectOptimal:    14,
			description:      "Should choose reliable strategy with high availability",
		},
		{
			name: "good_availability",
			strategy: &OptimizeRecoveryStrategy{
				AvailableShards: 12,
				TotalShards:     15,
				DataShards:      10,
				ParityShards:    5,
				NetworkLatency:  10 * time.Millisecond,
			},
			expectStrategy:   "fast",
			expectConfidence: 0.95,
			expectOptimal:    12, // DataShards + 2
			description:      "Should choose fast strategy with good availability",
		},
		{
			name: "low_availability",
			strategy: &OptimizeRecoveryStrategy{
				AvailableShards: 10,
				TotalShards:     15,
				DataShards:      10,
				ParityShards:    5,
				NetworkLatency:  10 * time.Millisecond,
			},
			expectStrategy:   "minimal",
			expectConfidence: 0.85,
			expectOptimal:    10,
			description:      "Should choose minimal strategy with low availability",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := OptimizeRecovery(tt.strategy)

			assert.Equal(t, tt.expectStrategy, plan.Strategy, tt.description)
			assert.Equal(t, tt.expectConfidence, plan.Confidence)
			assert.Equal(t, tt.expectOptimal, plan.OptimalShards)
			assert.Equal(t, tt.strategy.DataShards, plan.MinimumShards)
			assert.NotZero(t, plan.ExpectedTime)

			t.Logf("%s: Strategy=%s, Confidence=%.2f, Optimal=%d, Time=%v",
				tt.name, plan.Strategy, plan.Confidence, plan.OptimalShards, plan.ExpectedTime)
		})
	}
}
