// Package reconstruction provides domain logic for shard reconstruction and data recovery.
package reconstruction

import (
"fmt"
"strings"
"time"

"github.com/google/uuid"
)

// SessionID is a unique identifier for a reconstruction session.
type SessionID struct {
value string
}

// NewSessionID creates a new session identifier.
func NewSessionID() SessionID {
return SessionID{value: uuid.New().String()}
}

// ParseSessionID parses a session ID from a string.
func ParseSessionID(s string) (SessionID, error) {
if s == "" {
return SessionID{}, ErrInvalidSessionID
}
if _, err := uuid.Parse(s); err != nil {
return SessionID{}, fmt.Errorf("invalid session ID format: %w", err)
}
return SessionID{value: s}, nil
}

// IsZero returns true if the session ID is zero value.
func (id SessionID) IsZero() bool {
return id.value == ""
}

// String returns the string representation of the session ID.
func (id SessionID) String() string {
return id.value
}

// AttemptID is a unique identifier for a recovery attempt.
type AttemptID struct {
value string
}

// NewAttemptID creates a new attempt identifier.
func NewAttemptID() AttemptID {
return AttemptID{value: uuid.New().String()}
}

// ParseAttemptID parses an attempt ID from a string.
func ParseAttemptID(s string) (AttemptID, error) {
if s == "" {
return AttemptID{}, ErrInvalidAttemptID
}
if _, err := uuid.Parse(s); err != nil {
return AttemptID{}, fmt.Errorf("invalid attempt ID format: %w", err)
}
return AttemptID{value: s}, nil
}

// IsZero returns true if the attempt ID is zero value.
func (id AttemptID) IsZero() bool {
return id.value == ""
}

// String returns the string representation of the attempt ID.
func (id AttemptID) String() string {
return id.value
}

// RecoveryAlgorithm represents the algorithm used for data recovery.
type RecoveryAlgorithm string

const (
// AlgorithmReedSolomon uses Reed-Solomon erasure coding for recovery.
AlgorithmReedSolomon RecoveryAlgorithm = "reed_solomon"

// AlgorithmShamir uses Shamir's Secret Sharing for recovery.
AlgorithmShamir RecoveryAlgorithm = "shamir"

// AlgorithmRabin uses Rabin's Information Dispersal Algorithm.
AlgorithmRabin RecoveryAlgorithm = "rabin"
)

// IsValid returns true if the algorithm is valid.
func (a RecoveryAlgorithm) IsValid() bool {
switch a {
case AlgorithmReedSolomon, AlgorithmShamir, AlgorithmRabin:
return true
default:
return false
}
}

// String returns the string representation of the algorithm.
func (a RecoveryAlgorithm) String() string {
return string(a)
}

// RecoveryStrategy defines the strategy for data recovery.
type RecoveryStrategy struct {
Algorithm  RecoveryAlgorithm
Parameters map[string]interface{}
Threshold  int
}

// NewRecoveryStrategy creates a new recovery strategy.
func NewRecoveryStrategy(algorithm RecoveryAlgorithm, params map[string]interface{}, threshold int) (RecoveryStrategy, error) {
if !algorithm.IsValid() {
return RecoveryStrategy{}, ErrInvalidStrategy
}
if threshold <= 0 {
return RecoveryStrategy{}, fmt.Errorf("threshold must be positive: %d", threshold)
}
if params == nil {
params = make(map[string]interface{})
}

return RecoveryStrategy{
Algorithm:  algorithm,
Parameters: params,
Threshold:  threshold,
}, nil
}

// IsValid returns true if the strategy is valid.
func (s RecoveryStrategy) IsValid() bool {
return s.Algorithm.IsValid() && s.Threshold > 0
}

// String returns the string representation of the strategy.
func (s RecoveryStrategy) String() string {
return fmt.Sprintf("RecoveryStrategy{Algorithm=%s, Threshold=%d, Params=%d}",
s.Algorithm, s.Threshold, len(s.Parameters))
}

// RecoveryStatus represents the status of a reconstruction session.
type RecoveryStatus string

const (
// StatusPending indicates the session is pending start.
StatusPending RecoveryStatus = "pending"

// StatusInProgress indicates reconstruction is in progress.
StatusInProgress RecoveryStatus = "in_progress"

// StatusComplete indicates reconstruction completed successfully.
StatusComplete RecoveryStatus = "complete"

// StatusFailed indicates reconstruction failed.
StatusFailed RecoveryStatus = "failed"

// StatusPartialRecovery indicates partial data was recovered.
StatusPartialRecovery RecoveryStatus = "partial_recovery"

// StatusAbandoned indicates the session was abandoned.
StatusAbandoned RecoveryStatus = "abandoned"
)

// IsValid returns true if the status is valid.
func (s RecoveryStatus) IsValid() bool {
switch s {
case StatusPending, StatusInProgress, StatusComplete, StatusFailed, StatusPartialRecovery, StatusAbandoned:
return true
default:
return false
}
}

