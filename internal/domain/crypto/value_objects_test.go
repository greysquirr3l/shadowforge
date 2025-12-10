package crypto_test

import (
	"testing"

	"github.com/greysquirr3l/shadowforge/internal/domain/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPayloadID_NewPayloadID tests the creation of PayloadID with validation
func TestPayloadID_NewPayloadID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantErr   bool
		errCheck  error
		wantValue string
	}{
		{
			name:      "valid_uuid",
			id:        "123e4567-e89b-12d3-a456-426614174000",
			wantErr:   false,
			wantValue: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:      "valid_custom_string",
			id:        "payload-001",
			wantErr:   false,
			wantValue: "payload-001",
		},
		{
			name:     "empty_string",
			id:       "",
			wantErr:  true,
			errCheck: crypto.ErrInvalidPayloadID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			payloadID, err := crypto.NewPayloadID(tt.id)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.errCheck)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantValue, payloadID.String())
			}
		})
	}
}

// TestPayloadID_GeneratePayloadID tests random UUID generation
func TestPayloadID_GeneratePayloadID(t *testing.T) {
	// Act
	id1 := crypto.GeneratePayloadID()
	id2 := crypto.GeneratePayloadID()

	// Assert
	assert.NotEmpty(t, id1.String())
	assert.NotEmpty(t, id2.String())
	assert.NotEqual(t, id1.String(), id2.String(), "generated IDs should be unique")
	assert.Len(t, id1.String(), 36, "UUID v4 should be 36 characters")
}

// TestPayloadID_String tests the string representation
func TestPayloadID_String(t *testing.T) {
	// Arrange
	id, _ := crypto.NewPayloadID("test-payload-id")

	// Act
	result := id.String()

	// Assert
	assert.Equal(t, "test-payload-id", result)
}

// TestPayloadID_IsZero tests zero value detection
func TestPayloadID_IsZero(t *testing.T) {
	tests := []struct {
		name      string
		payloadID crypto.PayloadID
		expected  bool
	}{
		{
			name:      "zero_value",
			payloadID: crypto.PayloadID{},
			expected:  true,
		},
		{
			name:      "generated_id",
			payloadID: crypto.GeneratePayloadID(),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.payloadID.IsZero()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPayloadID_Equals tests equality comparison
func TestPayloadID_Equals(t *testing.T) {
	// Arrange
	id1, _ := crypto.NewPayloadID("test-id-1")
	id2, _ := crypto.NewPayloadID("test-id-1")
	id3, _ := crypto.NewPayloadID("test-id-2")

	// Act & Assert
	assert.True(t, id1.Equals(id2), "identical IDs should be equal")
	assert.False(t, id1.Equals(id3), "different IDs should not be equal")
}

// TestPQCAlgorithm_Constants tests the algorithm constants
func TestPQCAlgorithm_Constants(t *testing.T) {
	t.Run("kyber1024", func(t *testing.T) {
		assert.Equal(t, "kyber1024", crypto.Kyber1024.Name())
		assert.Equal(t, crypto.Kyber1024KeySize, crypto.Kyber1024.KeySize())
		assert.Equal(t, 1568, crypto.Kyber1024KeySize)
		assert.True(t, crypto.Kyber1024.IsValid())
	})

	t.Run("dilithium3", func(t *testing.T) {
		assert.Equal(t, "dilithium3", crypto.Dilithium3.Name())
		assert.Equal(t, crypto.Dilithium3KeySize, crypto.Dilithium3.KeySize())
		assert.Equal(t, 1952, crypto.Dilithium3KeySize)
		assert.True(t, crypto.Dilithium3.IsValid())
	})

	t.Run("unknown_algo", func(t *testing.T) {
		assert.Equal(t, "unknown", crypto.UnknownAlgo.Name())
		assert.Equal(t, 0, crypto.UnknownAlgo.KeySize())
		assert.False(t, crypto.UnknownAlgo.IsValid())
	})

	t.Run("dilithium3_signature_size", func(t *testing.T) {
		assert.Equal(t, 3293, crypto.Dilithium3SignSize)
	})
}

// TestPQCAlgorithm_NewPQCAlgorithm tests algorithm creation from string
func TestPQCAlgorithm_NewPQCAlgorithm(t *testing.T) {
	tests := []struct {
		name         string
		algoName     string
		wantErr      bool
		errCheck     error
		expectedAlgo crypto.PQCAlgorithm
	}{
		{
			name:         "kyber1024",
			algoName:     "kyber1024",
			wantErr:      false,
			expectedAlgo: crypto.Kyber1024,
		},
		{
			name:         "dilithium3",
			algoName:     "dilithium3",
			wantErr:      false,
			expectedAlgo: crypto.Dilithium3,
		},
		{
			name:         "invalid_algorithm",
			algoName:     "invalid",
			wantErr:      true,
			errCheck:     crypto.ErrInvalidAlgorithm,
			expectedAlgo: crypto.UnknownAlgo,
		},
		{
			name:         "empty_string",
			algoName:     "",
			wantErr:      true,
			errCheck:     crypto.ErrInvalidAlgorithm,
			expectedAlgo: crypto.UnknownAlgo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			algo, err := crypto.NewPQCAlgorithm(tt.algoName)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.errCheck)
				assert.Equal(t, tt.expectedAlgo.Name(), algo.Name())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedAlgo.Name(), algo.Name())
				assert.Equal(t, tt.expectedAlgo.KeySize(), algo.KeySize())
			}
		})
	}
}

