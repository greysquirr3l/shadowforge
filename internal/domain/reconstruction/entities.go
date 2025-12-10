// Package reconstruction provides reconstruction domain entities.
package reconstruction

import (
	"time"

	"github.com/greysquirr3l/shadowforge/pkg/logger"
	"github.com/sirupsen/logrus"
)

// ReconstructionSession represents a session for reconstructing data from distributed shards.
type ReconstructionSession struct {
	ID               SessionID
	StrategyID       string // References Distribution.StrategyID
	ManifestID       string // References Distribution.ManifestID
	AvailableShards  []int
	RecoveryStrategy RecoveryStrategy
	Progress         float64
	Status           RecoveryStatus
	Attempts         []*RecoveryAttempt
	CreatedAt        time.Time
	CompletedAt      *time.Time
	FailedAt         *time.Time
}

// NewReconstructionSession creates a new reconstruction session.
func NewReconstructionSession(strategyID, manifestID string, availableShards []int) (*ReconstructionSession, error) {
	if strategyID == "" {
		return nil, ErrInvalidStrategyID
	}
	if manifestID == "" {
		return nil, ErrInvalidManifestID
	}
	if len(availableShards) == 0 {
		return nil, ErrNoAvailableShards
	}

	session := &ReconstructionSession{
		ID:              NewSessionID(),
		StrategyID:      strategyID,
		ManifestID:      manifestID,
		AvailableShards: availableShards,
		Status:          StatusPending,
		Progress:        0.0,
		CreatedAt:       time.Now(),
	}

	logger.WithFields(logrus.Fields{
		"session_id":  session.ID.String(),
		"strategy_id": strategyID,
		"manifest_id": manifestID,
		"shard_count": len(availableShards),
	}).Info("Reconstruction session created")

	return session, nil
}

// StartReconstruction starts the reconstruction process.
func (s *ReconstructionSession) StartReconstruction(strategy RecoveryStrategy) error {
	if s.Status != StatusPending {
		return ErrCannotStartNotPending
	}

	s.RecoveryStrategy = strategy
	s.Status = StatusInProgress

	logger.WithFields(logrus.Fields{
		"session_id": s.ID.String(),
		"algorithm":  strategy.Algorithm.String(),
		"threshold":  strategy.Threshold,
	}).Info("Reconstruction started")

	return nil
}

// UpdateProgress updates the reconstruction progress.
func (s *ReconstructionSession) UpdateProgress(progress float64) error {
	if progress < 0 || progress > 1.0 {
		return ErrInvalidProgress
	}

	s.Progress = progress
	return nil
}

// Complete marks the session as successfully completed.
func (s *ReconstructionSession) Complete() error {
	if s.Status != StatusInProgress {
		return ErrCannotCompleteNotInProgress
	}

	now := time.Now()
	s.Status = StatusComplete
	s.Progress = 1.0
	s.CompletedAt = &now

	duration := now.Sub(s.CreatedAt)
	logger.WithFields(logrus.Fields{
		"session_id":   s.ID.String(),
		"attempts":     len(s.Attempts),
		"duration_sec": duration.Seconds(),
	}).Info("Reconstruction completed successfully")

	return nil
}

// MarkPartialRecovery marks the session as partially recovered.
func (s *ReconstructionSession) MarkPartialRecovery() error {
	if s.Status != StatusInProgress {
		return ErrCannotMarkPartialNotInProgress
	}

	now := time.Now()
	s.Status = StatusPartialRecovery
	s.CompletedAt = &now
	return nil
}

// Fail marks the session as failed.
func (s *ReconstructionSession) Fail(reason string) error {
	if s.Status == StatusComplete || s.Status == StatusAbandoned {
		return ErrCannotFailFinalized
	}

	now := time.Now()
	s.Status = StatusFailed
	s.FailedAt = &now

	logger.WithFields(logrus.Fields{
		"session_id": s.ID.String(),
		"reason":     reason,
		"attempts":   len(s.Attempts),
	}).Error("Reconstruction failed")

	return nil
}

// Abandon abandons the reconstruction session.
func (s *ReconstructionSession) Abandon(reason string) error {
	if s.Status == StatusComplete {
		return ErrCannotAbandonComplete
	}

	s.Status = StatusAbandoned
	return nil
}

// AddAttempt adds a recovery attempt to the session.
func (s *ReconstructionSession) AddAttempt(attempt *RecoveryAttempt) error {
	if attempt == nil {
		return ErrNilAttempt
	}

	s.Attempts = append(s.Attempts, attempt)
	return nil
}

// CanRetry checks if another recovery attempt can be made.
func (s *ReconstructionSession) CanRetry(maxAttempts int) bool {
	return s.Status == StatusFailed && len(s.Attempts) < maxAttempts
}

// RecoveryAttempt represents a single attempt to reconstruct data.
type RecoveryAttempt struct {
	ID            AttemptID
	SessionID     SessionID
	ShardsUsed    []int
	Success       bool
	FailureReason string
	Duration      time.Duration
	AttemptedAt   time.Time
}

// NewRecoveryAttempt creates a new recovery attempt.
func NewRecoveryAttempt(sessionID SessionID, shardsUsed []int) (*RecoveryAttempt, error) {
	if sessionID.IsZero() {
		return nil, ErrInvalidSessionID
	}
	if len(shardsUsed) == 0 {
		return nil, ErrNoShardsUsed
	}

	attempt := &RecoveryAttempt{
		ID:          NewAttemptID(),
		SessionID:   sessionID,
		ShardsUsed:  shardsUsed,
		AttemptedAt: time.Now(),
	}

	logger.WithFields(logrus.Fields{
		"attempt_id":  attempt.ID.String(),
		"session_id":  sessionID.String(),
		"shard_count": len(shardsUsed),
	}).Debug("Recovery attempt created")

	return attempt, nil
}

// MarkSuccess marks the attempt as successful.
func (a *RecoveryAttempt) MarkSuccess(duration time.Duration) {
	a.Success = true
	a.Duration = duration
}

// MarkFailure marks the attempt as failed.
func (a *RecoveryAttempt) MarkFailure(reason string, duration time.Duration) {
	a.Success = false
	a.FailureReason = reason
	a.Duration = duration
}
