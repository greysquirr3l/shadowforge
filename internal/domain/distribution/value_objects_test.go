package distribution

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// StrategyID Tests
// ============================================================================

func TestStrategyID_New_Success(t *testing.T) {
	id, err := NewStrategyID()

	require.NoError(t, err)
	assert.NotEmpty(t, id.String())
	_, parseErr := uuid.Parse(id.String())
	assert.NoError(t, parseErr)
	assert.False(t, id.IsZero())
}

func TestStrategyID_Parse_Success(t *testing.T) {
	validUUID := uuid.New().String()

	id, err := ParseStrategyID(validUUID)

	require.NoError(t, err)
	assert.Equal(t, validUUID, id.String())
	assert.False(t, id.IsZero())
}

func TestStrategyID_Parse_EmptyString(t *testing.T) {
	id, err := ParseStrategyID("")

	assert.ErrorIs(t, err, ErrInvalidStrategyID)
	assert.True(t, id.IsZero())
}

func TestStrategyID_Parse_InvalidUUID(t *testing.T) {
	id, err := ParseStrategyID("not-a-valid-uuid")

	assert.ErrorIs(t, err, ErrInvalidStrategyID)
	assert.True(t, id.IsZero())
}

func TestStrategyID_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		id       StrategyID
		expected bool
	}{
		{
			name:     "valid_id",
			id:       func() StrategyID { id, _ := NewStrategyID(); return id }(),
			expected: false,
		},
		{
			name:     "zero_value",
			id:       StrategyID{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.id.IsZero())
		})
	}
}

func TestStrategyID_String(t *testing.T) {
	validUUID := uuid.New().String()
	id, _ := ParseStrategyID(validUUID)

	assert.Equal(t, validUUID, id.String())
}

// ============================================================================
// ManifestID Tests
// ============================================================================

func TestManifestID_New_Success(t *testing.T) {
	id, err := NewManifestID()

	require.NoError(t, err)
	assert.NotEmpty(t, id.String())
	_, parseErr := uuid.Parse(id.String())
	assert.NoError(t, parseErr)
	assert.False(t, id.IsZero())
}

func TestManifestID_Parse_Success(t *testing.T) {
	validUUID := uuid.New().String()

	id, err := ParseManifestID(validUUID)

	require.NoError(t, err)
	assert.Equal(t, validUUID, id.String())
	assert.False(t, id.IsZero())
}

func TestManifestID_Parse_EmptyString(t *testing.T) {
	id, err := ParseManifestID("")

	assert.ErrorIs(t, err, ErrInvalidManifestID)
	assert.True(t, id.IsZero())
}

func TestManifestID_Parse_InvalidUUID(t *testing.T) {
	id, err := ParseManifestID("invalid-uuid")

	assert.ErrorIs(t, err, ErrInvalidManifestID)
	assert.True(t, id.IsZero())
}

func TestManifestID_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		id       ManifestID
		expected bool
	}{
		{
			name:     "valid_id",
			id:       func() ManifestID { id, _ := NewManifestID(); return id }(),
			expected: false,
		},
		{
			name:     "zero_value",
			id:       ManifestID{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.id.IsZero())
		})
	}
}

func TestManifestID_String(t *testing.T) {
	validUUID := uuid.New().String()
	id, _ := ParseManifestID(validUUID)

	assert.Equal(t, validUUID, id.String())
}

// ============================================================================
// PatternType Tests
// ============================================================================

func TestPatternType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		pattern  PatternType
		expected bool
	}{
		{"one_to_one", PatternOneToOne, true},
		{"one_to_many", PatternOneToMany, true},
		{"many_to_one", PatternManyToOne, true},
		{"many_to_many", PatternManyToMany, true},
		{"invalid", PatternType("invalid"), false},
		{"empty", PatternType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.pattern.IsValid())
		})
	}
}

func TestPatternType_String(t *testing.T) {
	tests := []struct {
		pattern  PatternType
		expected string
	}{
		{PatternOneToOne, "one_to_one"},
		{PatternOneToMany, "one_to_many"},
		{PatternManyToOne, "many_to_one"},
		{PatternManyToMany, "many_to_many"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.pattern.String())
		})
	}
}

// ============================================================================
// DistributionStatus Tests
// ============================================================================

func TestDistributionStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   DistributionStatus
		expected bool
	}{
		{"pending", StatusPending, true},
		{"in_progress", StatusInProgress, true},
		{"complete", StatusComplete, true},
		{"failed", StatusFailed, true},
		{"revoked", StatusRevoked, true},
		{"invalid", DistributionStatus("invalid"), false},
		{"empty", DistributionStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestDistributionStatus_String(t *testing.T) {
	tests := []struct {
		status   DistributionStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusInProgress, "in_progress"},
		{StatusComplete, "complete"},
		{StatusFailed, "failed"},
		{StatusRevoked, "revoked"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

// ============================================================================
// ShardMetadata Tests
// ============================================================================

func TestShardMetadata_New_Success(t *testing.T) {
	tests := []struct {
		name        string
		index       int
		size        int64
		checksum    string
		mediaID     string
		recipientID string
	}{
		{
			name:        "valid_metadata",
			index:       0,
			size:        1024,
			checksum:    "abc123",
			mediaID:     "media-1",
			recipientID: "recipient-1",
		},
		{
			name:        "large_shard",
			index:       5,
			size:        104857600,
			checksum:    "def456",
			mediaID:     "media-2",
			recipientID: "recipient-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := NewShardMetadata(
				tt.index,
				tt.size,
				tt.checksum,
				tt.mediaID,
				tt.recipientID,
			)

			require.NoError(t, err)
			assert.Equal(t, tt.index, metadata.Index)
			assert.Equal(t, tt.size, metadata.Size)
			assert.Equal(t, tt.checksum, metadata.Checksum)
			assert.Equal(t, tt.mediaID, metadata.MediaID)
			assert.Equal(t, tt.recipientID, metadata.RecipientID)
		})
	}
}

func TestShardMetadata_New_InvalidIndex(t *testing.T) {
	metadata, err := NewShardMetadata(-1, 1024, "abc123", "media-1", "recipient-1")

	assert.ErrorIs(t, err, ErrInvalidShardIndex)
	assert.Equal(t, ShardMetadata{}, metadata)
}

func TestShardMetadata_New_InvalidSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
	}{
		{"zero_size", 0},
		{"negative_size", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := NewShardMetadata(0, tt.size, "abc123", "media-1", "recipient-1")

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid shard size")
			assert.Equal(t, ShardMetadata{}, metadata)
		})
	}
}

func TestShardMetadata_New_EmptyChecksum(t *testing.T) {
	metadata, err := NewShardMetadata(0, 1024, "", "media-1", "recipient-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum is required")
	assert.Equal(t, ShardMetadata{}, metadata)
}

func TestShardMetadata_New_EmptyMediaID(t *testing.T) {
	metadata, err := NewShardMetadata(0, 1024, "abc123", "", "recipient-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "media ID is required")
	assert.Equal(t, ShardMetadata{}, metadata)
}

func TestShardMetadata_Validate_Success(t *testing.T) {
	metadata, _ := NewShardMetadata(0, 1024, "abc123", "media-1", "recipient-1")

	err := metadata.Validate()

	assert.NoError(t, err)
}

func TestShardMetadata_Validate_InvalidIndex(t *testing.T) {
	metadata := ShardMetadata{
		Index:    -1,
		Size:     1024,
		Checksum: "abc123",
		MediaID:  "media-1",
	}

	err := metadata.Validate()

	assert.ErrorIs(t, err, ErrInvalidShardIndex)
}

func TestShardMetadata_Validate_InvalidSize(t *testing.T) {
	metadata := ShardMetadata{
		Index:    0,
		Size:     0,
		Checksum: "abc123",
		MediaID:  "media-1",
	}

	err := metadata.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid shard size")
}

func TestShardMetadata_Validate_EmptyChecksum(t *testing.T) {
	metadata := ShardMetadata{
		Index:   0,
		Size:    1024,
		MediaID: "media-1",
	}

	err := metadata.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum is required")
}

func TestShardMetadata_Validate_EmptyMediaID(t *testing.T) {
	metadata := ShardMetadata{
		Index:    0,
		Size:     1024,
		Checksum: "abc123",
	}

	err := metadata.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "media ID is required")
}
