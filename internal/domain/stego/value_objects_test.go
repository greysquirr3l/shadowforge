// Package stego provides comprehensive tests for stego domain value objects.
package stego

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ContainerID Tests
// =============================================================================

// TestGenerateContainerID tests ContainerID generation.
func TestGenerateContainerID(t *testing.T) {
	// Act
	id := GenerateContainerID()

	// Assert
	assert.NotEmpty(t, id.String())
	assert.False(t, id.IsZero())
}

// TestNewContainerID_Success tests creating ContainerID from valid UUID string.
func TestNewContainerID_Success(t *testing.T) {
	// Arrange
	original := GenerateContainerID()
	idStr := original.String()

	// Act
	parsed, err := NewContainerID(idStr)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, original, parsed)
}

// TestNewContainerID_EmptyString tests creating ContainerID from empty string.
func TestNewContainerID_EmptyString(t *testing.T) {
	// Act
	id, err := NewContainerID("")

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidContainerID)
	assert.True(t, id.IsZero())
}

// TestNewContainerID_InvalidUUID tests creating ContainerID from invalid UUID.
func TestNewContainerID_InvalidUUID(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"not a UUID", "not-a-uuid"},
		{"partial UUID", "12345678"},
		{"malformed", "xxxxx-xxxx-xxxx-xxxx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			id, err := NewContainerID(tt.input)

			// Assert
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidContainerID)
			assert.True(t, id.IsZero())
		})
	}
}

// TestContainerID_String tests string representation.
func TestContainerID_String(t *testing.T) {
	// Arrange
	id := GenerateContainerID()

	// Act
	str := id.String()

	// Assert
	assert.NotEmpty(t, str)
	assert.Len(t, str, 36) // UUID format: 8-4-4-4-12
	assert.Contains(t, str, "-")
}

// TestContainerID_Equals tests ContainerID equality.
func TestContainerID_Equals(t *testing.T) {
	// Arrange
	id1 := GenerateContainerID()
	id2 := id1
	id3 := GenerateContainerID()

	// Act & Assert
	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}

// TestContainerID_IsZero tests zero value detection.
func TestContainerID_IsZero(t *testing.T) {
	// Zero value
	zero := ContainerID{}
	assert.True(t, zero.IsZero())

	// Valid ID
	valid := GenerateContainerID()
	assert.False(t, valid.IsZero())
}

// =============================================================================
// StegoTechnique Tests
// =============================================================================

// TestStegoTechnique_IsValid tests validation of all techniques.
func TestStegoTechnique_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		technique StegoTechnique
		wantValid bool
	}{
		{"LSB", LSB, true},
		{"DCT", DCT, true},
		{"PhaseEncoding", PhaseEncoding, true},
		{"EchoHiding", EchoHiding, true},
		{"ZeroWidth", ZeroWidth, true},
		{"Palette", Palette, true},
		{"Hybrid", Hybrid, true},
		{"Invalid", StegoTechnique("unknown"), false},
		{"Empty", StegoTechnique(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantValid, tt.technique.IsValid())
		})
	}
}

// TestStegoTechnique_String tests string representation.
func TestStegoTechnique_String(t *testing.T) {
	assert.Equal(t, "lsb", LSB.String())
	assert.Equal(t, "dct", DCT.String())
	assert.Equal(t, "phase_encoding", PhaseEncoding.String())
}

// TestStegoTechnique_Name tests human-readable names.
func TestStegoTechnique_Name(t *testing.T) {
	tests := []struct {
		technique StegoTechnique
		wantName  string
	}{
		{LSB, "Least Significant Bit"},
		{DCT, "Discrete Cosine Transform"},
		{PhaseEncoding, "Phase Encoding"},
		{EchoHiding, "Echo Hiding"},
		{ZeroWidth, "Zero-Width Characters"},
		{Palette, "Palette Manipulation"},
		{Hybrid, "Hybrid Technique"},
		{StegoTechnique("unknown"), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			assert.Equal(t, tt.wantName, tt.technique.Name())
		})
	}
}

