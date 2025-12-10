package crypto_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

// TestCryptoPayload_New tests CryptoPayload creation with validation
func TestCryptoPayload_New(t *testing.T) {
	tests := []struct {
		name      string
		setupID   func() crypto.PayloadID
		data      []byte
		algorithm crypto.PQCAlgorithm
		wantErr   bool
	}{
		{
			name:      "valid_kyber1024_payload",
			setupID:   func() crypto.PayloadID { return crypto.GeneratePayloadID() },
			data:      []byte("test payload data"),
			algorithm: crypto.Kyber1024,
			wantErr:   false,
		},
		{
			name:      "valid_dilithium3_payload",
			setupID:   func() crypto.PayloadID { return crypto.GeneratePayloadID() },
			data:      []byte("signature test data"),
			algorithm: crypto.Dilithium3,
			wantErr:   false,
		},
		{
			name:      "empty_data_returns_error",
			setupID:   func() crypto.PayloadID { return crypto.GeneratePayloadID() },
			data:      []byte{},
			algorithm: crypto.Kyber1024,
			wantErr:   true,
		},
		{
			name:      "nil_data_returns_error",
			setupID:   func() crypto.PayloadID { return crypto.GeneratePayloadID() },
			data:      nil,
			algorithm: crypto.Kyber1024,
			wantErr:   true,
		},
		{
			name:      "invalid_algorithm_returns_error",
			setupID:   func() crypto.PayloadID { return crypto.GeneratePayloadID() },
			data:      []byte("test data"),
			algorithm: crypto.UnknownAlgo,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.setupID()
			payload, err := crypto.NewCryptoPayload(id, tt.data, tt.algorithm)

			if tt.wantErr {
				assert.Error(t, err, "expected error but got nil")
				assert.Nil(t, payload, "payload should be nil on error")
			} else {
				require.NoError(t, err, "unexpected error creating payload")
				require.NotNil(t, payload, "payload should not be nil")
				assert.Equal(t, id, payload.ID)
				assert.Equal(t, tt.data, payload.Data)
				assert.Equal(t, tt.algorithm, payload.Algorithm)
				assert.False(t, payload.EncryptedAt.IsZero(), "EncryptedAt should be set")
				assert.WithinDuration(t, time.Now(), payload.EncryptedAt, 5*time.Second)
			}
		})
	}
}

// TestCryptoPayload_SetSignature tests attaching signatures to payloads
func TestCryptoPayload_SetSignature(t *testing.T) {
	tests := []struct {
		name      string
		signature []byte
		publicKey []byte
		wantErr   bool
	}{
		{
			name:      "valid_signature_and_key",
			signature: []byte("mock_dilithium3_signature"),
			publicKey: []byte("mock_dilithium3_public_key"),
			wantErr:   false,
		},
		{
			name:      "empty_signature_returns_error",
			signature: []byte{},
			publicKey: []byte("public_key"),
			wantErr:   true,
		},
		{
			name:      "nil_signature_returns_error",
			signature: nil,
			publicKey: []byte("public_key"),
			wantErr:   true,
		},
		{
			name:      "empty_public_key_returns_error",
			signature: []byte("signature"),
			publicKey: []byte{},
			wantErr:   true,
		},
		{
			name:      "nil_public_key_returns_error",
			signature: []byte("signature"),
			publicKey: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create a valid payload
			id := crypto.GeneratePayloadID()
			payload, err := crypto.NewCryptoPayload(id, []byte("test data"), crypto.Dilithium3)
			require.NoError(t, err, "setup: failed to create payload")

			// Act
			err = payload.SetSignature(tt.signature, tt.publicKey)

			// Assert
			if tt.wantErr {
				assert.Error(t, err, "expected error but got nil")
			} else {
				require.NoError(t, err, "unexpected error setting signature")
				assert.Equal(t, tt.signature, payload.Signature)
				assert.Equal(t, tt.publicKey, payload.PublicKey)
			}
		})
	}
}

// TestCryptoPayload_SetExpiration tests expiration management
func TestCryptoPayload_SetExpiration(t *testing.T) {
	tests := []struct {
		name        string
		setupExpiry func(createdAt time.Time) time.Time
		wantErr     bool
	}{
		{
			name: "valid_future_expiration",
			setupExpiry: func(createdAt time.Time) time.Time {
				return createdAt.Add(24 * time.Hour)
			},
			wantErr: false,
		},
		{
			name: "expiration_before_creation_returns_error",
			setupExpiry: func(createdAt time.Time) time.Time {
				return createdAt.Add(-24 * time.Hour)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			id := crypto.GeneratePayloadID()
			payload, err := crypto.NewCryptoPayload(id, []byte("test data"), crypto.Kyber1024)
			require.NoError(t, err, "setup: failed to create payload")

			expiry := tt.setupExpiry(payload.EncryptedAt)

			// Act
			err = payload.SetExpiration(expiry)

			// Assert
			if tt.wantErr {
				assert.Error(t, err, "expected error but got nil")
			} else {
				require.NoError(t, err, "unexpected error setting expiration")
				require.NotNil(t, payload.ExpiresAt, "ExpiresAt should not be nil")
				assert.Equal(t, expiry, *payload.ExpiresAt)
			}
		})
	}
}

