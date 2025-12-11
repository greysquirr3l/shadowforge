// Package errorcorrection implements the Error Correction bounded context.
// It handles Reed-Solomon encoding/decoding for data resilience.
package errorcorrection

import (
	"time"
)

// ProtectedMessage is the root aggregate for error-corrected data.
// It represents data that has been encoded with Reed-Solomon for fault tolerance.
type ProtectedMessage struct {
	ID           MessageID
	OriginalData []byte
	DataShards   int
	ParityShards int
	Shards       []*Shard
	EncodedAt    time.Time
	Redundancy   RedundancyLevel
}

// NewProtectedMessage creates a new ProtectedMessage with validation.
func NewProtectedMessage(id MessageID, data []byte, dataShards, parityShards int) (*ProtectedMessage, error) {
	if len(data) == 0 {
		return nil, ErrEmptyData
	}

	if dataShards <= 0 || parityShards <= 0 {
		return nil, ErrInvalidShardCount
	}

	redundancy := CalculateRedundancy(dataShards, parityShards)

	return &ProtectedMessage{
		ID:           id,
		OriginalData: data,
		DataShards:   dataShards,
		ParityShards: parityShards,
		EncodedAt:    time.Now(),
		Redundancy:   redundancy,
	}, nil
}

// AddShard adds an encoded shard to the message.
func (m *ProtectedMessage) AddShard(shard *Shard) error {
	if shard == nil {
		return ErrNilShard
	}

	if !shard.MessageID.Equals(m.ID) {
		return ErrShardMessageMismatch
	}

	m.Shards = append(m.Shards, shard)
	return nil
}

// TotalShards returns the total number of shards (data + parity).
func (m *ProtectedMessage) TotalShards() int {
	return m.DataShards + m.ParityShards
}

// MinimumShardsNeeded returns the minimum shards required for recovery.
func (m *ProtectedMessage) MinimumShardsNeeded() int {
	return m.DataShards
}

// CanRecover checks if we have enough shards to reconstruct the data.
func (m *ProtectedMessage) CanRecover() bool {
	return len(m.Shards) >= m.MinimumShardsNeeded()
}

// Validate checks the integrity of the ProtectedMessage.
func (m *ProtectedMessage) Validate() error {
	if m.ID.IsZero() {
		return ErrInvalidMessageID
	}

	if m.DataShards <= 0 || m.ParityShards <= 0 {
		return ErrInvalidShardCount
	}

	if m.DataShards+m.ParityShards > MaxTotalShards {
		return ErrTooManyShards
	}

	return nil
}

// Shard represents a single piece of encoded data.
type Shard struct {
	ID        ShardID
	MessageID MessageID
	Index     int
	Data      []byte
	IsParity  bool
	Checksum  []byte
	CreatedAt time.Time
}

// NewShard creates a new Shard with validation.
func NewShard(id ShardID, messageID MessageID, index int, data []byte, isParity bool) (*Shard, error) {
	if id.IsZero() {
		return nil, ErrInvalidShardID
	}

	if messageID.IsZero() {
		return nil, ErrInvalidMessageID
	}

	if index < 0 {
		return nil, ErrInvalidShardIndex
	}

	if len(data) == 0 {
		return nil, ErrEmptyShardData
	}

	return &Shard{
		ID:        id,
		MessageID: messageID,
		Index:     index,
		Data:      data,
		IsParity:  isParity,
		CreatedAt: time.Now(),
	}, nil
}

// SetChecksum sets the checksum for integrity verification.
func (s *Shard) SetChecksum(checksum []byte) {
	s.Checksum = checksum
}

// VerifyChecksum checks if the shard's checksum matches the calculated value.
func (s *Shard) VerifyChecksum(calculated []byte) bool {
	if len(s.Checksum) == 0 || len(calculated) == 0 {
		return false
	}

	// SECURITY: Use constant-time comparison for checksums
	if len(s.Checksum) != len(calculated) {
		return false
	}

	result := byte(0)
	for i := range s.Checksum {
		result |= s.Checksum[i] ^ calculated[i]
	}

	return result == 0
}

// Validate checks the integrity of the Shard.
func (s *Shard) Validate() error {
	if s.ID.IsZero() {
		return ErrInvalidShardID
	}

	if s.MessageID.IsZero() {
		return ErrInvalidMessageID
	}

	if s.Index < 0 {
		return ErrInvalidShardIndex
	}

	if len(s.Data) == 0 {
		return ErrEmptyShardData
	}

	return nil
}
