// Package distribution provides distribution domain events.
package distribution

import "time"

// StrategyCreated is emitted when a new distribution strategy is created.
type StrategyCreated struct {
	StrategyID       StrategyID
	Pattern          PatternType
	DataShards       int
	ParityShards     int
	TotalShards      int
	TargetMediaCount int
	CreatedAt        time.Time
}

// DistributionStarted is emitted when shard distribution begins.
type DistributionStarted struct {
	StrategyID  StrategyID
	ManifestID  ManifestID
	TotalShards int
	StartedAt   time.Time
}

// ShardDistributed is emitted when an individual shard is distributed.
type ShardDistributed struct {
	ManifestID    ManifestID
	ShardIndex    int
	ShardID       string
	MediaID       string
	Size          int64
	DistributedAt time.Time
}

// DistributionCompleted is emitted when all shards are successfully distributed.
type DistributionCompleted struct {
	StrategyID  StrategyID
	ManifestID  ManifestID
	TotalShards int
	Duration    time.Duration
	CompletedAt time.Time
}

// DistributionFailed is emitted when distribution fails.
type DistributionFailed struct {
	StrategyID         StrategyID
	FailureReason      string
	FailedShardIndices []int
	FailedAt           time.Time
}

// DistributionRevoked is emitted when a distribution is revoked.
type DistributionRevoked struct {
	StrategyID StrategyID
	RevokedBy  string
	Reason     string
	RevokedAt  time.Time
}