// TestCryptoPayload_IsExpired tests expiration checking
func TestCryptoPayload_IsExpired(t *testing.T) {
	t.Run("no_expiration_never_expires", func(t *testing.T) {
		id := crypto.GeneratePayloadID()
		payload, err := crypto.NewCryptoPayload(id, []byte("test data"), crypto.Kyber1024)
		require.NoError(t, err)

		assert.False(t, payload.IsExpired(), "payload without expiration should not expire")
	})

	t.Run("future_expiration_not_expired", func(t *testing.T) {
		id := crypto.GeneratePayloadID()
		payload, err := crypto.NewCryptoPayload(id, []byte("test data"), crypto.Kyber1024)
		require.NoError(t, err)

		futureExpiry := time.Now().Add(1 * time.Hour)
		err = payload.SetExpiration(futureExpiry)
		require.NoError(t, err)

		assert.False(t, payload.IsExpired(), "future expiration should not be expired")
	})

	t.Run("past_expiration_is_expired", func(t *testing.T) {
		id := crypto.GeneratePayloadID()
		payload, err := crypto.NewCryptoPayload(id, []byte("test data"), crypto.Kyber1024)
		require.NoError(t, err)

		// Manually set past expiration (bypassing validation for test)
		pastExpiry := time.Now().Add(-1 * time.Hour)
		payload.ExpiresAt = &pastExpiry

		assert.True(t, payload.IsExpired(), "past expiration should be expired")
	})
}

// TestCryptoPayload_Validate tests payload validation
func TestCryptoPayload_Validate(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func() *crypto.CryptoPayload
		wantErr    bool
		errContain string
	}{
		{
			name: "valid_payload_passes_validation",
			setupFunc: func() *crypto.CryptoPayload {
				id := crypto.GeneratePayloadID()
				payload, _ := crypto.NewCryptoPayload(id, []byte("test data"), crypto.Kyber1024)
				return payload
			},
			wantErr: false,
		},
		{
			name: "zero_id_fails_validation",
			setupFunc: func() *crypto.CryptoPayload {
				id := crypto.GeneratePayloadID()
				payload, _ := crypto.NewCryptoPayload(id, []byte("test data"), crypto.Kyber1024)
				payload.ID = crypto.PayloadID{} // Zero value
				return payload
			},
			wantErr:    true,
			errContain: "invalid payload ID",
		},
		{
			name: "empty_data_fails_validation",
			setupFunc: func() *crypto.CryptoPayload {
				id := crypto.GeneratePayloadID()
				payload, _ := crypto.NewCryptoPayload(id, []byte("test data"), crypto.Kyber1024)
				payload.Data = []byte{} // Empty data
				return payload
			},
			wantErr:    true,
			errContain: "payload data cannot be empty",
		},
		{
			name: "invalid_algorithm_fails_validation",
			setupFunc: func() *crypto.CryptoPayload {
				id := crypto.GeneratePayloadID()
				payload, _ := crypto.NewCryptoPayload(id, []byte("test data"), crypto.Kyber1024)
				payload.Algorithm = crypto.UnknownAlgo
				return payload
			},
			wantErr:    true,
			errContain: "invalid or unsupported PQC algorithm",
		},
		{
			name: "expired_payload_fails_validation",
			setupFunc: func() *crypto.CryptoPayload {
				id := crypto.GeneratePayloadID()
				payload, _ := crypto.NewCryptoPayload(id, []byte("test data"), crypto.Kyber1024)
				pastExpiry := time.Now().Add(-1 * time.Hour)
				payload.ExpiresAt = &pastExpiry
				return payload
			},
			wantErr:    true,
			errContain: "expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := tt.setupFunc()

			err := payload.Validate()

			if tt.wantErr {
				require.Error(t, err, "expected error but got nil")
				if tt.errContain != "" {
					assert.Contains(t, err.Error(), tt.errContain)
				}
			} else {
				assert.NoError(t, err, "unexpected error during validation")
			}
		})
	}
}

