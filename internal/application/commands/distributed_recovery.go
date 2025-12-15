package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/greysquirr3l/shadowforge/internal/domain/distribution"
	"github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
	errorcorrection_infra "github.com/greysquirr3l/shadowforge/internal/infrastructure/errorcorrection"
)

// RecoverDistributedCommand attempts to recover encrypted payload from available shards
type RecoverDistributedCommand struct {
	AvailableShards  []*errorcorrection.Shard
	Manifest         *distribution.ShardManifest
	MinimumThreshold int // Minimum shards required for recovery
	AllowPartial     bool
}

// RecoverDistributedResult contains the result of recovery attempt
type RecoverDistributedResult struct {
	Success          bool
	RecoveredPayload []byte
	ShardsUsed       int
	ShardsAvailable  int
	ShardsRequired   int
	RecoveryMethod   string // "full", "partial", "failed"
	MissingShards    []int
	CorruptedShards  []int
	RecoveryTime     time.Duration
	QualityScore     float64 // 0.0 to 1.0 based on shard availability
	Warnings         []string
	Errors           []RecoveryError
}

// RecoveryError represents an error during recovery
type RecoveryError struct {
	Code     string
	Message  string
	Severity string
	ShardID  int
}

// RecoverDistributedHandler handles distributed payload recovery
type RecoverDistributedHandler struct {
	ecService errorcorrection.ErrorCorrectionService
}

// NewRecoverDistributedHandler creates a new recovery handler
func NewRecoverDistributedHandler(logger *logrus.Logger) *RecoverDistributedHandler {
	return &RecoverDistributedHandler{
		ecService: errorcorrection_infra.NewRSService(logger),
	}
}

// Handle executes the recovery command
func (h *RecoverDistributedHandler) Handle(ctx context.Context, cmd *RecoverDistributedCommand) (*RecoverDistributedResult, error) {
	startTime := time.Now()

	result := &RecoverDistributedResult{
		ShardsAvailable: len(cmd.AvailableShards),
		ShardsRequired:  cmd.MinimumThreshold,
		Errors:          []RecoveryError{},
		Warnings:        []string{},
		MissingShards:   []int{},
		CorruptedShards: []int{},
	}

	// Validate inputs
	if err := h.validateInputs(cmd, result); err != nil {
		return result, err
	}

	// Analyze shard quality
	h.analyzeShardQuality(cmd, result)

	// Check if recovery is possible
	if result.ShardsAvailable < result.ShardsRequired {
		result.Success = false
		result.RecoveryMethod = "failed"
		result.Errors = append(result.Errors, RecoveryError{
			Code:     "INSUFFICIENT_SHARDS",
			Message:  fmt.Sprintf("have %d shards, need %d for recovery", result.ShardsAvailable, result.ShardsRequired),
			Severity: "error",
		})
		result.RecoveryTime = time.Since(startTime)
		return result, nil
	}

	// Attempt recovery
	recovered, err := h.performRecovery(ctx, cmd, result)
	if err != nil {
		result.Success = false
		result.RecoveryMethod = "failed"
		result.Errors = append(result.Errors, RecoveryError{
			Code:     "RECOVERY_FAILED",
			Message:  err.Error(),
			Severity: "error",
		})
		result.RecoveryTime = time.Since(startTime)
		return result, nil
	}

	result.Success = true
	result.RecoveredPayload = recovered
	result.ShardsUsed = result.ShardsAvailable
	result.RecoveryTime = time.Since(startTime)

	// Determine recovery method
	if result.ShardsAvailable == cmd.Manifest.TotalShards {
		result.RecoveryMethod = "full"
		result.QualityScore = 1.0
	} else if result.ShardsAvailable >= result.ShardsRequired {
		result.RecoveryMethod = "partial"
		result.QualityScore = float64(result.ShardsAvailable) / float64(cmd.Manifest.TotalShards)
		result.Warnings = append(result.Warnings, fmt.Sprintf("recovered with %d/%d shards (%.1f%% redundancy)",
			result.ShardsAvailable, cmd.Manifest.TotalShards, result.QualityScore*100))
	}

	return result, nil
}

