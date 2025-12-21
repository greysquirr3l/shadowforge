// Package stego provides comprehensive tests for stego domain entities.
package stego

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewStegoContainer_Success tests successful container creation.
func TestNewStegoContainer_Success(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media data")
	technique := LSB

	// Act
	container, err := NewStegoContainer(id, coverMedia, technique)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, container)
	assert.Equal(t, id, container.ID)
	assert.Equal(t, coverMedia, container.CoverMedia)
	assert.Equal(t, technique, container.Technique)
	assert.False(t, container.IsEmbedded)
	assert.WithinDuration(t, time.Now(), container.CreatedAt, time.Second)
}

// TestNewStegoContainer_InvalidID tests container creation with invalid ID.
func TestNewStegoContainer_InvalidID(t *testing.T) {
	// Arrange
	id := ContainerID{} // Zero ID
	coverMedia := []byte("test cover media")
	technique := LSB

	// Act
	container, err := NewStegoContainer(id, coverMedia, technique)

	// Assert
	assert.Nil(t, container)
	assert.ErrorIs(t, err, ErrInvalidContainerID)
}

// TestNewStegoContainer_EmptyCoverMedia tests container creation with empty media.
func TestNewStegoContainer_EmptyCoverMedia(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte{} // Empty
	technique := DCT

	// Act
	container, err := NewStegoContainer(id, coverMedia, technique)

	// Assert
	assert.Nil(t, container)
	assert.ErrorIs(t, err, ErrEmptyCoverMedia)
}

// TestNewStegoContainer_InvalidTechnique tests container creation with invalid technique.
func TestNewStegoContainer_InvalidTechnique(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	technique := StegoTechnique("invalid")

	// Act
	container, err := NewStegoContainer(id, coverMedia, technique)

	// Assert
	assert.Nil(t, container)
	assert.ErrorIs(t, err, ErrInvalidTechnique)
}

// TestStegoContainer_Embed_Success tests successful data embedding.
func TestStegoContainer_Embed_Success(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media with sufficient space")
	technique := LSB
	container, _ := NewStegoContainer(id, coverMedia, technique)
	require.NoError(t, container.SetCapacity(100))

	data := []byte("secret payload")
	embeddingMap := []int{1, 5, 10, 15, 20}

	// Act
	err := container.Embed(data, embeddingMap)

	// Assert
	require.NoError(t, err)
	assert.True(t, container.IsEmbedded)
	assert.Equal(t, data, container.EmbeddedData)
	assert.Equal(t, embeddingMap, container.EmbeddingMap)
	assert.Equal(t, int64(len(data)), container.UsedCapacity)
	assert.WithinDuration(t, time.Now(), container.ModifiedAt, time.Second)
}

// TestStegoContainer_Embed_EmptyPayload tests embedding with empty data.
func TestStegoContainer_Embed_EmptyPayload(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, PhaseEncoding)
	require.NoError(t, container.SetCapacity(100))

	data := []byte{} // Empty
	embeddingMap := []int{}

	// Act
	err := container.Embed(data, embeddingMap)

	// Assert
	assert.ErrorIs(t, err, ErrEmptyPayload)
	assert.False(t, container.IsEmbedded)
}

// TestStegoContainer_Embed_AlreadyEmbedded tests embedding when already embedded.
func TestStegoContainer_Embed_AlreadyEmbedded(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, EchoHiding)
	require.NoError(t, container.SetCapacity(100))
	require.NoError(t, container.Embed([]byte("first payload"), []int{1, 2, 3}))

	// Act
	err := container.Embed([]byte("second payload"), []int{4, 5, 6})

	// Assert
	assert.ErrorIs(t, err, ErrAlreadyEmbedded)
	assert.Equal(t, []byte("first payload"), container.EmbeddedData) // Unchanged
}

// TestStegoContainer_Embed_CapacityExceeded tests embedding exceeding capacity.
func TestStegoContainer_Embed_CapacityExceeded(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("small cover")
	container, _ := NewStegoContainer(id, coverMedia, ZeroWidth)
	require.NoError(t, container.SetCapacity(5)) // Small capacity

	data := []byte("this is a very large payload exceeding capacity")

	// Act
	err := container.Embed(data, []int{})

	// Assert
	assert.ErrorIs(t, err, ErrCapacityExceeded)
	assert.False(t, container.IsEmbedded)
}

// TestStegoContainer_Extract_Success tests successful data extraction.
func TestStegoContainer_Extract_Success(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, Palette)
	require.NoError(t, container.SetCapacity(100))
	expectedData := []byte("hidden message")
	require.NoError(t, container.Embed(expectedData, []int{1, 2, 3}))

	// Act
	extracted, err := container.Extract()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedData, extracted)
}

// TestStegoContainer_Extract_NoEmbeddedData tests extraction without embedding.
func TestStegoContainer_Extract_NoEmbeddedData(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, Hybrid)

	// Act
	extracted, err := container.Extract()

	// Assert
	assert.Nil(t, extracted)
	assert.ErrorIs(t, err, ErrNoEmbeddedData)
}

