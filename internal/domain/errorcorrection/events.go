package errorcorrection

import (
	"time"
)

// DomainEvent represents something that happened in the Error Correction domain.
type DomainEvent interface {
	OccurredAt() time.Time
	EventType() string
}

// DataEncoded is emitted when data is successfully encoded with Reed-Solomon.
type DataEncoded struct {
	MessageID    MessageID
	DataShards   int
	ParityShards int
	EncodedAt    time.Time
}

func (e DataEncoded) OccurredAt() time.Time {
	return e.EncodedAt
}

func (e DataEncoded) EventType() string {
	return "errorcorrection.data_encoded"
}

// DataSharded is emitted when data is split into shards.
type DataSharded struct {
	MessageID  MessageID
	ShardCount int
	ShardedAt  time.Time
}

func (e DataSharded) OccurredAt() time.Time {
	return e.ShardedAt
}

func (e DataSharded) EventType() string {
	return "errorcorrection.data_sharded"
}

// DataDecoded is emitted when shards are successfully reconstructed.
type DataDecoded struct {
	MessageID  MessageID
	ShardsUsed int
	DecodedAt  time.Time
}

func (e DataDecoded) OccurredAt() time.Time {
	return e.DecodedAt
}

func (e DataDecoded) EventType() string {
	return "errorcorrection.data_decoded"
}

// ShardCorrupted is emitted when a corrupted shard is detected.
type ShardCorrupted struct {
	ShardID    ShardID
	MessageID  MessageID
	DetectedAt time.Time
}

func (e ShardCorrupted) OccurredAt() time.Time {
	return e.DetectedAt
}

func (e ShardCorrupted) EventType() string {
	return "errorcorrection.shard_corrupted"
}

// RecoveryCompleted is emitted when data is successfully recovered from partial shards.
type RecoveryCompleted struct {
	MessageID       MessageID
	AvailableShards int
	RequiredShards  int
	RecoveredAt     time.Time
}

func (e RecoveryCompleted) OccurredAt() time.Time {
	return e.RecoveredAt
}

func (e RecoveryCompleted) EventType() string {
	return "errorcorrection.recovery_completed"
}
