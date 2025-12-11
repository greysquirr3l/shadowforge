// Package reconstruction provides reconstruction domain repository interfaces.
package reconstruction

import "context"

// Repository defines the interface for reconstruction persistence.
type Repository interface {
	// SaveSession persists a reconstruction session.
	SaveSession(ctx context.Context, session *ReconstructionSession) error

	// GetSession retrieves a reconstruction session by ID.
	GetSession(ctx context.Context, sessionID SessionID) (*ReconstructionSession, error)

	// GetSessionByStrategy retrieves a reconstruction session by strategy ID.
	GetSessionByStrategy(ctx context.Context, strategyID string) (*ReconstructionSession, error)

	// ListSessions retrieves reconstruction sessions with optional status filter.
	ListSessions(ctx context.Context, status RecoveryStatus, limit, offset int) ([]*ReconstructionSession, error)

	// DeleteSession removes a reconstruction session.
	DeleteSession(ctx context.Context, sessionID SessionID) error

	// SaveAttempt persists a recovery attempt.
	SaveAttempt(ctx context.Context, attempt *RecoveryAttempt) error

	// GetAttemptsBySession retrieves all recovery attempts for a session.
	GetAttemptsBySession(ctx context.Context, sessionID SessionID) ([]*RecoveryAttempt, error)

	// GetAttempt retrieves a specific recovery attempt by ID.
	GetAttempt(ctx context.Context, attemptID AttemptID) (*RecoveryAttempt, error)
}
