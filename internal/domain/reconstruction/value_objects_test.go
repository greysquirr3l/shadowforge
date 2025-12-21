// Package reconstruction_test provides comprehensive tests for reconstruction value objects.
package reconstruction

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// SessionID Tests
// =============================================================================

func TestNewSessionID(t *testing.T) {
	id := NewSessionID()

	assert.False(t, id.IsZero())
	assert.NotEmpty(t, id.String())
}

func TestSessionID_Parse_Success(t *testing.T) {
	original := NewSessionID()

	parsed, err := ParseSessionID(original.String())

	require.NoError(t, err)
	assert.Equal(t, original, parsed)
}

func TestSessionID_Parse_EmptyString(t *testing.T) {
	_, err := ParseSessionID("")

	assert.ErrorIs(t, err, ErrInvalidSessionID)
}

func TestSessionID_Parse_InvalidFormat(t *testing.T) {
	_, err := ParseSessionID("not-a-uuid")

	assert.ErrorIs(t, err, ErrInvalidSessionID)
}

func TestSessionID_IsZero(t *testing.T) {
	zero := SessionID{}

	assert.True(t, zero.IsZero())
}

func TestSessionID_String(t *testing.T) {
	id := NewSessionID()

	str := id.String()

	assert.NotEmpty(t, str)
	assert.Len(t, str, 36) // UUID format
}

// =============================================================================
// AttemptID Tests
// =============================================================================

func TestNewAttemptID(t *testing.T) {
	id := NewAttemptID()

	assert.False(t, id.IsZero())
	assert.NotEmpty(t, id.String())
}

func TestAttemptID_Parse_Success(t *testing.T) {
	original := NewAttemptID()

	parsed, err := ParseAttemptID(original.String())

	require.NoError(t, err)
	assert.Equal(t, original, parsed)
}

func TestAttemptID_Parse_EmptyString(t *testing.T) {
	_, err := ParseAttemptID("")

	assert.ErrorIs(t, err, ErrInvalidAttemptID)
}

func TestAttemptID_Parse_InvalidFormat(t *testing.T) {
	_, err := ParseAttemptID("not-a-uuid")

	assert.ErrorIs(t, err, ErrInvalidAttemptID)
}

func TestAttemptID_IsZero(t *testing.T) {
	zero := AttemptID{}

	assert.True(t, zero.IsZero())
}

func TestAttemptID_String(t *testing.T) {
	id := NewAttemptID()

	str := id.String()

	assert.NotEmpty(t, str)
	assert.Len(t, str, 36) // UUID format
}

// =============================================================================
// RecoveryAlgorithm Tests
// =============================================================================

func TestRecoveryAlgorithm_IsValid_ReedSolomon(t *testing.T) {
	assert.True(t, AlgorithmReedSolomon.IsValid())
}

func TestRecoveryAlgorithm_IsValid_Shamir(t *testing.T) {
	assert.True(t, AlgorithmShamir.IsValid())
}

func TestRecoveryAlgorithm_IsValid_Rabin(t *testing.T) {
	assert.True(t, AlgorithmRabin.IsValid())
}

func TestRecoveryAlgorithm_IsValid_Invalid(t *testing.T) {
	invalid := RecoveryAlgorithm("invalid")

	assert.False(t, invalid.IsValid())
}

func TestRecoveryAlgorithm_String_ReedSolomon(t *testing.T) {
	assert.Equal(t, "reed_solomon", AlgorithmReedSolomon.String())
}

func TestRecoveryAlgorithm_String_Shamir(t *testing.T) {
	assert.Equal(t, "shamir", AlgorithmShamir.String())
}

func TestRecoveryAlgorithm_String_Rabin(t *testing.T) {
	assert.Equal(t, "rabin", AlgorithmRabin.String())
}

// =============================================================================
// RecoveryStrategy Tests
// =============================================================================

func TestNewRecoveryStrategy_ReedSolomon(t *testing.T) {
	params := map[string]interface{}{"data": 10, "parity": 3}

	strategy, err := NewRecoveryStrategy(AlgorithmReedSolomon, params, 10)

	require.NoError(t, err)
	assert.Equal(t, AlgorithmReedSolomon, strategy.Algorithm)
	assert.Equal(t, params, strategy.Parameters)
	assert.Equal(t, 10, strategy.Threshold)
}

func TestNewRecoveryStrategy_Shamir(t *testing.T) {
	params := map[string]interface{}{"shares": 5, "threshold": 3}

	strategy, err := NewRecoveryStrategy(AlgorithmShamir, params, 3)

	require.NoError(t, err)
	assert.Equal(t, AlgorithmShamir, strategy.Algorithm)
}

func TestNewRecoveryStrategy_Rabin(t *testing.T) {
	strategy, err := NewRecoveryStrategy(AlgorithmRabin, nil, 8)

	require.NoError(t, err)
	assert.Equal(t, AlgorithmRabin, strategy.Algorithm)
}

