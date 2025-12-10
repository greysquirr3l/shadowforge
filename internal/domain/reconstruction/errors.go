// Package reconstruction provides domain logic for shard reconstruction and data recovery.
package reconstruction

import "errors"

// Validation errors
var (
// ErrInvalidSessionID indicates an invalid or empty session ID.
ErrInvalidSessionID = errors.New("invalid session ID")

// ErrInvalidAttemptID indicates an invalid or empty attempt ID.
ErrInvalidAttemptID = errors.New("invalid attempt ID")

// ErrInvalidStrategy indicates an invalid recovery strategy.
ErrInvalidStrategy = errors.New("invalid recovery strategy")

// ErrInvalidStrategyID indicates an invalid or empty strategy ID.
ErrInvalidStrategyID = errors.New("invalid strategy ID")

// ErrInvalidManifestID indicates an invalid or empty manifest ID.
ErrInvalidManifestID = errors.New("invalid manifest ID")

// ErrInvalidProgress indicates progress value outside valid range.
ErrInvalidProgress = errors.New("invalid progress value (must be 0.0-1.0)")
)

// Shard errors
var (
// ErrNoAvailableShards indicates no shards are available for reconstruction.
ErrNoAvailableShards = errors.New("no available shards for reconstruction")

// ErrNoShardsUsed indicates no shards were specified for recovery attempt.
ErrNoShardsUsed = errors.New("no shards used in recovery attempt")

// ErrInsufficientShards indicates not enough shards for successful recovery.
ErrInsufficientShards = errors.New("insufficient shards for recovery")

// ErrShardVerificationFailed indicates shard integrity verification failed.
ErrShardVerificationFailed = errors.New("shard verification failed")
)

// State errors
var (
// ErrCannotStartNotPending indicates reconstruction cannot start from current state.
ErrCannotStartNotPending = errors.New("cannot start reconstruction: session not in pending state")

// ErrCannotCompleteNotInProgress indicates session cannot be completed from current state.
ErrCannotCompleteNotInProgress = errors.New("cannot complete: session not in progress")

// ErrCannotMarkPartialNotInProgress indicates partial recovery cannot be marked from current state.
ErrCannotMarkPartialNotInProgress = errors.New("cannot mark partial recovery: session not in progress")

// ErrCannotFailFinalized indicates cannot fail an already finalized session.
ErrCannotFailFinalized = errors.New("cannot fail session: already complete or abandoned")

// ErrCannotAbandonComplete indicates cannot abandon a completed session.
ErrCannotAbandonComplete = errors.New("cannot abandon completed session")
)

// Nil check errors
var (
// ErrNilAttempt indicates a nil recovery attempt was provided.
ErrNilAttempt = errors.New("recovery attempt cannot be nil")
)
