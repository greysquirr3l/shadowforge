// Package reconstruction provides reconstruction domain services.
package reconstruction

import "context"

// Service defines the interface for reconstruction domain operations.
type Service interface {
	// StartReconstruction initiates a new reconstruction session.
	StartReconstruction(ctx context.Context, strategyID, manifestID string, availableShards []int, algorithm RecoveryAlgorithm) (*ReconstructionSession, error)

	// VerifyShards validates the integrity of shard data.
	VerifyShards(ctx context.Context, sessionID SessionID, shardData map[int][]byte) ([]ShardVerification, error)

	// ReconstructData performs the actual data reconstruction from shards.
	ReconstructData(ctx context.Context, sessionID SessionID, shardData map[int][]byte) ([]byte, error)

	// ValidateRecovery validates the reconstructed data against expected checksum.
	ValidateRecovery(ctx context.Context, sessionID SessionID, reconstructedData []byte, expectedChecksum string) error

	// RetryReconstruction attempts reconstruction with a different algorithm.
	RetryReconstruction(ctx context.Context, sessionID SessionID, newAlgorithm RecoveryAlgorithm) (*RecoveryAttempt, error)

	// AbandonSession abandons a reconstruction session.
	AbandonSession(ctx context.Context, sessionID SessionID, reason string) error

	// GetSessionStatus retrieves the current status of a reconstruction session.
	GetSessionStatus(ctx context.Context, sessionID SessionID) (RecoveryStatus, error)

	// GetRecoveryProgress retrieves detailed progress information for a session.
	GetRecoveryProgress(ctx context.Context, sessionID SessionID) (*RecoveryProgress, error)

	// EstimateRecoveryTime estimates time to complete reconstruction.
	EstimateRecoveryTime(ctx context.Context, sessionID SessionID) (int64, error)
}