func TestNewRecoveryStrategy_InvalidAlgorithm(t *testing.T) {
	_, err := NewRecoveryStrategy(RecoveryAlgorithm("invalid"), nil, 10)

	assert.ErrorIs(t, err, ErrInvalidStrategy)
}

func TestNewRecoveryStrategy_ZeroThreshold(t *testing.T) {
	_, err := NewRecoveryStrategy(AlgorithmReedSolomon, nil, 0)

	assert.ErrorIs(t, err, ErrInvalidStrategy)
}

func TestNewRecoveryStrategy_NegativeThreshold(t *testing.T) {
	_, err := NewRecoveryStrategy(AlgorithmReedSolomon, nil, -1)

	assert.ErrorIs(t, err, ErrInvalidStrategy)
}

func TestRecoveryStrategy_IsValid(t *testing.T) {
	strategy, _ := NewRecoveryStrategy(AlgorithmReedSolomon, nil, 10)

	assert.True(t, strategy.IsValid())
}

func TestRecoveryStrategy_String(t *testing.T) {
	strategy, _ := NewRecoveryStrategy(AlgorithmReedSolomon, nil, 10)

	str := strategy.String()

	assert.Contains(t, str, "reed_solomon")
	assert.Contains(t, str, "10")
}

// =============================================================================
// RecoveryStatus Tests
// =============================================================================

func TestRecoveryStatus_IsValid_Pending(t *testing.T) {
	assert.True(t, StatusPending.IsValid())
}

func TestRecoveryStatus_IsValid_InProgress(t *testing.T) {
	assert.True(t, StatusInProgress.IsValid())
}

func TestRecoveryStatus_IsValid_Complete(t *testing.T) {
	assert.True(t, StatusComplete.IsValid())
}

func TestRecoveryStatus_IsValid_Failed(t *testing.T) {
	assert.True(t, StatusFailed.IsValid())
}

func TestRecoveryStatus_IsValid_PartialRecovery(t *testing.T) {
	assert.True(t, StatusPartialRecovery.IsValid())
}

func TestRecoveryStatus_IsValid_Abandoned(t *testing.T) {
	assert.True(t, StatusAbandoned.IsValid())
}

func TestRecoveryStatus_IsValid_Invalid(t *testing.T) {
	invalid := RecoveryStatus("invalid")

	assert.False(t, invalid.IsValid())
}

func TestRecoveryStatus_IsFinal_Pending(t *testing.T) {
	assert.False(t, StatusPending.IsFinal())
}

func TestRecoveryStatus_IsFinal_InProgress(t *testing.T) {
	assert.False(t, StatusInProgress.IsFinal())
}

func TestRecoveryStatus_IsFinal_Complete(t *testing.T) {
	assert.True(t, StatusComplete.IsFinal())
}

func TestRecoveryStatus_IsFinal_Failed(t *testing.T) {
	assert.True(t, StatusFailed.IsFinal())
}

func TestRecoveryStatus_IsFinal_PartialRecovery(t *testing.T) {
	assert.True(t, StatusPartialRecovery.IsFinal())
}

func TestRecoveryStatus_IsFinal_Abandoned(t *testing.T) {
	assert.True(t, StatusAbandoned.IsFinal())
}

func TestRecoveryStatus_String_Pending(t *testing.T) {
	assert.Equal(t, "pending", StatusPending.String())
}

func TestRecoveryStatus_String_InProgress(t *testing.T) {
	assert.Equal(t, "in_progress", StatusInProgress.String())
}

func TestRecoveryStatus_String_Complete(t *testing.T) {
	assert.Equal(t, "complete", StatusComplete.String())
}

func TestRecoveryStatus_String_Failed(t *testing.T) {
	assert.Equal(t, "failed", StatusFailed.String())
}

func TestRecoveryStatus_String_PartialRecovery(t *testing.T) {
	assert.Equal(t, "partial_recovery", StatusPartialRecovery.String())
}

func TestRecoveryStatus_String_Abandoned(t *testing.T) {
	assert.Equal(t, "abandoned", StatusAbandoned.String())
}

// =============================================================================
// ShardVerification Tests
// =============================================================================

func TestNewShardVerification_Success(t *testing.T) {
	verification, err := NewShardVerification(5, true, 0.95)

	require.NoError(t, err)
	assert.Equal(t, 5, verification.ShardIndex)
	assert.True(t, verification.ChecksumValid)
	assert.Equal(t, 0.95, verification.IntegrityScore)
	assert.NotZero(t, verification.VerifiedAt)
}

func TestNewShardVerification_ValidScore(t *testing.T) {
	tests := []struct {
		name  string
		score float64
	}{
		{"Zero", 0.0},
		{"Half", 0.5},
		{"One", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verification, err := NewShardVerification(0, true, tt.score)

			require.NoError(t, err)
			assert.Equal(t, tt.score, verification.IntegrityScore)
		})
	}
}

func TestNewShardVerification_NegativeIndex(t *testing.T) {
	_, err := NewShardVerification(-1, true, 0.5)

	assert.Error(t, err)
}

func TestNewShardVerification_ScoreBelowZero(t *testing.T) {
	_, err := NewShardVerification(0, true, -0.1)

	assert.Error(t, err)
}

