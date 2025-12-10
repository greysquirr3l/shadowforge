// Package stego implements the Steganography bounded context.
package stego

import (
	"fmt"

	"github.com/google/uuid"
)

// ContainerID is a unique identifier for a steganographic container.
type ContainerID struct {
	value string
}

// NewContainerID creates a new ContainerID from a string.
func NewContainerID(id string) (ContainerID, error) {
	if id == "" {
		return ContainerID{}, ErrInvalidContainerID
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return ContainerID{}, ErrInvalidContainerID
	}

	return ContainerID{value: id}, nil
}

// GenerateContainerID creates a new random ContainerID.
func GenerateContainerID() ContainerID {
	return ContainerID{value: uuid.New().String()}
}

// String returns the string representation of the ContainerID.
func (c ContainerID) String() string {
	return c.value
}

// Equals checks if two ContainerIDs are equal.
func (c ContainerID) Equals(other ContainerID) bool {
	return c.value == other.value
}

// IsZero checks if the ContainerID is the zero value.
func (c ContainerID) IsZero() bool {
	return c.value == ""
}

// StegoTechnique represents the steganography technique used.
type StegoTechnique string

const (
	// LSB (Least Significant Bit) - Image steganography
	LSB StegoTechnique = "lsb"

	// DCT (Discrete Cosine Transform) - JPEG steganography
	DCT StegoTechnique = "dct"

	// PhaseEncoding - Audio steganography using phase
	PhaseEncoding StegoTechnique = "phase_encoding"

	// EchoHiding - Audio steganography using echo
	EchoHiding StegoTechnique = "echo_hiding"

	// ZeroWidth - Text steganography using zero-width characters
	ZeroWidth StegoTechnique = "zero_width"

	// Palette - Image steganography using palette manipulation
	Palette StegoTechnique = "palette"

	// Hybrid - Combination of multiple techniques
	Hybrid StegoTechnique = "hybrid"
)

// IsValid checks if the technique is valid.
func (t StegoTechnique) IsValid() bool {
	switch t {
	case LSB, DCT, PhaseEncoding, EchoHiding, ZeroWidth, Palette, Hybrid:
		return true
	default:
		return false
	}
}

// String returns the string representation of the technique.
func (t StegoTechnique) String() string {
	return string(t)
}

// Name returns a human-readable name for the technique.
func (t StegoTechnique) Name() string {
	switch t {
	case LSB:
		return "Least Significant Bit"
	case DCT:
		return "Discrete Cosine Transform"
	case PhaseEncoding:
		return "Phase Encoding"
	case EchoHiding:
		return "Echo Hiding"
	case ZeroWidth:
		return "Zero-Width Characters"
	case Palette:
		return "Palette Manipulation"
	case Hybrid:
		return "Hybrid Technique"
	default:
		return "Unknown"
	}
}

// MediaType returns the primary media type for this technique.
func (t StegoTechnique) MediaType() string {
	switch t {
	case LSB, DCT, Palette:
		return "image"
	case PhaseEncoding, EchoHiding:
		return "audio"
	case ZeroWidth:
		return "text"
	case Hybrid:
		return "mixed"
	default:
		return "unknown"
	}
}

// Capacity represents the embedding capacity in bytes.
type Capacity struct {
	Total     int64
	Available int64
	Used      int64
}

// NewCapacity creates a new Capacity value object.
func NewCapacity(total int64) (Capacity, error) {
	if total <= 0 {
		return Capacity{}, ErrInvalidCapacity
	}

	return Capacity{
		Total:     total,
		Available: total,
		Used:      0,
	}, nil
}

// Reserve reserves capacity for embedding.
func (c *Capacity) Reserve(size int64) error {
	if size <= 0 {
		return ErrInvalidPayloadSize
	}

	if size > c.Available {
		return ErrCapacityExceeded
	}

	c.Used += size
	c.Available -= size
	return nil
}

// Release releases previously reserved capacity.
func (c *Capacity) Release(size int64) error {
	if size <= 0 {
		return ErrInvalidPayloadSize
	}

	if size > c.Used {
		return fmt.Errorf("cannot release more than used capacity")
	}

	c.Used -= size
	c.Available += size
	return nil
}

// UtilizationPercentage returns the percentage of capacity used.
func (c Capacity) UtilizationPercentage() float64 {
	if c.Total == 0 {
		return 0.0
	}
	return float64(c.Used) / float64(c.Total) * 100.0
}

// CanFit checks if a given size can fit in available capacity.
func (c Capacity) CanFit(size int64) bool {
	return size <= c.Available
}

// Quality represents the quality score of steganographic embedding.
type Quality struct {
	Score         float64 // 0.0 to 1.0
	Detectability float64 // 0.0 (undetectable) to 1.0 (obvious)
	FidelityScore float64 // 0.0 (poor) to 1.0 (perfect)
}

// NewQuality creates a new Quality value object with validation.
func NewQuality(score float64) (Quality, error) {
	if score < 0.0 || score > 1.0 {
		return Quality{}, ErrInvalidQuality
	}

	return Quality{
		Score: score,
	}, nil
}

// SetDetectability sets the detectability score.
func (q *Quality) SetDetectability(score float64) error {
	if score < 0.0 || score > 1.0 {
		return ErrInvalidScore
	}

	q.Detectability = score
	return nil
}

// SetFidelityScore sets the fidelity score.
func (q *Quality) SetFidelityScore(score float64) error {
	if score < 0.0 || score > 1.0 {
		return ErrInvalidScore
	}

	q.FidelityScore = score
	return nil
}

// IsValid checks if the quality metrics are valid.
func (q Quality) IsValid() bool {
	return q.Score >= 0.0 && q.Score <= 1.0 &&
		q.Detectability >= 0.0 && q.Detectability <= 1.0 &&
		q.FidelityScore >= 0.0 && q.FidelityScore <= 1.0
}

// IsAcceptable checks if quality meets minimum thresholds.
func (q Quality) IsAcceptable() bool {
	const minScore = 0.6
	const maxDetectability = 0.3
	const minFidelity = 0.7

	return q.Score >= minScore &&
		q.Detectability <= maxDetectability &&
		q.FidelityScore >= minFidelity
}