// TestStegoContainer_SetCapacity_Success tests setting valid capacity.
func TestStegoContainer_SetCapacity_Success(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, DCT)
	capacity := int64(1024)

	// Act
	err := container.SetCapacity(capacity)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, capacity, container.Capacity)
}

// TestStegoContainer_SetCapacity_Invalid tests setting invalid capacity.
func TestStegoContainer_SetCapacity_Invalid(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, LSB)

	tests := []struct {
		name     string
		capacity int64
	}{
		{"zero capacity", 0},
		{"negative capacity", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := container.SetCapacity(tt.capacity)

			// Assert
			assert.ErrorIs(t, err, ErrInvalidCapacity)
		})
	}
}

// TestStegoContainer_RemainingCapacity tests capacity calculation.
func TestStegoContainer_RemainingCapacity(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, PhaseEncoding)
	require.NoError(t, container.SetCapacity(100))
	require.NoError(t, container.Embed([]byte("secret"), []int{1, 2, 3})) // 6 bytes used

	// Act
	remaining := container.RemainingCapacity()

	// Assert
	assert.Equal(t, int64(94), remaining) // 100 - 6
}

// TestStegoContainer_CanEmbed_Success tests capacity check with available space.
func TestStegoContainer_CanEmbed_Success(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, EchoHiding)
	require.NoError(t, container.SetCapacity(100))

	// Act
	canEmbed := container.CanEmbed(50)

	// Assert
	assert.True(t, canEmbed)
}

// TestStegoContainer_CanEmbed_ExceedsCapacity tests capacity check exceeding limit.
func TestStegoContainer_CanEmbed_ExceedsCapacity(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, ZeroWidth)
	require.NoError(t, container.SetCapacity(100))

	// Act
	canEmbed := container.CanEmbed(150)

	// Assert
	assert.False(t, canEmbed)
}

// TestStegoContainer_CanEmbed_AlreadyEmbedded tests capacity check when embedded.
func TestStegoContainer_CanEmbed_AlreadyEmbedded(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, Palette)
	require.NoError(t, container.SetCapacity(100))
	require.NoError(t, container.Embed([]byte("data"), []int{1, 2}))

	// Act
	canEmbed := container.CanEmbed(10)

	// Assert
	assert.False(t, canEmbed) // False because already embedded
}

// TestStegoContainer_SetQuality_Success tests setting valid quality.
func TestStegoContainer_SetQuality_Success(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, Hybrid)
	quality, _ := NewQuality(0.85)

	// Act
	err := container.SetQuality(quality)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, quality, container.Quality)
}

// TestStegoContainer_SetQuality_Invalid tests setting invalid quality.
func TestStegoContainer_SetQuality_Invalid(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, DCT)
	invalidQuality := Quality{Score: 1.5} // Invalid score > 1.0

	// Act
	err := container.SetQuality(invalidQuality)

	// Assert
	assert.ErrorIs(t, err, ErrInvalidQuality)
}

// TestStegoContainer_Validate_Success tests validation of valid container.
func TestStegoContainer_Validate_Success(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, LSB)
	require.NoError(t, container.SetCapacity(100))

	// Act
	err := container.Validate()

	// Assert
	assert.NoError(t, err)
}

// TestStegoContainer_Validate_InvalidContainerID tests validation with invalid ID.
func TestStegoContainer_Validate_InvalidContainerID(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, PhaseEncoding)
	container.ID = ContainerID{} // Zero ID

	// Act
	err := container.Validate()

	// Assert
	assert.ErrorIs(t, err, ErrInvalidContainerID)
}

// TestStegoContainer_Validate_EmptyCoverMedia tests validation with empty media.
func TestStegoContainer_Validate_EmptyCoverMedia(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, EchoHiding)
	container.CoverMedia = []byte{} // Empty

	// Act
	err := container.Validate()

	// Assert
	assert.ErrorIs(t, err, ErrEmptyCoverMedia)
}

// TestStegoContainer_Validate_InvalidTechnique tests validation with invalid technique.
func TestStegoContainer_Validate_InvalidTechnique(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, ZeroWidth)
	container.Technique = StegoTechnique("invalid")

	// Act
	err := container.Validate()

	// Assert
	assert.ErrorIs(t, err, ErrInvalidTechnique)
}

// TestStegoContainer_Validate_NegativeCapacity tests validation with negative capacity.
func TestStegoContainer_Validate_NegativeCapacity(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, Palette)
	container.Capacity = -100 // Invalid

	// Act
	err := container.Validate()

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCapacity)
}

// TestStegoContainer_Validate_EmbeddedButNoData tests validation inconsistency.
func TestStegoContainer_Validate_EmbeddedButNoData(t *testing.T) {
	// Arrange
	id := GenerateContainerID()
	coverMedia := []byte("test cover media")
	container, _ := NewStegoContainer(id, coverMedia, Hybrid)
	container.IsEmbedded = true
	container.EmbeddedData = []byte{} // Inconsistent state

	// Act
	err := container.Validate()

	// Assert
	assert.ErrorIs(t, err, ErrNoEmbeddedData)
}

