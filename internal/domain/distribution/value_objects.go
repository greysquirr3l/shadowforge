// Package distribution provides distribution domain value objects.
package distribution

import (
	"fmt"

	"github.com/google/uuid"
)

// StrategyID represents a unique identifier for a distribution strategy.
type StrategyID struct {
	value uuid.UUID
}

// NewStrategyID creates a new strategy ID.
func NewStrategyID() (StrategyID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return StrategyID{}, fmt.Errorf("failed to generate strategy ID: %w", err)
	}
	return StrategyID{value: id}, nil
}

// ParseStrategyID parses a string into a StrategyID.
func ParseStrategyID(s string) (StrategyID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return StrategyID{}, ErrInvalidStrategyID
	}
	return StrategyID{value: id}, nil
}

// String returns the string representation.
func (id StrategyID) String() string {
	return id.value.String()
}

// IsZero returns true if this is a zero value.
func (id StrategyID) IsZero() bool {
	return id.value == uuid.Nil
}

// ManifestID represents a unique identifier for a shard manifest.
type ManifestID struct {
	value uuid.UUID
}

// NewManifestID creates a new manifest ID.
func NewManifestID() (ManifestID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return ManifestID{}, fmt.Errorf("failed to generate manifest ID: %w", err)
	}
	return ManifestID{value: id}, nil
}

// ParseManifestID parses a string into a ManifestID.
func ParseManifestID(s string) (ManifestID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ManifestID{}, ErrInvalidManifestID
	}
	return ManifestID{value: id}, nil
}

// String returns the string representation.
func (id ManifestID) String() string {
	return id.value.String()
}

// IsZero returns true if this is a zero value.
func (id ManifestID) IsZero() bool {
	return id.value == uuid.Nil
}

// PatternType represents distribution patterns.
type PatternType string

const (
	PatternOneToOne   PatternType = "one_to_one"
	PatternOneToMany  PatternType = "one_to_many"
	PatternManyToOne  PatternType = "many_to_one"
	PatternManyToMany PatternType = "many_to_many"
)

// IsValid returns true if the pattern type is valid.
func (p PatternType) IsValid() bool {
	switch p {
	case PatternOneToOne, PatternOneToMany, PatternManyToOne, PatternManyToMany:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (p PatternType) String() string {
	return string(p)
}

// DistributionStatus represents the status of a distribution.
type DistributionStatus string

const (
	StatusPending    DistributionStatus = "pending"
	StatusInProgress DistributionStatus = "in_progress"
	StatusComplete   DistributionStatus = "complete"
	StatusFailed     DistributionStatus = "failed"
	StatusRevoked    DistributionStatus = "revoked"
)

// IsValid returns true if the status is valid.
func (s DistributionStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusComplete, StatusFailed, StatusRevoked:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (s DistributionStatus) String() string {
	return string(s)
}

// ShardMetadata represents metadata for a single shard.
type ShardMetadata struct {
	Index       int
	Size        int64
	Checksum    string
	MediaID     string
	RecipientID string
}

// NewShardMetadata creates new shard metadata.
func NewShardMetadata(
	index int,
	size int64,
	checksum string,
	mediaID string,
	recipientID string,
) (ShardMetadata, error) {
	if index < 0 {
		return ShardMetadata{}, ErrInvalidShardIndex
	}
	if size <= 0 {
		return ShardMetadata{}, fmt.Errorf("invalid shard size: %d", size)
	}
	if checksum == "" {
		return ShardMetadata{}, fmt.Errorf("checksum is required")
	}
	if mediaID == "" {
		return ShardMetadata{}, fmt.Errorf("media ID is required")
	}

	return ShardMetadata{
		Index:       index,
		Size:        size,
		Checksum:    checksum,
		MediaID:     mediaID,
		RecipientID: recipientID,
	}, nil
}

// Validate validates the shard metadata.
func (m ShardMetadata) Validate() error {
	if m.Index < 0 {
		return ErrInvalidShardIndex
	}
	if m.Size <= 0 {
		return fmt.Errorf("invalid shard size: %d", m.Size)
	}
	if m.Checksum == "" {
		return fmt.Errorf("checksum is required")
	}
	if m.MediaID == "" {
		return fmt.Errorf("media ID is required")
	}
	return nil
}