// TestStegoTechnique_MediaType tests media type detection.
func TestStegoTechnique_MediaType(t *testing.T) {
	tests := []struct {
		technique StegoTechnique
		wantType  string
	}{
		{LSB, "image"},
		{DCT, "image"},
		{Palette, "image"},
		{PhaseEncoding, "audio"},
		{EchoHiding, "audio"},
		{ZeroWidth, "text"},
		{Hybrid, "mixed"},
		{StegoTechnique("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.wantType, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.technique.MediaType())
		})
	}
}

// =============================================================================
// Capacity Tests
// =============================================================================

// TestNewCapacity_Success tests creating valid Capacity.
func TestNewCapacity_Success(t *testing.T) {
	tests := []struct {
		name  string
		total int64
	}{
		{"small capacity", 100},
		{"medium capacity", 10000},
		{"large capacity", 1000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			capacity, err := NewCapacity(tt.total)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.total, capacity.Total)
			assert.Equal(t, tt.total, capacity.Available)
			assert.Equal(t, int64(0), capacity.Used)
		})
	}
}

// TestNewCapacity_Invalid tests creating invalid Capacity.
func TestNewCapacity_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		total int64
	}{
		{"zero capacity", 0},
		{"negative capacity", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			capacity, err := NewCapacity(tt.total)

			// Assert
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidCapacity)
			assert.Equal(t, Capacity{}, capacity)
		})
	}
}

// TestCapacity_Reserve_Success tests reserving capacity.
func TestCapacity_Reserve_Success(t *testing.T) {
	// Arrange
	capacity, _ := NewCapacity(1000)

	// Act
	err := capacity.Reserve(300)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int64(300), capacity.Used)
	assert.Equal(t, int64(700), capacity.Available)
}

// TestCapacity_Reserve_ExceedsAvailable tests reserving too much.
func TestCapacity_Reserve_ExceedsAvailable(t *testing.T) {
	// Arrange
	capacity, _ := NewCapacity(1000)
	require.NoError(t, capacity.Reserve(800))

	// Act
	err := capacity.Reserve(300) // Would exceed capacity

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCapacityExceeded)
	assert.Equal(t, int64(800), capacity.Used) // Unchanged
}

// TestCapacity_Release_Success tests releasing capacity.
func TestCapacity_Release_Success(t *testing.T) {
	// Arrange
	capacity, _ := NewCapacity(1000)
	require.NoError(t, capacity.Reserve(500))

	// Act
	err := capacity.Release(200)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int64(300), capacity.Used)
	assert.Equal(t, int64(700), capacity.Available)
}

// TestCapacity_Release_ExceedsUsed tests releasing more than used.
func TestCapacity_Release_ExceedsUsed(t *testing.T) {
	// Arrange
	capacity, _ := NewCapacity(1000)
	require.NoError(t, capacity.Reserve(300))

	// Act
	err := capacity.Release(500) // More than used

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot release more than used capacity")
	assert.Equal(t, int64(300), capacity.Used) // Unchanged
}

// TestCapacity_UtilizationPercentage tests utilization calculation.
func TestCapacity_UtilizationPercentage(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		used     int64
		wantUtil float64
	}{
		{"empty", 1000, 0, 0.0},
		{"half full", 1000, 500, 50.0},
		{"fully used", 1000, 1000, 100.0},
		{"partial", 1000, 333, 33.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			capacity, _ := NewCapacity(tt.total)
			if tt.used > 0 {
				require.NoError(t, capacity.Reserve(tt.used))
			}

			// Act
			util := capacity.UtilizationPercentage()

			// Assert
			assert.InDelta(t, tt.wantUtil, util, 0.01)
		})
	}
}

// TestCapacity_CanFit tests capacity check.
func TestCapacity_CanFit(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		used     int64
		testSize int64
		wantFit  bool
	}{
		{"fits easily", 1000, 200, 500, true},
		{"exact fit", 1000, 900, 100, true},
		{"does not fit", 1000, 900, 200, false},
		{"empty capacity fits", 1000, 0, 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			capacity, _ := NewCapacity(tt.total)
			if tt.used > 0 {
				require.NoError(t, capacity.Reserve(tt.used))
			}

			// Act
			canFit := capacity.CanFit(tt.testSize)

			// Assert
			assert.Equal(t, tt.wantFit, canFit)
		})
	}
}

// =============================================================================
// Quality Tests
// =============================================================================