// validateInputs validates recovery command inputs
func (h *RecoverDistributedHandler) validateInputs(cmd *RecoverDistributedCommand, result *RecoverDistributedResult) error {
	if cmd.Manifest == nil {
		return fmt.Errorf("manifest is required")
	}

	if len(cmd.AvailableShards) == 0 {
		result.Errors = append(result.Errors, RecoveryError{
			Code:     "NO_SHARDS",
			Message:  "no shards available for recovery",
			Severity: "error",
		})
		return fmt.Errorf("no shards provided")
	}

	if cmd.MinimumThreshold <= 0 {
		cmd.MinimumThreshold = cmd.Manifest.RequiredShards
	}

	result.ShardsRequired = cmd.MinimumThreshold

	return nil
}

// analyzeShardQuality analyzes available shards for quality and integrity
func (h *RecoverDistributedHandler) analyzeShardQuality(cmd *RecoverDistributedCommand, result *RecoverDistributedResult) {
	// Track which shards we have
	shardPresence := make(map[int]bool)
	for _, shard := range cmd.AvailableShards {
		if shard == nil {
			result.Warnings = append(result.Warnings, "skipping nil shard")
			continue
		}
		shardPresence[shard.Index] = true
	}

	// Identify missing shards
	totalExpected := cmd.Manifest.TotalShards
	for i := 0; i < totalExpected; i++ {
		if !shardPresence[i] {
			result.MissingShards = append(result.MissingShards, i)
		}
	}

	// Warn if too many shards missing
	missingCount := len(result.MissingShards)
	parityShards := cmd.Manifest.TotalShards - cmd.Manifest.RequiredShards
	if missingCount > parityShards {
		result.Warnings = append(result.Warnings, fmt.Sprintf("missing %d shards (max recoverable loss: %d)", missingCount, parityShards))
	}

	// Calculate quality score
	if totalExpected > 0 {
		availableRatio := float64(result.ShardsAvailable) / float64(totalExpected)
		result.QualityScore = availableRatio

		if availableRatio < 0.7 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("low shard availability: %.1f%%", availableRatio*100))
		}
	}
}

// performRecovery performs the actual Reed-Solomon recovery
func (h *RecoverDistributedHandler) performRecovery(ctx context.Context, cmd *RecoverDistributedCommand, result *RecoverDistributedResult) ([]byte, error) {
	// Use error correction service to decode
	config, err := errorcorrection.NewShardConfiguration(
		cmd.Manifest.RequiredShards,
		cmd.Manifest.TotalShards-cmd.Manifest.RequiredShards,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid shard configuration: %w", err)
	}

	// Decode with error correction (Decode expects []*Shard, returns []byte)
	recovered, err := h.ecService.Decode(ctx, cmd.AvailableShards, config)
	if err != nil {
		return nil, fmt.Errorf("Reed-Solomon decode failed: %w", err)
	}

	return recovered, nil
}

// OptimizeRecoveryStrategy determines the optimal recovery strategy
type OptimizeRecoveryStrategy struct {
	AvailableShards int
	TotalShards     int
	DataShards      int
	ParityShards    int
	NetworkLatency  time.Duration
}

// OptimizedRecoveryPlan contains the optimized recovery plan
type OptimizedRecoveryPlan struct {
	RecommendedShards []int
	MinimumShards     int
	OptimalShards     int
	ExpectedTime      time.Duration
	Strategy          string // "fast", "reliable", "minimal"
	Confidence        float64
}

// OptimizeRecovery determines optimal recovery strategy
func OptimizeRecovery(strategy *OptimizeRecoveryStrategy) *OptimizedRecoveryPlan {
	plan := &OptimizedRecoveryPlan{
		MinimumShards: strategy.DataShards,
		OptimalShards: strategy.DataShards + (strategy.ParityShards / 2), // Use half parity for reliability
	}

	// Choose strategy based on shard availability
	availabilityRatio := float64(strategy.AvailableShards) / float64(strategy.TotalShards)

	if availabilityRatio >= 0.9 {
		// High availability - use all shards for maximum reliability
		plan.Strategy = "reliable"
		plan.OptimalShards = strategy.AvailableShards
		plan.Confidence = 0.99
	} else if availabilityRatio >= 0.7 {
		// Good availability - balance speed and reliability
		plan.Strategy = "fast"
		plan.OptimalShards = strategy.DataShards + 2 // Use 2 parity shards
		plan.Confidence = 0.95
	} else {
		// Low availability - use minimum shards
		plan.Strategy = "minimal"
		plan.OptimalShards = strategy.DataShards
		plan.Confidence = 0.85
	}

	// Estimate recovery time based on network latency and shard count
	plan.ExpectedTime = strategy.NetworkLatency * time.Duration(plan.OptimalShards)

	return plan
}
