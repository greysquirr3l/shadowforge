package crypto

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

func TestMemoryRepository_SaveAndGetPayload(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	payloadID, _ := domain_crypto.NewPayloadID("test-payload-1")
	payload := &domain_crypto.CryptoPayload{
		ID:          payloadID,
		Data:        []byte("encrypted data"),
		Algorithm:   domain_crypto.Kyber1024,
		EncryptedAt: time.Now(),
	}

	// Act - Save
	err := repo.SavePayload(ctx, payload)
	require.NoError(t, err)

	// Act - Get
	retrieved, err := repo.GetPayload(ctx, payloadID)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, payload.ID, retrieved.ID)
	assert.Equal(t, payload.Data, retrieved.Data)
	assert.Equal(t, payload.Algorithm, retrieved.Algorithm)
	assert.Equal(t, 1, repo.PayloadCount())
}

func TestMemoryRepository_GetPayload_NotFound(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	payloadID, _ := domain_crypto.NewPayloadID("non-existent")

	// Act
	payload, err := repo.GetPayload(ctx, payloadID)

	// Assert
	assert.Nil(t, payload)
	assert.ErrorIs(t, err, ErrPayloadNotFound)
}

func TestMemoryRepository_SavePayload_Nil(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	// Act
	err := repo.SavePayload(ctx, nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestMemoryRepository_DeletePayload(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	payloadID, _ := domain_crypto.NewPayloadID("test-payload-delete")
	payload := &domain_crypto.CryptoPayload{
		ID:          payloadID,
		Data:        []byte("data to delete"),
		Algorithm:   domain_crypto.Kyber1024,
		EncryptedAt: time.Now(),
	}

	// Save first
	err := repo.SavePayload(ctx, payload)
	require.NoError(t, err)
	assert.Equal(t, 1, repo.PayloadCount())

	// Act - Delete
	err = repo.DeletePayload(ctx, payloadID)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, 0, repo.PayloadCount())

	// Verify it's gone
	retrieved, err := repo.GetPayload(ctx, payloadID)
	assert.Nil(t, retrieved)
	assert.ErrorIs(t, err, ErrPayloadNotFound)
}

func TestMemoryRepository_DeletePayload_NotFound(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	payloadID, _ := domain_crypto.NewPayloadID("non-existent")

	// Act
	err := repo.DeletePayload(ctx, payloadID)

	// Assert
	assert.ErrorIs(t, err, ErrPayloadNotFound)
}

func TestMemoryRepository_SaveAndGetKeyPair(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	keyPair := &domain_crypto.KeyPair{
		PublicKey:  []byte("public key data"),
		PrivateKey: []byte("private key data"),
		Algorithm:  domain_crypto.Kyber1024,
		CreatedAt:  time.Now(),
	}

	// Act - Save
	err := repo.SaveKeyPair(ctx, keyPair)
	require.NoError(t, err)

	// Act - Get
	retrieved, err := repo.GetKeyPair(ctx, domain_crypto.Kyber1024)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, keyPair.PublicKey, retrieved.PublicKey)
	assert.Equal(t, keyPair.PrivateKey, retrieved.PrivateKey)
	assert.Equal(t, keyPair.Algorithm, retrieved.Algorithm)
	assert.Equal(t, 1, repo.KeyPairCount())
}

func TestMemoryRepository_GetKeyPair_NotFound(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	// Act
	keyPair, err := repo.GetKeyPair(ctx, domain_crypto.Dilithium3)

	// Assert
	assert.Nil(t, keyPair)
	assert.ErrorIs(t, err, ErrKeyPairNotFound)
}

