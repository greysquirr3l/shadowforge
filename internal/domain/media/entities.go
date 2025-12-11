// Package media implements the Media Processing bounded context.
// It handles media file loading, format detection, and capacity calculation.
package media

import (
	"time"
)

// MediaAsset is the root aggregate for media processing operations.
// It represents a media file (Image, Audio, or Text) that can be used as a cover.
type MediaAsset struct {
	ID              AssetID
	Type            MediaType
	Format          MediaFormat
	Data            []byte
	Dimensions      *Dimensions // For images
	Resolution      *Resolution // For images
	SampleRate      *SampleRate // For audio
	ColorSpace      *ColorSpace // For images
	Metadata        map[string]string
	Capacity        int64
	LoadedAt        time.Time
	ModifiedAt      time.Time
	IsSanitized     bool
	FormatValidated bool
}

// NewMediaAsset creates a new MediaAsset with validation.
func NewMediaAsset(id AssetID, mediaType MediaType, format MediaFormat, data []byte) (*MediaAsset, error) {
	if id.IsZero() {
		return nil, ErrInvalidAssetID
	}

	if !mediaType.IsValid() {
		return nil, ErrInvalidMediaType
	}

	if !format.IsValid() {
		return nil, ErrInvalidFormat
	}

	if len(data) == 0 {
		return nil, ErrEmptyMediaData
	}

	return &MediaAsset{
		ID:              id,
		Type:            mediaType,
		Format:          format,
		Data:            data,
		Metadata:        make(map[string]string),
		LoadedAt:        time.Now(),
		IsSanitized:     false,
		FormatValidated: false,
	}, nil
}

// SetDimensions sets the dimensions for image media.
func (m *MediaAsset) SetDimensions(dimensions *Dimensions) error {
	if m.Type != MediaTypeImage {
		return ErrInvalidMediaType
	}

	if dimensions == nil {
		return ErrInvalidDimensions
	}

	if !dimensions.IsValid() {
		return ErrInvalidDimensions
	}

	m.Dimensions = dimensions
	m.ModifiedAt = time.Now()
	return nil
}

// SetResolution sets the resolution for image media.
func (m *MediaAsset) SetResolution(resolution *Resolution) error {
	if m.Type != MediaTypeImage {
		return ErrInvalidMediaType
	}

	if resolution == nil {
		return ErrInvalidResolution
	}

	if !resolution.IsValid() {
		return ErrInvalidResolution
	}

	m.Resolution = resolution
	m.ModifiedAt = time.Now()
	return nil
}

// SetSampleRate sets the sample rate for audio media.
func (m *MediaAsset) SetSampleRate(sampleRate *SampleRate) error {
	if m.Type != MediaTypeAudio {
		return ErrInvalidMediaType
	}

	if sampleRate == nil {
		return ErrInvalidSampleRate
	}

	if !sampleRate.IsValid() {
		return ErrInvalidSampleRate
	}

	m.SampleRate = sampleRate
	m.ModifiedAt = time.Now()
	return nil
}

// SetColorSpace sets the color space for image media.
func (m *MediaAsset) SetColorSpace(colorSpace *ColorSpace) error {
	if m.Type != MediaTypeImage {
		return ErrInvalidMediaType
	}

	if colorSpace == nil {
		return ErrInvalidColorSpace
	}

	if !colorSpace.IsValid() {
		return ErrInvalidColorSpace
	}

	m.ColorSpace = colorSpace
	m.ModifiedAt = time.Now()
	return nil
}

// SetCapacity sets the embedding capacity for this media.
func (m *MediaAsset) SetCapacity(capacity int64) error {
	if capacity < 0 {
		return ErrInvalidCapacity
	}

	m.Capacity = capacity
	m.ModifiedAt = time.Now()
	return nil
}

// Sanitize marks the media as having metadata sanitized.
func (m *MediaAsset) Sanitize() {
	m.Metadata = make(map[string]string) // Clear all metadata
	m.IsSanitized = true
	m.ModifiedAt = time.Now()
}

// ValidateFormat marks the format as validated.
func (m *MediaAsset) ValidateFormat() {
	m.FormatValidated = true
	m.ModifiedAt = time.Now()
}

// Validate checks the integrity of the MediaAsset.
func (m *MediaAsset) Validate() error {
	if m.ID.IsZero() {
		return ErrInvalidAssetID
	}

	if !m.Type.IsValid() {
		return ErrInvalidMediaType
	}

	if !m.Format.IsValid() {
		return ErrInvalidFormat
	}

	if len(m.Data) == 0 {
		return ErrEmptyMediaData
	}

	// Type-specific validation
	switch m.Type {
	case MediaTypeImage:
		if m.Dimensions == nil {
			return ErrInvalidDimensions
		}
		if !m.Dimensions.IsValid() {
			return ErrInvalidDimensions
		}
	case MediaTypeAudio:
		if m.SampleRate == nil {
			return ErrInvalidSampleRate
		}
		if !m.SampleRate.IsValid() {
			return ErrInvalidSampleRate
		}
	case MediaTypeText:
	// Text media doesn't require additional properties
	default:
		return ErrUnsupportedMediaType
	}

	return nil
}

// CapacityInfo contains information about media embedding capacity.
type CapacityInfo struct {
	AssetID        AssetID
	MediaType      MediaType
	TotalCapacity  int64
	UsableCapacity int64
	RecommendedMax int64
	CalculatedAt   time.Time
	QualityImpact  float64 // 0.0 to 1.0
}

// NewCapacityInfo creates capacity information with validation.
func NewCapacityInfo(assetID AssetID, mediaType MediaType, totalCapacity, usableCapacity, recommendedMax int64) (*CapacityInfo, error) {
	if assetID.IsZero() {
		return nil, ErrInvalidAssetID
	}

	if !mediaType.IsValid() {
		return nil, ErrInvalidMediaType
	}

	if totalCapacity < 0 || usableCapacity < 0 || recommendedMax < 0 {
		return nil, ErrInvalidCapacity
	}

	if usableCapacity > totalCapacity {
		return nil, ErrInvalidCapacity
	}

	if recommendedMax > usableCapacity {
		return nil, ErrInvalidCapacity
	}

	return &CapacityInfo{
		AssetID:        assetID,
		MediaType:      mediaType,
		TotalCapacity:  totalCapacity,
		UsableCapacity: usableCapacity,
		RecommendedMax: recommendedMax,
		CalculatedAt:   time.Now(),
	}, nil
}

// SetQualityImpact sets the quality impact score.
func (c *CapacityInfo) SetQualityImpact(impact float64) error {
	if impact < 0.0 || impact > 1.0 {
		return ErrInvalidQualityScore
	}

	c.QualityImpact = impact
	return nil
}
