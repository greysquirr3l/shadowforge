package errorcorrection

import (
	"context"
)

// ErrorCorrectionService defines the domain service interface for Reed-Solomon operations.
type ErrorCorrectionService interface {
	// Encode encodes data into shards using Reed-Solomon encoding.
	Encode(ctx context.Context, data []byte, config *ShardConfiguration) (*ProtectedMessage, error)

	// Decode reconstructs the original data from available shards.
	Decode(ctx context.Context, shards []*Shard, config *ShardConfiguration) ([]byte, error)

	// VerifyShards checks the integrity of all shards.
	VerifyShards(ctx context.Context, shards []*Shard) error

	// CalculateCapacity determines the maximum data size for a given shard configuration.
	CalculateCapacity(config *ShardConfiguration, shardSize int) int

	// OptimizeConfiguration recommends optimal shard configuration for given data size.
	OptimizeConfiguration(ctx context.Context, dataSize int, redundancyLevel RedundancyLevel) (*ShardConfiguration, error)
}