// TestNewQuality_Success tests creating valid Quality.
func TestNewQuality_Success(t *testing.T) {
	tests := []struct {
		name  string
		score float64
	}{
		{"minimum score", 0.0},
		{"medium score", 0.5},
		{"high score", 0.85},
		{"maximum score", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			quality, err := NewQuality(tt.score)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.score, quality.Score)
			assert.Equal(t, 0.0, quality.Detectability) // Default
			assert.Equal(t, 0.0, quality.FidelityScore) // Default
		})
	}
}

// TestNewQuality_Invalid tests creating invalid Quality.
func TestNewQuality_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		score float64
	}{
		{"below range", -0.5},
		{"above range", 1.5},
		{"way below", -10.0},
		{"way above", 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			quality, err := NewQuality(tt.score)

			// Assert
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidQuality)
			assert.Equal(t, Quality{}, quality)
		})
	}
}

// TestQuality_SetDetectability_Success tests setting valid detectability.
func TestQuality_SetDetectability_Success(t *testing.T) {
	// Arrange
	quality, _ := NewQuality(0.8)

	// Act
	err := quality.SetDetectability(0.2)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0.2, quality.Detectability)
}

// TestQuality_SetDetectability_Invalid tests setting invalid detectability.
func TestQuality_SetDetectability_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		score float64
	}{
		{"negative", -0.1},
		{"above range", 1.5},
	}

	// Arrange
	quality, _ := NewQuality(0.8)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := quality.SetDetectability(tt.score)

			// Assert
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidScore)
		})
	}
}

// TestQuality_SetFidelityScore_Success tests setting valid fidelity.
func TestQuality_SetFidelityScore_Success(t *testing.T) {
	// Arrange
	quality, _ := NewQuality(0.7)

	// Act
	err := quality.SetFidelityScore(0.95)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0.95, quality.FidelityScore)
}

// TestQuality_SetFidelityScore_Invalid tests setting invalid fidelity.
func TestQuality_SetFidelityScore_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		score float64
	}{
		{"below zero", -0.5},
		{"excessive", 2.0},
	}

	// Arrange
	quality, _ := NewQuality(0.7)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := quality.SetFidelityScore(tt.score)

			// Assert
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidScore)
		})
	}
}

// TestQuality_IsValid tests quality validation.
func TestQuality_IsValid(t *testing.T) {
	tests := []struct {
		name          string
		score         float64
		detectability float64
		fidelity      float64
		wantValid     bool
	}{
		{"all valid", 0.8, 0.2, 0.9, true},
		{"all zeros", 0.0, 0.0, 0.0, true},
		{"all ones", 1.0, 1.0, 1.0, true},
		{"invalid score", 1.5, 0.5, 0.8, false},
		{"invalid detectability", 0.8, 1.5, 0.9, false},
		{"invalid fidelity", 0.7, 0.3, -0.1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			quality := Quality{
				Score:         tt.score,
				Detectability: tt.detectability,
				FidelityScore: tt.fidelity,
			}

			// Act
			valid := quality.IsValid()

			// Assert
			assert.Equal(t, tt.wantValid, valid)
		})
	}
}

// TestQuality_IsAcceptable tests quality threshold checking.
func TestQuality_IsAcceptable(t *testing.T) {
	tests := []struct {
		name          string
		score         float64
		detectability float64
		fidelity      float64
		wantOK        bool
		desc          string
	}{
		{"high quality", 0.8, 0.2, 0.9, true, "meets all thresholds"},
		{"minimum acceptable", 0.6, 0.3, 0.7, true, "at threshold limits"},
		{"low score", 0.5, 0.2, 0.9, false, "score below 0.6"},
		{"high detectability", 0.8, 0.4, 0.9, false, "detectability above 0.3"},
		{"low fidelity", 0.8, 0.2, 0.6, false, "fidelity below 0.7"},
		{"all bad", 0.3, 0.8, 0.4, false, "fails all thresholds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			quality := Quality{
				Score:         tt.score,
				Detectability: tt.detectability,
				FidelityScore: tt.fidelity,
			}

			// Act
			acceptable := quality.IsAcceptable()

			// Assert
			assert.Equal(t, tt.wantOK, acceptable, tt.desc)
		})
	}
}