// TestPQCAlgorithm_Name tests the Name method
func TestPQCAlgorithm_Name(t *testing.T) {
	tests := []struct {
		name     string
		algo     crypto.PQCAlgorithm
		expected string
	}{
		{
			name:     "kyber1024",
			algo:     crypto.Kyber1024,
			expected: "kyber1024",
		},
		{
			name:     "dilithium3",
			algo:     crypto.Dilithium3,
			expected: "dilithium3",
		},
		{
			name:     "unknown",
			algo:     crypto.UnknownAlgo,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.algo.Name()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPQCAlgorithm_KeySize tests the KeySize method
func TestPQCAlgorithm_KeySize(t *testing.T) {
	tests := []struct {
		name     string
		algo     crypto.PQCAlgorithm
		expected int
	}{
		{
			name:     "kyber1024",
			algo:     crypto.Kyber1024,
			expected: 1568,
		},
		{
			name:     "dilithium3",
			algo:     crypto.Dilithium3,
			expected: 1952,
		},
		{
			name:     "unknown",
			algo:     crypto.UnknownAlgo,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.algo.KeySize()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPQCAlgorithm_IsValid tests the IsValid method
func TestPQCAlgorithm_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		algo     crypto.PQCAlgorithm
		expected bool
	}{
		{
			name:     "kyber1024_is_valid",
			algo:     crypto.Kyber1024,
			expected: true,
		},
		{
			name:     "dilithium3_is_valid",
			algo:     crypto.Dilithium3,
			expected: true,
		},
		{
			name:     "unknown_is_invalid",
			algo:     crypto.UnknownAlgo,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.algo.IsValid()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPQCAlgorithm_String tests the String method
func TestPQCAlgorithm_String(t *testing.T) {
	tests := []struct {
		name     string
		algo     crypto.PQCAlgorithm
		expected string
	}{
		{
			name:     "kyber1024",
			algo:     crypto.Kyber1024,
			expected: "kyber1024",
		},
		{
			name:     "dilithium3",
			algo:     crypto.Dilithium3,
			expected: "dilithium3",
		},
		{
			name:     "unknown",
			algo:     crypto.UnknownAlgo,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.algo.String()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPQCAlgorithm_Equals tests the Equals method
func TestPQCAlgorithm_Equals(t *testing.T) {
	tests := []struct {
		name     string
		algo1    crypto.PQCAlgorithm
		algo2    crypto.PQCAlgorithm
		expected bool
	}{
		{
			name:     "kyber1024_equals_kyber1024",
			algo1:    crypto.Kyber1024,
			algo2:    crypto.Kyber1024,
			expected: true,
		},
		{
			name:     "dilithium3_equals_dilithium3",
			algo1:    crypto.Dilithium3,
			algo2:    crypto.Dilithium3,
			expected: true,
		},
		{
			name:     "kyber1024_not_equals_dilithium3",
			algo1:    crypto.Kyber1024,
			algo2:    crypto.Dilithium3,
			expected: false,
		},
		{
			name:     "unknown_equals_unknown",
			algo1:    crypto.UnknownAlgo,
			algo2:    crypto.UnknownAlgo,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.algo1.Equals(tt.algo2)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPQCAlgorithm_ComprehensiveProperties tests algorithm properties comprehensively
func TestPQCAlgorithm_ComprehensiveProperties(t *testing.T) {
	t.Run("kyber1024_properties", func(t *testing.T) {
		algo := crypto.Kyber1024

		assert.Equal(t, "kyber1024", algo.Name())
		assert.Equal(t, "kyber1024", algo.String())
		assert.Equal(t, 1568, algo.KeySize())
		assert.True(t, algo.IsValid())
		assert.True(t, algo.Equals(crypto.Kyber1024))
		assert.False(t, algo.Equals(crypto.Dilithium3))
	})

	t.Run("dilithium3_properties", func(t *testing.T) {
		algo := crypto.Dilithium3

		assert.Equal(t, "dilithium3", algo.Name())
		assert.Equal(t, "dilithium3", algo.String())
		assert.Equal(t, 1952, algo.KeySize())
		assert.True(t, algo.IsValid())
		assert.True(t, algo.Equals(crypto.Dilithium3))
		assert.False(t, algo.Equals(crypto.Kyber1024))
	})

	t.Run("unknown_properties", func(t *testing.T) {
		algo := crypto.UnknownAlgo

		assert.Equal(t, "unknown", algo.Name())
		assert.Equal(t, "unknown", algo.String())
		assert.Equal(t, 0, algo.KeySize())
		assert.False(t, algo.IsValid())
		assert.True(t, algo.Equals(crypto.UnknownAlgo))
		assert.False(t, algo.Equals(crypto.Kyber1024))
	})
}

// TestPQCAlgorithm_NISTStandardCompliance tests NIST standard key sizes
func TestPQCAlgorithm_NISTStandardCompliance(t *testing.T) {
	t.Run("kyber1024_nist_level5", func(t *testing.T) {
		// Kyber-1024 provides NIST security level 5 (equivalent to AES-256)
		assert.Equal(t, 1568, crypto.Kyber1024.KeySize())
		assert.Equal(t, 1568, crypto.Kyber1024KeySize)
	})

	t.Run("dilithium3_nist_level3", func(t *testing.T) {
		// Dilithium3 provides NIST security level 3 (equivalent to AES-192)
		assert.Equal(t, 1952, crypto.Dilithium3.KeySize())
		assert.Equal(t, 1952, crypto.Dilithium3KeySize)
		assert.Equal(t, 3293, crypto.Dilithium3SignSize)
	})
}