// TestNewEmbeddingMetadata_Success tests successful metadata creation.
func TestNewEmbeddingMetadata_Success(t *testing.T) {
	// Arrange
	containerID := GenerateContainerID()
	technique := DCT
	embeddingMap := []int{10, 20, 30, 40}
	payloadSize := int64(256)

	// Act
	metadata, err := NewEmbeddingMetadata(containerID, technique, embeddingMap, payloadSize)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, containerID, metadata.ContainerID)
	assert.Equal(t, technique, metadata.Technique)
	assert.Equal(t, embeddingMap, metadata.EmbeddingMap)
	assert.Equal(t, payloadSize, metadata.PayloadSize)
	assert.WithinDuration(t, time.Now(), metadata.EmbeddedAt, time.Second)
}

// TestNewEmbeddingMetadata_InvalidContainerID tests metadata with invalid ID.
func TestNewEmbeddingMetadata_InvalidContainerID(t *testing.T) {
	// Arrange
	containerID := ContainerID{} // Zero ID
	technique := LSB
	embeddingMap := []int{1, 2, 3}
	payloadSize := int64(128)

	// Act
	metadata, err := NewEmbeddingMetadata(containerID, technique, embeddingMap, payloadSize)

	// Assert
	assert.Nil(t, metadata)
	assert.ErrorIs(t, err, ErrInvalidContainerID)
}

// TestNewEmbeddingMetadata_InvalidTechnique tests metadata with invalid technique.
func TestNewEmbeddingMetadata_InvalidTechnique(t *testing.T) {
	// Arrange
	containerID := GenerateContainerID()
	technique := StegoTechnique("unknown")
	embeddingMap := []int{5, 10, 15}
	payloadSize := int64(64)

	// Act
	metadata, err := NewEmbeddingMetadata(containerID, technique, embeddingMap, payloadSize)

	// Assert
	assert.Nil(t, metadata)
	assert.ErrorIs(t, err, ErrInvalidTechnique)
}

// TestNewEmbeddingMetadata_InvalidPayloadSize tests metadata with invalid size.
func TestNewEmbeddingMetadata_InvalidPayloadSize(t *testing.T) {
	// Arrange
	containerID := GenerateContainerID()
	technique := PhaseEncoding
	embeddingMap := []int{1, 2, 3}

	tests := []struct {
		name        string
		payloadSize int64
	}{
		{"zero size", 0},
		{"negative size", -50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			metadata, err := NewEmbeddingMetadata(containerID, technique, embeddingMap, tt.payloadSize)

			// Assert
			assert.Nil(t, metadata)
			assert.ErrorIs(t, err, ErrInvalidPayloadSize)
		})
	}
}

// TestEmbeddingMetadata_SetEntropyScore_Success tests setting valid entropy score.
func TestEmbeddingMetadata_SetEntropyScore_Success(t *testing.T) {
	// Arrange
	containerID := GenerateContainerID()
	metadata, _ := NewEmbeddingMetadata(containerID, EchoHiding, []int{1}, 100)
	score := 0.75

	// Act
	err := metadata.SetEntropyScore(score)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, score, metadata.EntropyScore)
}

// TestEmbeddingMetadata_SetEntropyScore_Invalid tests setting invalid entropy score.
func TestEmbeddingMetadata_SetEntropyScore_Invalid(t *testing.T) {
	// Arrange
	containerID := GenerateContainerID()
	metadata, _ := NewEmbeddingMetadata(containerID, ZeroWidth, []int{1}, 50)

	tests := []struct {
		name  string
		score float64
	}{
		{"below range", -0.5},
		{"above range", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := metadata.SetEntropyScore(tt.score)

			// Assert
			assert.ErrorIs(t, err, ErrInvalidScore)
		})
	}
}

// TestEmbeddingMetadata_SetQualityScore_Success tests setting valid quality score.
func TestEmbeddingMetadata_SetQualityScore_Success(t *testing.T) {
	// Arrange
	containerID := GenerateContainerID()
	metadata, _ := NewEmbeddingMetadata(containerID, Palette, []int{1, 2}, 200)
	score := 0.90

	// Act
	err := metadata.SetQualityScore(score)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, score, metadata.QualityScore)
}

// TestEmbeddingMetadata_SetQualityScore_Invalid tests setting invalid quality score.
func TestEmbeddingMetadata_SetQualityScore_Invalid(t *testing.T) {
	// Arrange
	containerID := GenerateContainerID()
	metadata, _ := NewEmbeddingMetadata(containerID, Hybrid, []int{5}, 150)

	tests := []struct {
		name  string
		score float64
	}{
		{"negative score", -0.2},
		{"excessive score", 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := metadata.SetQualityScore(tt.score)

			// Assert
			assert.ErrorIs(t, err, ErrInvalidScore)
		})
	}
}
