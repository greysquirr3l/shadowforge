package errorcorrection

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
)

var (
	// ErrMessageNotFound is returned when a message is not found in the repository.
	ErrMessageNotFound = errors.New("message not found")

	// ErrShardNotFound is returned when a shard is not found in the repository.
	ErrShardNotFound = errors.New("shard not found")
)

// MemoryRepository is an in-memory implementation of ErrorCorrectionRepository.
// Thread-safe using sync.RWMutex for concurrent access.
type MemoryRepository struct {
	messages map[string]*errorcorrection.ProtectedMessage
	shards   map[string]*errorcorrection.Shard
	mu       sync.RWMutex
	logger   *slog.Logger
}

// NewMemoryRepository creates a new in-memory repository.
func NewMemoryRepository(logger *slog.Logger) *MemoryRepository {
	if logger == nil {
		logger = slog.Default()
	}

	return &MemoryRepository{
		messages: make(map[string]*errorcorrection.ProtectedMessage),
		shards:   make(map[string]*errorcorrection.Shard),
		logger:   logger,
	}
}

// SaveMessage persists a protected message.
func (r *MemoryRepository) SaveMessage(ctx context.Context, message *errorcorrection.ProtectedMessage) error {
	if message == nil {
		return errors.New("message cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.messages[message.ID.String()] = message

	r.logger.InfoContext(ctx, "Protected message saved to memory repository",
		slog.String("message_id", message.ID.String()),
		slog.Int("data_shards", message.DataShards),
		slog.Int("parity_shards", message.ParityShards),
		slog.Int("total_shards", len(message.Shards)))

	return nil
}

// GetMessage retrieves a protected message by ID.
func (r *MemoryRepository) GetMessage(ctx context.Context, id errorcorrection.MessageID) (*errorcorrection.ProtectedMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	message, exists := r.messages[id.String()]
	if !exists {
		r.logger.WarnContext(ctx, "Message not found in repository",
			slog.String("message_id", id.String()))
		return nil, fmt.Errorf("%w: %s", ErrMessageNotFound, id.String())
	}

	r.logger.InfoContext(ctx, "Protected message retrieved from memory repository",
		slog.String("message_id", id.String()))

	return message, nil
}

// DeleteMessage removes a protected message.
func (r *MemoryRepository) DeleteMessage(ctx context.Context, id errorcorrection.MessageID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.messages[id.String()]; !exists {
		r.logger.WarnContext(ctx, "Attempted to delete non-existent message",
			slog.String("message_id", id.String()))
		return fmt.Errorf("%w: %s", ErrMessageNotFound, id.String())
	}

	delete(r.messages, id.String())

	r.logger.InfoContext(ctx, "Protected message deleted from memory repository",
		slog.String("message_id", id.String()))

	return nil
}

// SaveShard persists a single shard.
func (r *MemoryRepository) SaveShard(ctx context.Context, shard *errorcorrection.Shard) error {
	if shard == nil {
		return errors.New("shard cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.shards[shard.ID.String()] = shard

	r.logger.InfoContext(ctx, "Shard saved to memory repository",
		slog.String("shard_id", shard.ID.String()),
		slog.String("message_id", shard.MessageID.String()),
		slog.Int("index", shard.Index),
		slog.Bool("is_parity", shard.IsParity))

	return nil
}

// GetShard retrieves a shard by ID.
func (r *MemoryRepository) GetShard(ctx context.Context, id errorcorrection.ShardID) (*errorcorrection.Shard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shard, exists := r.shards[id.String()]
	if !exists {
		r.logger.WarnContext(ctx, "Shard not found in repository",
			slog.String("shard_id", id.String()))
		return nil, fmt.Errorf("%w: %s", ErrShardNotFound, id.String())
	}

	r.logger.InfoContext(ctx, "Shard retrieved from memory repository",
		slog.String("shard_id", id.String()))

	return shard, nil
}

// GetShardsByMessage retrieves all shards belonging to a message.
func (r *MemoryRepository) GetShardsByMessage(ctx context.Context, messageID errorcorrection.MessageID) ([]*errorcorrection.Shard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var shards []*errorcorrection.Shard
	for _, shard := range r.shards {
		if shard.MessageID.String() == messageID.String() {
			shards = append(shards, shard)
		}
	}

	r.logger.InfoContext(ctx, "Shards retrieved by message ID",
		slog.String("message_id", messageID.String()),
		slog.Int("shard_count", len(shards)))

	return shards, nil
}

// DeleteShard removes a shard.
func (r *MemoryRepository) DeleteShard(ctx context.Context, id errorcorrection.ShardID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.shards[id.String()]; !exists {
		r.logger.WarnContext(ctx, "Attempted to delete non-existent shard",
			slog.String("shard_id", id.String()))
		return fmt.Errorf("%w: %s", ErrShardNotFound, id.String())
	}

	delete(r.shards, id.String())

	r.logger.InfoContext(ctx, "Shard deleted from memory repository",
		slog.String("shard_id", id.String()))

	return nil
}

// Clear removes all data from the repository (useful for testing).
func (r *MemoryRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.messages = make(map[string]*errorcorrection.ProtectedMessage)
	r.shards = make(map[string]*errorcorrection.Shard)

	r.logger.Info("Error correction memory repository cleared")
}

// MessageCount returns the number of stored messages (useful for testing).
func (r *MemoryRepository) MessageCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.messages)
}

// ShardCount returns the number of stored shards (useful for testing).
func (r *MemoryRepository) ShardCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.shards)
}
