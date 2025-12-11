// Package reconstruction_test provides comprehensive tests for reconstruction domain entities.
package reconstruction

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ReconstructionSession Tests
// =============================================================================

func TestNewReconstructionSession_Success(t *testing.T) {
	strategyID := "strategy-123"
	manifestID := "manifest-456"
	availableShards := []int{0, 1, 2, 3, 4}

	session, err := NewReconstructionSession(strategyID, manifestID, availableShards)

	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.False(t, session.ID.IsZero())
	assert.Equal(t, strategyID, session.StrategyID)
	assert.Equal(t, manifestID, session.ManifestID)
	assert.Equal(t, availableShards, session.AvailableShards)
	assert.Equal(t, StatusPending, session.Status)
	assert.Equal(t, 0.0, session.Progress)
	assert.NotZero(t, session.CreatedAt)
	assert.Nil(t, session.CompletedAt)
	assert.Nil(t, session.FailedAt)
}

func TestNewReconstructionSession_EmptyStrategyID(t *testing.T) {
	session, err := NewReconstructionSession("", "manifest-456", []int{0, 1, 2})

	assert.ErrorIs(t, err, ErrInvalidStrategyID)
	assert.Nil(t, session)
}

func TestNewReconstructionSession_EmptyManifestID(t *testing.T) {
	session, err := NewReconstructionSession("strategy-123", "", []int{0, 1, 2})

	assert.ErrorIs(t, err, ErrInvalidManifestID)
	assert.Nil(t, session)
}

func TestNewReconstructionSession_NoAvailableShards(t *testing.T) {
	session, err := NewReconstructionSession("strategy-123", "manifest-456", []int{})

	assert.ErrorIs(t, err, ErrNoAvailableShards)
	assert.Nil(t, session)
}

func TestReconstructionSession_StartReconstruction_Success(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})
	strategy, _ := NewRecoveryStrategy(AlgorithmReedSolomon, map[string]interface{}{"data": 10, "parity": 3}, 10)

	err := session.StartReconstruction(strategy)

	require.NoError(t, err)
	assert.Equal(t, StatusInProgress, session.Status)
	assert.Equal(t, strategy, session.RecoveryStrategy)
}

func TestReconstructionSession_StartReconstruction_NotPending(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})
	strategy, _ := NewRecoveryStrategy(AlgorithmReedSolomon, nil, 10)
	session.Status = StatusInProgress

	err := session.StartReconstruction(strategy)

	assert.ErrorIs(t, err, ErrCannotStartNotPending)
}

func TestReconstructionSession_UpdateProgress_Success(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})

	err := session.UpdateProgress(0.5)

	require.NoError(t, err)
	assert.Equal(t, 0.5, session.Progress)
}

func TestReconstructionSession_UpdateProgress_InvalidNegative(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})

	err := session.UpdateProgress(-0.1)

	assert.ErrorIs(t, err, ErrInvalidProgress)
}

func TestReconstructionSession_UpdateProgress_InvalidAboveOne(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})

	err := session.UpdateProgress(1.1)

	assert.ErrorIs(t, err, ErrInvalidProgress)
}

func TestReconstructionSession_Complete_Success(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})
	strategy, _ := NewRecoveryStrategy(AlgorithmReedSolomon, nil, 10)
	session.StartReconstruction(strategy)

	err := session.Complete()

	require.NoError(t, err)
	assert.Equal(t, StatusComplete, session.Status)
	assert.Equal(t, 1.0, session.Progress)
	assert.NotNil(t, session.CompletedAt)
}

func TestReconstructionSession_Complete_NotInProgress(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})

	err := session.Complete()

	assert.ErrorIs(t, err, ErrCannotCompleteNotInProgress)
}

func TestReconstructionSession_MarkPartialRecovery_Success(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})
	strategy, _ := NewRecoveryStrategy(AlgorithmReedSolomon, nil, 10)
	session.StartReconstruction(strategy)

	err := session.MarkPartialRecovery()

	require.NoError(t, err)
	assert.Equal(t, StatusPartialRecovery, session.Status)
	assert.NotNil(t, session.CompletedAt)
}

func TestReconstructionSession_MarkPartialRecovery_NotInProgress(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})

	err := session.MarkPartialRecovery()

	assert.ErrorIs(t, err, ErrCannotMarkPartialNotInProgress)
}

func TestReconstructionSession_Fail_Success(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})
	strategy, _ := NewRecoveryStrategy(AlgorithmReedSolomon, nil, 10)
	session.StartReconstruction(strategy)

	err := session.Fail("Insufficient shards")

	require.NoError(t, err)
	assert.Equal(t, StatusFailed, session.Status)
	assert.NotNil(t, session.FailedAt)
}

func TestReconstructionSession_Fail_AlreadyComplete(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})
	session.Status = StatusComplete

	err := session.Fail("test")

	assert.ErrorIs(t, err, ErrCannotFailFinalized)
}

