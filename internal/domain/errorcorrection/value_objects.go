package errorcorrection

import (
	"fmt"

	"github.com/google/uuid"
)

// MessageID is a value object representing a unique identifier for a protected message.
type MessageID struct {
	value string
}

// NewMessageID creates a new MessageID with validation.
func NewMessageID(id string) (MessageID, error) {
	if id == "" {
		return MessageID{}, ErrInvalidMessageID
	}
	return MessageID{value: id}, nil
}

// GenerateMessageID creates a new random MessageID using UUID v4.
func GenerateMessageID() MessageID {
	return MessageID{value: uuid.New().String()}
}

// String returns the string representation of the MessageID.
func (m MessageID) String() string {
	return m.value
}

// IsZero checks if the MessageID is the zero value.
func (m MessageID) IsZero() bool {
	return m.value == ""
}

// Equals checks if two MessageIDs are equal.
func (m MessageID) Equals(other MessageID) bool {
	return m.value == other.value
}

// ShardID is a value object representing a unique identifier for a shard.
type ShardID struct {
	value string
}

// NewShardID creates a new ShardID with validation.
func NewShardID(id string) (ShardID, error) {
	if id == "" {
		return ShardID{}, ErrInvalidShardID
	}
	return ShardID{value: id}, nil
}

// GenerateShardID creates a new random ShardID using UUID v4.
func GenerateShardID() ShardID {
	return ShardID{value: uuid.New().String()}
}

// String returns the string representation of the ShardID.
func (s ShardID) String() string {
	return s.value
}

// IsZero checks if the ShardID is the zero value.
func (s ShardID) IsZero() bool {
	return s.value == ""
}

// Equals checks if two ShardIDs are equal.
func (s ShardID) Equals(other ShardID) bool {
	return s.value == other.value
}

// RedundancyLevel represents the level of Reed-Solomon redundancy.
type RedundancyLevel struct {
	percentage float64
}

// Predefined redundancy levels
var (
	RedundancyLow          = RedundancyLevel{percentage: 0.17} // 17% - Aggressive
	RedundancyBalanced     = RedundancyLevel{percentage: 0.33} // 33% - Balanced
	RedundancyConservative = RedundancyLevel{percentage: 0.50} // 50% - Conservative
)

// NewRedundancyLevel creates a new RedundancyLevel with validation.
func NewRedundancyLevel(percentage float64) (RedundancyLevel, error) {
	if percentage <= 0 || percentage > 1.0 {
		return RedundancyLevel{}, ErrInvalidRedundancy
	}
	return RedundancyLevel{percentage: percentage}, nil
}

// Percentage returns the redundancy percentage (0.0 to 1.0).
func (r RedundancyLevel) Percentage() float64 {
	return r.percentage
}

// ParityShardCount calculates the number of parity shards needed.
func (r RedundancyLevel) ParityShardCount(dataShards int) int {
	return int(float64(dataShards) * r.percentage)
}

// String returns a human-readable representation.
func (r RedundancyLevel) String() string {
	return fmt.Sprintf("%.0f%%", r.percentage*100)
}

// CalculateRedundancy calculates the redundancy level from shard counts.
func CalculateRedundancy(dataShards, parityShards int) RedundancyLevel {
	if dataShards == 0 {
		return RedundancyLevel{percentage: 0}
	}
	percentage := float64(parityShards) / float64(dataShards)
	return RedundancyLevel{percentage: percentage}
}

// ShardConfiguration represents the configuration for Reed-Solomon encoding.
type ShardConfiguration struct {
	DataShards   int
	ParityShards int
	Redundancy   RedundancyLevel
}

// NewShardConfiguration creates a new ShardConfiguration with validation.
func NewShardConfiguration(dataShards, parityShards int) (*ShardConfiguration, error) {
	if dataShards <= 0 || parityShards <= 0 {
		return nil, ErrInvalidShardCount
	}

	if dataShards+parityShards > MaxTotalShards {
		return nil, ErrTooManyShards
	}

	redundancy := CalculateRedundancy(dataShards, parityShards)

	return &ShardConfiguration{
		DataShards:   dataShards,
		ParityShards: parityShards,
		Redundancy:   redundancy,
	}, nil
}

// TotalShards returns the total number of shards.
func (c *ShardConfiguration) TotalShards() int {
	return c.DataShards + c.ParityShards
}

// MinimumRequired returns the minimum shards needed for reconstruction.
func (c *ShardConfiguration) MinimumRequired() int {
	return c.DataShards
}

// MaximumFailures returns how many shards can be lost without data loss.
func (c *ShardConfiguration) MaximumFailures() int {
	return c.ParityShards
}
