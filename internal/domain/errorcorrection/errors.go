package errorcorrection

import "errors"

// Domain errors for the Error Correction bounded context.
var (
	// Message errors
	ErrInvalidMessageID = errors.New("invalid message ID")
	ErrEmptyData        = errors.New("data cannot be empty")

	// Shard errors
	ErrInvalidShardID       = errors.New("invalid shard ID")
	ErrInvalidShardIndex    = errors.New("invalid shard index")
	ErrInvalidShardCount    = errors.New("invalid shard count: must be positive")
	ErrTooManyShards        = errors.New("too many shards requested")
	ErrEmptyShardData       = errors.New("shard data cannot be empty")
	ErrNilShard             = errors.New("shard cannot be nil")
	ErrShardMessageMismatch = errors.New("shard does not belong to this message")
	ErrInsufficientShards   = errors.New("insufficient shards for recovery")
	ErrCorruptedShard       = errors.New("shard data is corrupted")

	// Redundancy errors
	ErrInvalidRedundancy = errors.New("invalid redundancy level: must be between 0 and 1")

	// Encoding/Decoding errors
	ErrEncodingFailed = errors.New("Reed-Solomon encoding failed")
	ErrDecodingFailed = errors.New("Reed-Solomon decoding failed")

	// Checksum errors
	ErrChecksumMismatch = errors.New("checksum verification failed")
	ErrMissingChecksum  = errors.New("checksum is missing")
)

// Constants for shard configuration limits
const (
	// MaxTotalShards is the maximum number of shards (data + parity) allowed.
	// This prevents excessive memory usage and processing time.
	MaxTotalShards = 256

	// DefaultDataShards is the default number of data shards.
	DefaultDataShards = 10

	// DefaultParityShards is the default number of parity shards (33% redundancy).
	DefaultParityShards = 5
)
