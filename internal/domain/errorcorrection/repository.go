package errorcorrection

import (
	"context"
)

// ErrorCorrectionRepository defines the persistence interface for error correction domain objects.
type ErrorCorrectionRepository interface {
	// SaveMessage persists a protected message.
	SaveMessage(ctx context.Context, message *ProtectedMessage) error

	// GetMessage retrieves a protected message by ID.
	GetMessage(ctx context.Context, id MessageID) (*ProtectedMessage, error)

	// DeleteMessage removes a protected message.
	DeleteMessage(ctx context.Context, id MessageID) error

	// SaveShard persists a single shard.
	SaveShard(ctx context.Context, shard *Shard) error

	// GetShard retrieves a shard by ID.
	GetShard(ctx context.Context, id ShardID) (*Shard, error)

	// GetShardsByMessage retrieves all shards belonging to a message.
	GetShardsByMessage(ctx context.Context, messageID MessageID) ([]*Shard, error)

	// DeleteShard removes a shard.
	DeleteShard(ctx context.Context, id ShardID) error
}
