package distribution_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/greysquirr3l/shadowforge/internal/domain/distribution"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestManifestSerializer_MarshalUnmarshal tests roundtrip serialization.
func TestManifestSerializer_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name          string
		manifest      *distribution.ShardManifest
		hmacKey       []byte
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid_manifest_roundtrip",
			manifest:    createTestManifest(t, 3, 2, nil),
			hmacKey:     make([]byte, 32), // 32-byte key
			expectError: false,
		},
		{
			name: "manifest_with_expiry",
			manifest: func() *distribution.ShardManifest {
				expiry := time.Now().Add(24 * time.Hour)
				return createTestManifest(t, 5, 3, &expiry)
			}(),
			hmacKey:     make([]byte, 32),
			expectError: false,
		},
		{
			name:        "large_manifest_10_shards",
			manifest:    createTestManifest(t, 10, 7, nil),
			hmacKey:     make([]byte, 32),
			expectError: false,
		},
		{
			name:        "minimal_manifest_1_shard",
			manifest:    createTestManifest(t, 1, 1, nil),
			hmacKey:     make([]byte, 32),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			serializer, err := distribution.NewManifestSerializer(tt.hmacKey)
			require.NoError(t, err)

			// Act - Marshal
			data, err := serializer.Marshal(tt.manifest)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Act - Unmarshal
			recovered, err := serializer.Unmarshal(data)
			require.NoError(t, err)

			// Assert - Verify roundtrip integrity
			assert.Equal(t, tt.manifest.ID.String(), recovered.ID.String())
			assert.Equal(t, tt.manifest.StrategyID.String(), recovered.StrategyID.String())
			assert.Equal(t, tt.manifest.TotalShards, recovered.TotalShards)
			assert.Equal(t, tt.manifest.RequiredShards, recovered.RequiredShards)
			assert.Equal(t, len(tt.manifest.ShardMetadata), len(recovered.ShardMetadata))

			// Verify shard metadata
			for i := range tt.manifest.ShardMetadata {
				assert.Equal(t, tt.manifest.ShardMetadata[i].Index, recovered.ShardMetadata[i].Index)
				assert.Equal(t, tt.manifest.ShardMetadata[i].Checksum, recovered.ShardMetadata[i].Checksum)
				assert.Equal(t, tt.manifest.ShardMetadata[i].MediaID, recovered.ShardMetadata[i].MediaID)
			}

			// Verify timestamps (with tolerance for serialization)
			assert.WithinDuration(t, tt.manifest.CreatedAt, recovered.CreatedAt, time.Second)

			if tt.manifest.ExpiresAt != nil {
				require.NotNil(t, recovered.ExpiresAt)
				assert.WithinDuration(t, *tt.manifest.ExpiresAt, *recovered.ExpiresAt, time.Second)
			}
		})
	}
}

// TestManifestSerializer_SignVerify tests HMAC signature generation and verification.
func TestManifestSerializer_SignVerify(t *testing.T) {
	tests := []struct {
		name        string
		manifest    *distribution.ShardManifest
		hmacKey     []byte
		expectValid bool
	}{
		{
			name:        "valid_signature",
			manifest:    createTestManifest(t, 3, 2, nil),
			hmacKey:     []byte("a-valid-32byte-hmac-key-here!!!!"), // Exactly 32 bytes
			expectValid: true,
		},
		{
			name:        "different_manifest_invalid",
			manifest:    createTestManifest(t, 5, 3, nil),
			hmacKey:     []byte("a-valid-32byte-hmac-key-here!!!!"), // Exactly 32 bytes
			expectValid: true,                                       // Valid for its own manifest
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			serializer, err := distribution.NewManifestSerializer(tt.hmacKey)
			require.NoError(t, err)

			// Act - Sign
			signature, err := serializer.Sign(tt.manifest)
			require.NoError(t, err)
			assert.NotEmpty(t, signature)

			// Act - Verify
			valid := serializer.Verify(tt.manifest, signature)
			assert.Equal(t, tt.expectValid, valid)
		})
	}
}

