// Package distribution provides distribution domain errors.
package distribution

import "errors"

var (
	// ErrInvalidStrategyID indicates an invalid or zero strategy ID.
	ErrInvalidStrategyID = errors.New("invalid strategy ID")

	// ErrInvalidManifestID indicates an invalid or zero manifest ID.
	ErrInvalidManifestID = errors.New("invalid manifest ID")

	// ErrInvalidPattern indicates an invalid distribution pattern.
	ErrInvalidPattern = errors.New("invalid distribution pattern")

	// ErrInvalidDataShards indicates invalid data shard count.
	ErrInvalidDataShards = errors.New("data shard count must be at least 1")

	// ErrInvalidParityShards indicates invalid parity shard count.
	ErrInvalidParityShards = errors.New("parity shard count must be non-negative")

	// ErrNoTargetMedia indicates no target media IDs provided.
	ErrNoTargetMedia = errors.New("no target media IDs provided")

	// ErrInsufficientMedia indicates not enough media for all shards.
	ErrInsufficientMedia = errors.New("insufficient target media for all shards")

	// ErrAlreadyComplete indicates strategy already marked complete.
	ErrAlreadyComplete = errors.New("distribution already complete")

	// ErrCannotCompleteAfterFailure indicates cannot complete after failure.
	ErrCannotCompleteAfterFailure = errors.New("cannot complete distribution after failure")

	// ErrManifestRequired indicates manifest required for completion.
	ErrManifestRequired = errors.New("manifest required to complete distribution")

	// ErrCannotFailAfterComplete indicates cannot fail after completion.
	ErrCannotFailAfterComplete = errors.New("cannot fail distribution after completion")

	// ErrNoShardMetadata indicates no shard metadata provided.
	ErrNoShardMetadata = errors.New("no shard metadata provided")

	// ErrInvalidThreshold indicates invalid required shards threshold.
	ErrInvalidThreshold = errors.New("invalid required shards threshold")

	// ErrInvalidShardIndex indicates shard index out of bounds.
	ErrInvalidShardIndex = errors.New("shard index out of bounds")

	// ErrInsufficientShards indicates insufficient shards for recovery.
	ErrInsufficientShards = errors.New("insufficient shards for recovery")
)