func TestReconstructionSession_Fail_AlreadyAbandoned(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})
	session.Status = StatusAbandoned

	err := session.Fail("test")

	assert.ErrorIs(t, err, ErrCannotFailFinalized)
}

func TestReconstructionSession_Abandon_Success(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})

	err := session.Abandon("User cancelled")

	require.NoError(t, err)
	assert.Equal(t, StatusAbandoned, session.Status)
}

func TestReconstructionSession_Abandon_AlreadyComplete(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})
	session.Status = StatusComplete

	err := session.Abandon("test")

	assert.ErrorIs(t, err, ErrCannotAbandonComplete)
}

func TestReconstructionSession_AddAttempt_Success(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})
	attempt, _ := NewRecoveryAttempt(session.ID, []int{0, 1, 2})

	err := session.AddAttempt(attempt)

	require.NoError(t, err)
	assert.Len(t, session.Attempts, 1)
	assert.Equal(t, attempt, session.Attempts[0])
}

func TestReconstructionSession_AddAttempt_Nil(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})

	err := session.AddAttempt(nil)

	assert.ErrorIs(t, err, ErrNilAttempt)
}

func TestReconstructionSession_CanRetry_True(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})
	session.Status = StatusFailed
	attempt1, _ := NewRecoveryAttempt(session.ID, []int{0, 1, 2})
	attempt2, _ := NewRecoveryAttempt(session.ID, []int{3, 4, 5})
	session.Attempts = []*RecoveryAttempt{attempt1, attempt2}

	canRetry := session.CanRetry(3)

	assert.True(t, canRetry)
}

func TestReconstructionSession_CanRetry_FalseMaxAttemptsReached(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})
	session.Status = StatusFailed
	attempt1, _ := NewRecoveryAttempt(session.ID, []int{0, 1, 2})
	attempt2, _ := NewRecoveryAttempt(session.ID, []int{3, 4, 5})
	attempt3, _ := NewRecoveryAttempt(session.ID, []int{6, 7, 8})
	session.Attempts = []*RecoveryAttempt{attempt1, attempt2, attempt3}

	canRetry := session.CanRetry(3)

	assert.False(t, canRetry)
}

func TestReconstructionSession_CanRetry_FalseNotFailed(t *testing.T) {
	session, _ := NewReconstructionSession("strategy-123", "manifest-456", []int{0, 1, 2})
	session.Status = StatusInProgress

	canRetry := session.CanRetry(3)

	assert.False(t, canRetry)
}

// =============================================================================
// RecoveryAttempt Tests
// =============================================================================

func TestNewRecoveryAttempt_Success(t *testing.T) {
	sessionID := NewSessionID()
	shardsUsed := []int{0, 1, 2, 3, 4}

	attempt, err := NewRecoveryAttempt(sessionID, shardsUsed)

	require.NoError(t, err)
	assert.NotNil(t, attempt)
	assert.False(t, attempt.ID.IsZero())
	assert.Equal(t, sessionID, attempt.SessionID)
	assert.Equal(t, shardsUsed, attempt.ShardsUsed)
	assert.False(t, attempt.Success)
	assert.Empty(t, attempt.FailureReason)
	assert.NotZero(t, attempt.AttemptedAt)
}

func TestNewRecoveryAttempt_InvalidSessionID(t *testing.T) {
	attempt, err := NewRecoveryAttempt(SessionID{}, []int{0, 1, 2})

	assert.ErrorIs(t, err, ErrInvalidSessionID)
	assert.Nil(t, attempt)
}

func TestNewRecoveryAttempt_NoShardsUsed(t *testing.T) {
	sessionID := NewSessionID()

	attempt, err := NewRecoveryAttempt(sessionID, []int{})

	assert.ErrorIs(t, err, ErrNoShardsUsed)
	assert.Nil(t, attempt)
}

func TestRecoveryAttempt_MarkSuccess(t *testing.T) {
	sessionID := NewSessionID()
	attempt, _ := NewRecoveryAttempt(sessionID, []int{0, 1, 2})
	duration := 500 * time.Millisecond

	attempt.MarkSuccess(duration)

	assert.True(t, attempt.Success)
	assert.Equal(t, duration, attempt.Duration)
	assert.Empty(t, attempt.FailureReason)
}

func TestRecoveryAttempt_MarkFailure(t *testing.T) {
	sessionID := NewSessionID()
	attempt, _ := NewRecoveryAttempt(sessionID, []int{0, 1, 2})
	duration := 300 * time.Millisecond
	reason := "Corrupted shard detected"

	attempt.MarkFailure(reason, duration)

	assert.False(t, attempt.Success)
	assert.Equal(t, reason, attempt.FailureReason)
	assert.Equal(t, duration, attempt.Duration)
}
