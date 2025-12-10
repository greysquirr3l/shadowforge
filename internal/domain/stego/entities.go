// Package stego implements the Steganography bounded context.
// It handles embedding and extracting data from cover media.
package stego

import (
	"time"
)

// StegoContainer is the root aggregate for steganographic operations.
// It represents cover media that contains or will contain hidden data.
type StegoContainer struct {
	ID           ContainerID
	CoverMedia   []byte
	EmbeddedData []byte
	Technique    StegoTechnique
	Capacity     int64
	UsedCapacity int64
	Quality      Quality
	EmbeddingMap []int
	CreatedAt    time.Time
	ModifiedAt   time.Time
	IsEmbedded   bool
}

// NewStegoContainer creates a new StegoContainer with validation.
func NewStegoContainer(id ContainerID, coverMedia []byte, technique StegoTechnique) (*StegoContainer, error) {
	if id.IsZero() {
		return nil, ErrInvalidContainerID
	}

	if len(coverMedia) == 0 {
		return nil, ErrEmptyCoverMedia
	}

	if !technique.IsValid() {
		return nil, ErrInvalidTechnique
	}

	return &StegoContainer{
		ID:         id,
		CoverMedia: coverMedia,
		Technique:  technique,
		CreatedAt:  time.Now(),
		IsEmbedded: false,
	}, nil
}

// Embed marks the container as having embedded data.
func (s *StegoContainer) Embed(data []byte, embeddingMap []int) error {
	if len(data) == 0 {
		return ErrEmptyPayload
	}

	if s.IsEmbedded {
		return ErrAlreadyEmbedded
	}

	if int64(len(data)) > s.Capacity {
		return ErrCapacityExceeded
	}

	s.EmbeddedData = data
	s.EmbeddingMap = embeddingMap
	s.UsedCapacity = int64(len(data))
	s.ModifiedAt = time.Now()
	s.IsEmbedded = true

	return nil
}

// Extract retrieves the embedded data from the container.
func (s *StegoContainer) Extract() ([]byte, error) {
	if !s.IsEmbedded {
		return nil, ErrNoEmbeddedData
	}

	if len(s.EmbeddedData) == 0 {
		return nil, ErrEmptyPayload
	}

	return s.EmbeddedData, nil
}

// SetCapacity sets the maximum embedding capacity for this container.
func (s *StegoContainer) SetCapacity(capacity int64) error {
	if capacity <= 0 {
		return ErrInvalidCapacity
	}

	s.Capacity = capacity
	return nil
}

// RemainingCapacity returns the available capacity for embedding.
func (s *StegoContainer) RemainingCapacity() int64 {
	return s.Capacity - s.UsedCapacity
}

// CanEmbed checks if data of given size can be embedded.
func (s *StegoContainer) CanEmbed(size int64) bool {
	return !s.IsEmbedded && size <= s.Capacity
}

// SetQuality sets the quality score for this container.
func (s *StegoContainer) SetQuality(quality Quality) error {
	if !quality.IsValid() {
		return ErrInvalidQuality
	}

	s.Quality = quality
	return nil
}

// Validate checks the integrity of the StegoContainer.
func (s *StegoContainer) Validate() error {
	if s.ID.IsZero() {
		return ErrInvalidContainerID
	}

	if len(s.CoverMedia) == 0 {
		return ErrEmptyCoverMedia
	}

	if !s.Technique.IsValid() {
		return ErrInvalidTechnique
	}

	if s.Capacity < 0 {
		return ErrInvalidCapacity
	}

	if s.IsEmbedded && len(s.EmbeddedData) == 0 {
		return ErrNoEmbeddedData
	}

	return nil
}

// EmbeddingMetadata contains metadata about the embedding process.
type EmbeddingMetadata struct {
	ContainerID  ContainerID
	Technique    StegoTechnique
	EmbeddedAt   time.Time
	EmbeddingMap []int
	PayloadSize  int64
	EntropyScore float64
	QualityScore float64
}

// NewEmbeddingMetadata creates embedding metadata with validation.
func NewEmbeddingMetadata(containerID ContainerID, technique StegoTechnique, embeddingMap []int, payloadSize int64) (*EmbeddingMetadata, error) {
	if containerID.IsZero() {
		return nil, ErrInvalidContainerID
	}

	if !technique.IsValid() {
		return nil, ErrInvalidTechnique
	}

	if payloadSize <= 0 {
		return nil, ErrInvalidPayloadSize
	}

	return &EmbeddingMetadata{
		ContainerID:  containerID,
		Technique:    technique,
		EmbeddedAt:   time.Now(),
		EmbeddingMap: embeddingMap,
		PayloadSize:  payloadSize,
	}, nil
}

// SetEntropyScore sets the entropy score for the embedding.
func (e *EmbeddingMetadata) SetEntropyScore(score float64) error {
	if score < 0.0 || score > 1.0 {
		return ErrInvalidScore
	}

	e.EntropyScore = score
	return nil
}

// SetQualityScore sets the quality score for the embedding.
func (e *EmbeddingMetadata) SetQualityScore(score float64) error {
	if score < 0.0 || score > 1.0 {
		return ErrInvalidScore
	}

	e.QualityScore = score
	return nil
}