func TestMemoryRepository_SaveKeyPair_Nil(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	// Act
	err := repo.SaveKeyPair(ctx, nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestMemoryRepository_DeleteKeyPair(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	keyPair := &domain_crypto.KeyPair{
		PublicKey:  []byte("public key"),
		PrivateKey: []byte("private key"),
		Algorithm:  domain_crypto.Kyber1024,
		CreatedAt:  time.Now(),
	}

	// Save first
	err := repo.SaveKeyPair(ctx, keyPair)
	require.NoError(t, err)
	assert.Equal(t, 1, repo.KeyPairCount())

	// Act - Delete
	err = repo.DeleteKeyPair(ctx, domain_crypto.Kyber1024)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, 0, repo.KeyPairCount())

	// Verify it's gone
	retrieved, err := repo.GetKeyPair(ctx, domain_crypto.Kyber1024)
	assert.Nil(t, retrieved)
	assert.ErrorIs(t, err, ErrKeyPairNotFound)
}

func TestMemoryRepository_DeleteKeyPair_NotFound(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	// Act
	err := repo.DeleteKeyPair(ctx, domain_crypto.Dilithium3)

	// Assert
	assert.ErrorIs(t, err, ErrKeyPairNotFound)
}

func TestMemoryRepository_Clear(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	// Add some data
	payloadID, _ := domain_crypto.NewPayloadID("test-payload")
	payload := &domain_crypto.CryptoPayload{
		ID:        payloadID,
		Data:      []byte("data"),
		Algorithm: domain_crypto.Kyber1024,
	}
	repo.SavePayload(ctx, payload)

	keyPair := &domain_crypto.KeyPair{
		PublicKey:  []byte("public"),
		PrivateKey: []byte("private"),
		Algorithm:  domain_crypto.Kyber1024,
	}
	repo.SaveKeyPair(ctx, keyPair)

	assert.Equal(t, 1, repo.PayloadCount())
	assert.Equal(t, 1, repo.KeyPairCount())

	// Act
	repo.Clear()

	// Assert
	assert.Equal(t, 0, repo.PayloadCount())
	assert.Equal(t, 0, repo.KeyPairCount())
}

func TestMemoryRepository_ConcurrentPayloadAccess(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	const goroutines = 10
	done := make(chan bool, goroutines)

	// Act - Concurrent save/get operations
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			payloadID, _ := domain_crypto.NewPayloadID(string(rune('A' + id)))
			payload := &domain_crypto.CryptoPayload{
				ID:        payloadID,
				Data:      []byte{byte(id)},
				Algorithm: domain_crypto.Kyber1024,
			}

			err := repo.SavePayload(ctx, payload)
			assert.NoError(t, err)

			retrieved, err := repo.GetPayload(ctx, payloadID)
			assert.NoError(t, err)
			assert.Equal(t, payload.Data, retrieved.Data)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Assert
	assert.Equal(t, goroutines, repo.PayloadCount())
}

func TestMemoryRepository_ConcurrentKeyPairAccess(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	// Pre-save key pairs for different algorithms
	algorithms := []domain_crypto.PQCAlgorithm{
		domain_crypto.Kyber1024,
		domain_crypto.Dilithium3,
	}

	for _, algo := range algorithms {
		keyPair := &domain_crypto.KeyPair{
			PublicKey:  []byte("public-" + algo.String()),
			PrivateKey: []byte("private-" + algo.String()),
			Algorithm:  algo,
		}
		repo.SaveKeyPair(ctx, keyPair)
	}

	const goroutines = 10
	done := make(chan bool, goroutines)

	// Act - Concurrent read operations
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			algo := algorithms[id%len(algorithms)]

			retrieved, err := repo.GetKeyPair(ctx, algo)
			assert.NoError(t, err)
			assert.Equal(t, algo, retrieved.Algorithm)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestMemoryRepository_UpdatePayload(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	payloadID, _ := domain_crypto.NewPayloadID("update-test")
	payload1 := &domain_crypto.CryptoPayload{
		ID:        payloadID,
		Data:      []byte("original data"),
		Algorithm: domain_crypto.Kyber1024,
	}

	// Save original
	err := repo.SavePayload(ctx, payload1)
	require.NoError(t, err)

	// Act - Update (save with same ID)
	payload2 := &domain_crypto.CryptoPayload{
		ID:        payloadID,
		Data:      []byte("updated data"),
		Algorithm: domain_crypto.Dilithium3,
	}
	err = repo.SavePayload(ctx, payload2)
	require.NoError(t, err)

	// Assert - Should have updated, not created new
	assert.Equal(t, 1, repo.PayloadCount())

	retrieved, err := repo.GetPayload(ctx, payloadID)
	require.NoError(t, err)
	assert.Equal(t, []byte("updated data"), retrieved.Data)
	assert.Equal(t, domain_crypto.Dilithium3, retrieved.Algorithm)
}

func TestMemoryRepository_UpdateKeyPair(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	keyPair1 := &domain_crypto.KeyPair{
		PublicKey:  []byte("original public"),
		PrivateKey: []byte("original private"),
		Algorithm:  domain_crypto.Kyber1024,
	}

	// Save original
	err := repo.SaveKeyPair(ctx, keyPair1)
	require.NoError(t, err)

	// Act - Update (save with same algorithm)
	keyPair2 := &domain_crypto.KeyPair{
		PublicKey:  []byte("updated public"),
		PrivateKey: []byte("updated private"),
		Algorithm:  domain_crypto.Kyber1024,
	}
	err = repo.SaveKeyPair(ctx, keyPair2)
	require.NoError(t, err)

	// Assert - Should have updated, not created new
	assert.Equal(t, 1, repo.KeyPairCount())

	retrieved, err := repo.GetKeyPair(ctx, domain_crypto.Kyber1024)
	require.NoError(t, err)
	assert.Equal(t, []byte("updated public"), retrieved.PublicKey)
	assert.Equal(t, []byte("updated private"), retrieved.PrivateKey)
}

func TestMemoryRepository_MultipleAlgorithms(t *testing.T) {
	// Arrange
	repo := NewMemoryRepository(nil)
	ctx := context.Background()

	kyberKeyPair := &domain_crypto.KeyPair{
		PublicKey:  []byte("kyber public"),
		PrivateKey: []byte("kyber private"),
		Algorithm:  domain_crypto.Kyber1024,
	}

	dilithiumKeyPair := &domain_crypto.KeyPair{
		PublicKey:  []byte("dilithium public"),
		PrivateKey: []byte("dilithium private"),
		Algorithm:  domain_crypto.Dilithium3,
	}

	// Act - Save both
	err := repo.SaveKeyPair(ctx, kyberKeyPair)
	require.NoError(t, err)

	err = repo.SaveKeyPair(ctx, dilithiumKeyPair)
	require.NoError(t, err)

	// Assert - Both stored
	assert.Equal(t, 2, repo.KeyPairCount())

	// Retrieve Kyber
	kyberRetrieved, err := repo.GetKeyPair(ctx, domain_crypto.Kyber1024)
	require.NoError(t, err)
	assert.Equal(t, kyberKeyPair.PublicKey, kyberRetrieved.PublicKey)

	// Retrieve Dilithium
	dilithiumRetrieved, err := repo.GetKeyPair(ctx, domain_crypto.Dilithium3)
	require.NoError(t, err)
	assert.Equal(t, dilithiumKeyPair.PublicKey, dilithiumRetrieved.PublicKey)

	// Delete one, other remains
	err = repo.DeleteKeyPair(ctx, domain_crypto.Kyber1024)
	require.NoError(t, err)
	assert.Equal(t, 1, repo.KeyPairCount())

	// Dilithium still there
	dilithiumRetrieved, err = repo.GetKeyPair(ctx, domain_crypto.Dilithium3)
	require.NoError(t, err)
	assert.NotNil(t, dilithiumRetrieved)
}
