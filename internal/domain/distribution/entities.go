// Package distribution provides distribution domain entities.
package distribution

import (
"fmt"
"time"
)

// DistributionStrategy represents the aggregate root for distribution planning.
type DistributionStrategy struct {
ID              StrategyID
Pattern         PatternType
TotalShards     int
DataShards      int
ParityShards    int
Threshold       int
TargetMediaIDs  []string
Manifest        *ShardManifest
Status          DistributionStatus
CreatedAt       time.Time
CompletedAt     *time.Time
}

// NewDistributionStrategy creates a new distribution strategy.
func NewDistributionStrategy(
pattern PatternType,
dataShards, parityShards int,
targetMediaIDs []string,
) (*DistributionStrategy, error) {
if !pattern.IsValid() {
return nil, ErrInvalidPattern
}
if dataShards < 1 {
return nil, ErrInvalidDataShards
}
if parityShards < 0 {
return nil, ErrInvalidParityShards
}
if len(targetMediaIDs) == 0 {
return nil, ErrNoTargetMedia
}

totalShards := dataShards + parityShards
if totalShards > len(targetMediaIDs) {
return nil, ErrInsufficientMedia
}

id, err := NewStrategyID()
if err != nil {
return nil, fmt.Errorf("failed to create strategy ID: %w", err)
}

return &DistributionStrategy{
ID:             id,
Pattern:        pattern,
TotalShards:    totalShards,
DataShards:     dataShards,
ParityShards:   parityShards,
Threshold:      dataShards,
TargetMediaIDs: targetMediaIDs,
Status:         StatusPending,
CreatedAt:      time.Now(),
}, nil
}

// Complete marks the distribution as complete.
func (d *DistributionStrategy) Complete(manifest *ShardManifest) error {
if d.Status == StatusComplete {
return ErrAlreadyComplete
}
if d.Status == StatusFailed {
return ErrCannotCompleteAfterFailure
}
if manifest == nil {
return ErrManifestRequired
}

d.Manifest = manifest
d.Status = StatusComplete
now := time.Now()
d.CompletedAt = &now
return nil
}

// MarkFailed marks the distribution as failed.
func (d *DistributionStrategy) MarkFailed() error {
if d.Status == StatusComplete {
return ErrCannotFailAfterComplete
}
d.Status = StatusFailed
return nil
}

// CanRecover returns true if data can be recovered with available shards.
func (d *DistributionStrategy) CanRecover(availableShards int) bool {
return availableShards >= d.Threshold
}

// RedundancyLevel returns the redundancy percentage.
func (d *DistributionStrategy) RedundancyLevel() float64 {
if d.DataShards == 0 {
return 0.0
}
return float64(d.ParityShards) / float64(d.DataShards)
}

// ShardManifest represents metadata about distributed shards.
type ShardManifest struct {
ID              ManifestID
StrategyID      StrategyID
ShardMetadata   []ShardMetadata
TotalShards     int
RequiredShards  int
CreatedAt       time.Time
ExpiresAt       *time.Time
}

// NewShardManifest creates a new shard manifest.
func NewShardManifest(
strategyID StrategyID,
shardMetadata []ShardMetadata,
requiredShards int,
) (*ShardManifest, error) {
if strategyID.IsZero() {
return nil, ErrInvalidStrategyID
}
if len(shardMetadata) == 0 {
return nil, ErrNoShardMetadata
}
if requiredShards < 1 || requiredShards > len(shardMetadata) {
return nil, ErrInvalidThreshold
}

id, err := NewManifestID()
if err != nil {
return nil, fmt.Errorf("failed to create manifest ID: %w", err)
}

return &ShardManifest{
ID:             id,
StrategyID:     strategyID,
ShardMetadata:  shardMetadata,
TotalShards:    len(shardMetadata),
RequiredShards: requiredShards,
CreatedAt:      time.Now(),
}, nil
}

// IsExpired returns true if the manifest has expired.
func (m *ShardManifest) IsExpired() bool {
if m.ExpiresAt == nil {
return false
}
return time.Now().After(*m.ExpiresAt)
}

// GetShardByIndex retrieves shard metadata by index.
func (m *ShardManifest) GetShardByIndex(index int) (ShardMetadata, error) {
if index < 0 || index >= len(m.ShardMetadata) {
return ShardMetadata{}, ErrInvalidShardIndex
}
return m.ShardMetadata[index], nil
}

// ValidateCompleteness checks if all required shards are present.
func (m *ShardManifest) ValidateCompleteness(availableIndices []int) error {
if len(availableIndices) < m.RequiredShards {
return ErrInsufficientShards
}
return nil
}