func TestNewShardVerification_ScoreAboveOne(t *testing.T) {
	_, err := NewShardVerification(0, true, 1.1)

	assert.Error(t, err)
}

func TestShardVerification_IsValid(t *testing.T) {
	verification, _ := NewShardVerification(0, true, 0.8)

	assert.True(t, verification.IsValid())
}

func TestShardVerification_String(t *testing.T) {
	verification, _ := NewShardVerification(5, true, 0.95)
	str := verification.String()

	assert.Contains(t, str, "5")
	assert.Contains(t, str, "0.95")
}

// =============================================================================
// RecoveryProgress Tests
// =============================================================================

func TestNewRecoveryProgress_Success(t *testing.T) {
	progress, err := NewRecoveryProgress(15, 10, 2)

	require.NoError(t, err)
	assert.Equal(t, 15, progress.TotalShards)
	assert.Equal(t, 10, progress.VerifiedShards)
	assert.Equal(t, 2, progress.FailedShards)
	assert.InDelta(t, 0.8, progress.PercentComplete, 0.01) // 12/15 = 0.8
}

func TestNewRecoveryProgress_AllVerified(t *testing.T) {
	progress, err := NewRecoveryProgress(10, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 1.0, progress.PercentComplete)
}

func TestNewRecoveryProgress_NoneVerified(t *testing.T) {
	progress, err := NewRecoveryProgress(10, 0, 0)

	require.NoError(t, err)
	assert.Equal(t, 0.0, progress.PercentComplete)
}

func TestNewRecoveryProgress_ZeroTotal(t *testing.T) {
	_, err := NewRecoveryProgress(0, 0, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "total shards must be positive")
	_, err = NewRecoveryProgress(-1, 0, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "total shards must be positive")
	_, err = NewRecoveryProgress(10, -1, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be non-negative")
	_, err = NewRecoveryProgress(10, 0, -1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be non-negative")
	_, err = NewRecoveryProgress(10, 8, 3) // 11 > 10

	assert.ErrorIs(t, err, ErrInvalidProgress)
}

func TestRecoveryProgress_Calculate(t *testing.T) {
	progress, _ := NewRecoveryProgress(15, 10, 2)

	progress.Calculate()

	assert.InDelta(t, 0.8, progress.PercentComplete, 0.01)
}

func TestRecoveryProgress_Calculate_ProgressChange(t *testing.T) {
	progress, _ := NewRecoveryProgress(15, 5, 0)
	assert.InDelta(t, 0.333, progress.PercentComplete, 0.01)

	progress.VerifiedShards = 10
	progress.Calculate()

	assert.InDelta(t, 0.666, progress.PercentComplete, 0.01)
}

func TestRecoveryProgress_IsValid(t *testing.T) {
	progress, _ := NewRecoveryProgress(10, 8, 1)

	assert.True(t, progress.IsValid())
}

func TestRecoveryProgress_String(t *testing.T) {
	progress, _ := NewRecoveryProgress(15, 10, 2)

	str := progress.String()

	assert.Contains(t, str, "10")
	assert.Contains(t, str, "15")
	assert.Contains(t, str, "80.0") // 10 verified + 2 failed = 12/15 = 80% complete
}

// =============================================================================
// ShardCollection Tests
// =============================================================================

func TestNewShardCollection_Success(t *testing.T) {
	indices := []int{0, 1, 2, 3, 4}

	collection, err := NewShardCollection(indices)

	require.NoError(t, err)
	assert.Equal(t, indices, collection.Indices())
}

func TestNewShardCollection_Empty(t *testing.T) {
	_, err := NewShardCollection([]int{})

	assert.ErrorIs(t, err, ErrNoAvailableShards)
}

func TestNewShardCollection_NegativeIndex(t *testing.T) {
	_, err := NewShardCollection([]int{0, 1, -1, 3})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be non-negative")
}

func TestShardCollection_Indices_ReturnsCopy(t *testing.T) {
	original := []int{0, 1, 2}
	collection, _ := NewShardCollection(original)

	indices := collection.Indices()
	indices[0] = 999

	assert.Equal(t, 0, collection.Indices()[0])
}

func TestShardCollection_Count(t *testing.T) {
	collection, _ := NewShardCollection([]int{0, 1, 2, 3, 4})

	assert.Equal(t, 5, collection.Count())
}

func TestShardCollection_Contains_Found(t *testing.T) {
	collection, _ := NewShardCollection([]int{0, 1, 2, 3, 4})

	assert.True(t, collection.Contains(2))
}

func TestShardCollection_Contains_NotFound(t *testing.T) {
	collection, _ := NewShardCollection([]int{0, 1, 2, 3, 4})

	assert.False(t, collection.Contains(10))
}

func TestShardCollection_String(t *testing.T) {
	collection, _ := NewShardCollection([]int{0, 1, 2})

	str := collection.String()

	assert.Contains(t, str, "3")
	assert.Contains(t, str, "0")
	assert.Contains(t, str, "2")
}