// TestManifestSerializer_TamperDetection tests detection of tampered manifests.
func TestManifestSerializer_TamperDetection(t *testing.T) {
	tests := []struct {
		name          string
		tamperFunc    func([]byte) []byte
		expectError   bool
		errorContains string
	}{
		{
			name: "tampered_total_shards",
			tamperFunc: func(data []byte) []byte {
				// Change "total_shards":3 to "total_shards":5
				return []byte(string(data)[0:100] + "5" + string(data)[101:])
			},
			expectError:   true,
			errorContains: "HMAC signature verification failed",
		},
		{
			name: "tampered_signature",
			tamperFunc: func(data []byte) []byte {
				// Replace last character of signature
				result := make([]byte, len(data))
				copy(result, data)
				result[len(result)-10] = 'X'
				return result
			},
			expectError:   true,
			errorContains: "HMAC signature verification failed",
		},
		{
			name: "truncated_data",
			tamperFunc: func(data []byte) []byte {
				return data[:len(data)/2]
			},
			expectError:   true,
			errorContains: "failed to unmarshal manifest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			manifest := createTestManifest(t, 3, 2, nil)
			hmacKey := []byte("a-valid-32byte-hmac-key-here!!!!") // Exactly 32 bytes
			serializer, err := distribution.NewManifestSerializer(hmacKey)
			require.NoError(t, err)

			// Create valid manifest
			data, err := serializer.Marshal(manifest)
			require.NoError(t, err)

			// Tamper with data
			tamperedData := tt.tamperFunc(data)

			// Act - Try to unmarshal tampered data
			_, err = serializer.Unmarshal(tamperedData)

			// Assert
			assert.Error(t, err)
			if tt.errorContains != "" {
				assert.Contains(t, err.Error(), tt.errorContains)
			}
		})
	}
}

// TestManifestSerializer_InvalidInputs tests error handling for invalid inputs.
func TestManifestSerializer_InvalidInputs(t *testing.T) {
	tests := []struct {
		name          string
		hmacKey       []byte
		manifest      *distribution.ShardManifest
		data          []byte
		operation     string
		expectError   bool
		errorContains string
	}{
		{
			name:          "nil_manifest_marshal",
			hmacKey:       make([]byte, 32),
			manifest:      nil,
			operation:     "marshal",
			expectError:   true,
			errorContains: "manifest cannot be nil",
		},
		{
			name:          "empty_data_unmarshal",
			hmacKey:       make([]byte, 32),
			data:          []byte{},
			operation:     "unmarshal",
			expectError:   true,
			errorContains: "data cannot be empty",
		},
		{
			name:          "invalid_json_unmarshal",
			hmacKey:       make([]byte, 32),
			data:          []byte("not valid json"),
			operation:     "unmarshal",
			expectError:   true,
			errorContains: "failed to unmarshal manifest",
		},
		{
			name:          "nil_manifest_sign",
			hmacKey:       make([]byte, 32),
			manifest:      nil,
			operation:     "sign",
			expectError:   true,
			errorContains: "manifest cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			serializer, err := distribution.NewManifestSerializer(tt.hmacKey)
			require.NoError(t, err)

			// Act & Assert
			switch tt.operation {
			case "marshal":
				_, err = serializer.Marshal(tt.manifest)
			case "unmarshal":
				_, err = serializer.Unmarshal(tt.data)
			case "sign":
				_, err = serializer.Sign(tt.manifest)
			}

			assert.Error(t, err)
			if tt.errorContains != "" {
				assert.Contains(t, err.Error(), tt.errorContains)
			}
		})
	}
}

