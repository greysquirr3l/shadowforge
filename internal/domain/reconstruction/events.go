// Package reconstruction provides reconstruction domain events.
package reconstruction

import "time"

// ReconstructionStarted is emitted when a reconstruction session begins.
type ReconstructionStarted struct {
SessionID           SessionID
StrategyID          string
ManifestID          string
AvailableShardCount int
Algorithm           RecoveryAlgorithm
StartedAt           time.Time
}

// ShardVerified is emitted when a shard passes verification.
type ShardVerified struct {
SessionID       SessionID
ShardIndex      int
ChecksumValid   bool
IntegrityScore  float64
VerifiedAt      time.Time
}

// ShardVerificationFailed is emitted when shard verification fails.
type ShardVerificationFailed struct {
SessionID     SessionID
ShardIndex    int
FailureReason string
FailedAt      time.Time
}

// ReconstructionCompleted is emitted when reconstruction succeeds.
type ReconstructionCompleted struct {
SessionID     SessionID
Duration      time.Duration
TotalAttempts int
DataSize      int64
CompletedAt   time.Time
}

// ReconstructionFailed is emitted when reconstruction fails.
type ReconstructionFailed struct {
SessionID     SessionID
FailureReason string
AttemptCount  int
FailedAt      time.Time
}

// AttemptRetried is emitted when reconstruction is retried with a different algorithm.
type AttemptRetried struct {
SessionID         SessionID
AttemptID         AttemptID
PreviousAlgorithm RecoveryAlgorithm
NewAlgorithm      RecoveryAlgorithm
RetriedAt         time.Time
}

// SessionAbandoned is emitted when a reconstruction session is abandoned.
type SessionAbandoned struct {
SessionID   SessionID
Reason      string
AbandonedAt time.Time
}

// ProgressUpdated is emitted when reconstruction progress changes significantly.
type ProgressUpdated struct {
SessionID       SessionID
Progress        float64
VerifiedShards  int
RemainingShards int
UpdatedAt       time.Time
}
