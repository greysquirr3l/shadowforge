// Package distribution provides distribution domain service interfaces.
package distribution

import "context"

// Service defines domain operations for distribution.
type Service interface {
// CreateStrategy creates a new distribution strategy.
CreateStrategy(
ctx context.Context,
pattern PatternType,
dataShards, parityShards int,
targetMediaIDs []string,
) (*DistributionStrategy, error)

// DistributeShards distributes shards according to strategy.
DistributeShards(
ctx context.Context,
strategyID StrategyID,
shardData [][]byte,
) (*ShardManifest, error)

// ValidateStrategy validates a distribution strategy configuration.
ValidateStrategy(ctx context.Context, strategy *DistributionStrategy) error

// CalculateOptimalSharding calculates optimal data/parity shard configuration.
CalculateOptimalSharding(
ctx context.Context,
dataSize int64,
targetCount int,
) (dataShards, parityShards int, err error)

// CheckRecoverability checks if data can be recovered with available shards.
CheckRecoverability(
ctx context.Context,
manifestID ManifestID,
availableIndices []int,
) (bool, error)

// RevokeDistribution revokes a distribution strategy.
RevokeDistribution(ctx context.Context, strategyID StrategyID) error

// GetDistributionStatus retrieves current distribution status.
GetDistributionStatus(ctx context.Context, strategyID StrategyID) (DistributionStatus, error)
}