// IsFinal returns true if the status represents a final state.
func (s RecoveryStatus) IsFinal() bool {
return s == StatusComplete || s == StatusFailed || s == StatusPartialRecovery || s == StatusAbandoned
}

// String returns the string representation of the status.
func (s RecoveryStatus) String() string {
return string(s)
}

// ShardVerification represents the result of verifying a shard's integrity.
type ShardVerification struct {
ShardIndex     int
ChecksumValid  bool
IntegrityScore float64
VerifiedAt     time.Time
}

// NewShardVerification creates a new shard verification result.
func NewShardVerification(index int, checksumValid bool, score float64) (ShardVerification, error) {
if index < 0 {
return ShardVerification{}, fmt.Errorf("shard index must be non-negative: %d", index)
}
if score < 0.0 || score > 1.0 {
return ShardVerification{}, fmt.Errorf("integrity score must be between 0.0 and 1.0: %f", score)
}

return ShardVerification{
ShardIndex:     index,
ChecksumValid:  checksumValid,
IntegrityScore: score,
VerifiedAt:     time.Now(),
}, nil
}

// IsValid returns true if the verification result is valid.
func (v ShardVerification) IsValid() bool {
return v.ShardIndex >= 0 && v.IntegrityScore >= 0.0 && v.IntegrityScore <= 1.0
}

// String returns the string representation of the verification.
func (v ShardVerification) String() string {
return fmt.Sprintf("ShardVerification{Index=%d, Valid=%t, Score=%.2f}",
v.ShardIndex, v.ChecksumValid, v.IntegrityScore)
}

// RecoveryProgress tracks the progress of a reconstruction session.
type RecoveryProgress struct {
TotalShards    int
VerifiedShards int
FailedShards   int
PercentComplete float64
}

// NewRecoveryProgress creates a new recovery progress tracker.
func NewRecoveryProgress(total, verified, failed int) (RecoveryProgress, error) {
if total <= 0 {
return RecoveryProgress{}, fmt.Errorf("total shards must be positive: %d", total)
}
if verified < 0 || failed < 0 {
return RecoveryProgress{}, fmt.Errorf("verified and failed counts must be non-negative")
}
if verified+failed > total {
return RecoveryProgress{}, fmt.Errorf("verified + failed (%d) exceeds total (%d)", verified+failed, total)
}

progress := RecoveryProgress{
TotalShards:    total,
VerifiedShards: verified,
FailedShards:   failed,
}
progress.Calculate()

return progress, nil
}

// Calculate updates the percent complete based on verified shards.
func (p *RecoveryProgress) Calculate() {
if p.TotalShards > 0 {
p.PercentComplete = float64(p.VerifiedShards) / float64(p.TotalShards)
} else {
p.PercentComplete = 0.0
}
}

// IsValid returns true if the progress is valid.
func (p RecoveryProgress) IsValid() bool {
return p.TotalShards > 0 &&
p.VerifiedShards >= 0 &&
p.FailedShards >= 0 &&
p.VerifiedShards+p.FailedShards <= p.TotalShards
}

// String returns the string representation of the progress.
func (p RecoveryProgress) String() string {
return fmt.Sprintf("RecoveryProgress{%d/%d verified, %d failed, %.1f%% complete}",
p.VerifiedShards, p.TotalShards, p.FailedShards, p.PercentComplete*100)
}

// ShardCollection represents a collection of shard indices.
type ShardCollection struct {
indices []int
}

// NewShardCollection creates a new shard collection.
func NewShardCollection(indices []int) (ShardCollection, error) {
if len(indices) == 0 {
return ShardCollection{}, ErrNoAvailableShards
}

// Validate no negative indices
for _, idx := range indices {
if idx < 0 {
return ShardCollection{}, fmt.Errorf("shard index must be non-negative: %d", idx)
}
}

// Create a copy to prevent external modification
copied := make([]int, len(indices))
copy(copied, indices)

return ShardCollection{indices: copied}, nil
}

// Indices returns a copy of the shard indices.
func (c ShardCollection) Indices() []int {
copied := make([]int, len(c.indices))
copy(copied, c.indices)
return copied
}

// Count returns the number of shards in the collection.
func (c ShardCollection) Count() int {
return len(c.indices)
}

// Contains returns true if the collection contains the given index.
func (c ShardCollection) Contains(index int) bool {
for _, idx := range c.indices {
if idx == index {
return true
}
}
return false
}

// String returns the string representation of the collection.
func (c ShardCollection) String() string {
if len(c.indices) == 0 {
return "ShardCollection{empty}"
}

// Format as comma-separated list
strs := make([]string, len(c.indices))
for i, idx := range c.indices {
strs[i] = fmt.Sprintf("%d", idx)
}

return fmt.Sprintf("ShardCollection{%s}", strings.Join(strs, ", "))
}