// TestKeyPair_New tests KeyPair creation with validation
func TestKeyPair_New(t *testing.T) {
	tests := []struct {
		name       string
		algorithm  crypto.PQCAlgorithm
		publicKey  []byte
		privateKey []byte
		wantErr    bool
	}{
		{
			name:       "valid_kyber1024_keypair",
			algorithm:  crypto.Kyber1024,
			publicKey:  []byte("mock_kyber1024_public_key"),
			privateKey: []byte("mock_kyber1024_private_key"),
			wantErr:    false,
		},
		{
			name:       "valid_dilithium3_keypair",
			algorithm:  crypto.Dilithium3,
			publicKey:  []byte("mock_dilithium3_public_key"),
			privateKey: []byte("mock_dilithium3_private_key"),
			wantErr:    false,
		},
		{
			name:       "invalid_algorithm_returns_error",
			algorithm:  crypto.UnknownAlgo,
			publicKey:  []byte("public_key"),
			privateKey: []byte("private_key"),
			wantErr:    true,
		},
		{
			name:       "empty_public_key_returns_error",
			algorithm:  crypto.Kyber1024,
			publicKey:  []byte{},
			privateKey: []byte("private_key"),
			wantErr:    true,
		},
		{
			name:       "nil_public_key_returns_error",
			algorithm:  crypto.Kyber1024,
			publicKey:  nil,
			privateKey: []byte("private_key"),
			wantErr:    true,
		},
		{
			name:       "empty_private_key_returns_error",
			algorithm:  crypto.Kyber1024,
			publicKey:  []byte("public_key"),
			privateKey: []byte{},
			wantErr:    true,
		},
		{
			name:       "nil_private_key_returns_error",
			algorithm:  crypto.Kyber1024,
			publicKey:  []byte("public_key"),
			privateKey: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyPair, err := crypto.NewKeyPair(tt.algorithm, tt.publicKey, tt.privateKey)

			if tt.wantErr {
				assert.Error(t, err, "expected error but got nil")
				assert.Nil(t, keyPair, "keyPair should be nil on error")
			} else {
				require.NoError(t, err, "unexpected error creating keyPair")
				require.NotNil(t, keyPair, "keyPair should not be nil")
				assert.Equal(t, tt.algorithm, keyPair.Algorithm)
				assert.Equal(t, tt.publicKey, keyPair.PublicKey)
				assert.Equal(t, tt.privateKey, keyPair.PrivateKey)
				assert.False(t, keyPair.CreatedAt.IsZero(), "CreatedAt should be set")
				assert.WithinDuration(t, time.Now(), keyPair.CreatedAt, 5*time.Second)
			}
		})
	}
}

// TestKeyPair_Destroy tests secure key destruction (SECURITY-CRITICAL)
func TestKeyPair_Destroy(t *testing.T) {
	t.Run("zeroes_private_key_memory", func(t *testing.T) {
		privateKey := []byte("sensitive_private_key_data_12345")
		keyPair, err := crypto.NewKeyPair(
			crypto.Kyber1024,
			[]byte("public_key"),
			privateKey,
		)
		require.NoError(t, err, "setup: failed to create keyPair")

		// Verify private key is not zero before destroy
		allZero := true
		for _, b := range keyPair.PrivateKey {
			if b != 0 {
				allZero = false
				break
			}
		}
		assert.False(t, allZero, "private key should not be all zeros before Destroy")

		// Act: Destroy the key
		keyPair.Destroy()

		// Assert: Verify all bytes are zeroed (SECURITY TEST)
		for i, b := range keyPair.PrivateKey {
			assert.Equal(t, byte(0), b, "byte at index %d not zeroed", i)
		}
	})

	t.Run("idempotent_destroy", func(t *testing.T) {
		keyPair, err := crypto.NewKeyPair(
			crypto.Kyber1024,
			[]byte("public_key"),
			[]byte("private_key"),
		)
		require.NoError(t, err, "setup: failed to create keyPair")

		// Act & Assert: Destroy multiple times should not panic
		assert.NotPanics(t, func() {
			keyPair.Destroy()
			keyPair.Destroy()
			keyPair.Destroy()
		}, "Destroy should be idempotent")
	})

	t.Run("original_slice_is_affected", func(t *testing.T) {
		// This test verifies that Destroy actually zeroes the original slice
		originalKey := []byte("original_sensitive_key")
		keyPair, err := crypto.NewKeyPair(
			crypto.Kyber1024,
			[]byte("public_key"),
			originalKey,
		)
		require.NoError(t, err)

		keyPair.Destroy()

		// The original slice should also be zeroed (they share the same backing array)
		for i, b := range originalKey {
			assert.Equal(t, byte(0), b, "original slice byte at index %d not zeroed", i)
		}
	})
}