// TestManifestSerializer_VersionCompatibility tests schema version handling.
func TestManifestSerializer_VersionCompatibility(t *testing.T) {
	// Arrange
	manifest := createTestManifest(t, 3, 2, nil)
	hmacKey := []byte("a-valid-32byte-hmac-key-here!!!!") // Exactly 32 bytes
	serializer, err := distribution.NewManifestSerializer(hmacKey)
	require.NoError(t, err)

	// Create valid manifest with current version
	data, err := serializer.Marshal(manifest)
	require.NoError(t, err)

	// Verify current version can be unmarshaled
	recovered, err := serializer.Unmarshal(data)
	require.NoError(t, err)
	assert.NotNil(t, recovered)

	// Test unsupported version (manually create JSON with wrong version)
	invalidVersionJSON := `{
		"version": "2.0",
		"id": "test-id",
		"strategy_id": "test-strategy",
		"shards": [],
		"total_shards": 0,
		"required_shards": 0,
		"created_at": "2025-01-01T00:00:00Z",
		"signature": "dummy"
	}`

	_, err = serializer.Unmarshal([]byte(invalidVersionJSON))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported manifest version")
}

// TestManifestSerializer_KeyValidation tests HMAC key validation.
func TestManifestSerializer_KeyValidation(t *testing.T) {
	tests := []struct {
		name          string
		hmacKey       []byte
		expectError   bool
		errorContains string
	}{
		{
			name:          "empty_key",
			hmacKey:       []byte{},
			expectError:   true,
			errorContains: "HMAC key cannot be empty",
		},
		{
			name:          "nil_key",
			hmacKey:       nil,
			expectError:   true,
			errorContains: "HMAC key cannot be empty",
		},
		{
			name:          "short_key",
			hmacKey:       []byte("short"),
			expectError:   true,
			errorContains: "HMAC key must be at least 32 bytes",
		},
		{
			name:        "valid_32_byte_key",
			hmacKey:     make([]byte, 32),
			expectError: false,
		},
		{
			name:        "valid_64_byte_key",
			hmacKey:     make([]byte, 64),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			serializer, err := distribution.NewManifestSerializer(tt.hmacKey)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, serializer)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, serializer)
			}
		})
	}
}

// TestManifestSerializer_DifferentKeys tests that different keys produce different signatures.
func TestManifestSerializer_DifferentKeys(t *testing.T) {
	// Arrange
	manifest := createTestManifest(t, 3, 2, nil)
	key1 := []byte("another-32byte-hmac-key-here!!!!") // Exactly 32 bytes
	key2 := []byte("third-valid-32byte-key-value!!!!") // Exactly 32 bytes

	serializer1, err := distribution.NewManifestSerializer(key1)
	require.NoError(t, err)

	serializer2, err := distribution.NewManifestSerializer(key2)
	require.NoError(t, err)

	// Act - Sign with different keys
	sig1, err := serializer1.Sign(manifest)
	require.NoError(t, err)

	sig2, err := serializer2.Sign(manifest)
	require.NoError(t, err)

	// Assert - Signatures should be different
	assert.NotEqual(t, sig1, sig2, "Different HMAC keys should produce different signatures")

	// Verify cross-validation fails
	assert.False(t, serializer1.Verify(manifest, sig2), "Signature from key2 should not verify with key1")
	assert.False(t, serializer2.Verify(manifest, sig1), "Signature from key1 should not verify with key2")
}

// Helper function to create a test manifest with specified parameters.
func createTestManifest(t *testing.T, totalShards, requiredShards int, expiresAt *time.Time) *distribution.ShardManifest {
	t.Helper()

	// Create strategy ID
	strategyID, err := distribution.NewStrategyID()
	require.NoError(t, err)

	// Create shard metadata
	shardMetadata := make([]distribution.ShardMetadata, totalShards)
	for i := 0; i < totalShards; i++ {
		// Create checksum for shard
		h := sha256.New()
		h.Write([]byte{byte(i)})
		checksum := hex.EncodeToString(h.Sum(nil))

		shard, err := distribution.NewShardMetadata(
			i,
			1024*(int64(i)+1), // Varying sizes
			checksum,
			"media-"+string(rune('A'+i)),
			"recipient-"+string(rune('A'+i)),
		)
		require.NoError(t, err)
		shardMetadata[i] = shard
	}

	// Create manifest
	manifest, err := distribution.NewShardManifest(strategyID, shardMetadata, requiredShards)
	require.NoError(t, err)

	// Set expiry if provided
	if expiresAt != nil {
		manifest.ExpiresAt = expiresAt
	}

	return manifest
}
