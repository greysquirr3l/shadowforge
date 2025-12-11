package crypto

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCirclCryptoService_GenerateKyberKeyPair(t *testing.T) {
	// Arrange
	service := NewCirclCryptoService(nil)
	ctx := context.Background()

	// Act
	publicKey, privateKey, err := service.GenerateKyberKeyPair(ctx)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, publicKey)
	assert.NotNil(t, privateKey)
	assert.Equal(t, 1568, len(publicKey), "Kyber-1024 public key should be 1568 bytes")
	assert.Equal(t, 3168, len(privateKey), "Kyber-1024 private key should be 3168 bytes")
	assert.NotEqual(t, make([]byte, len(publicKey)), publicKey, "Public key should not be all zeros")
	assert.NotEqual(t, make([]byte, len(privateKey)), privateKey, "Private key should not be all zeros")
}

func TestCirclCryptoService_GenerateDilithiumKeyPair(t *testing.T) {
	// Arrange
	service := NewCirclCryptoService(nil)
	ctx := context.Background()

	// Act
	publicKey, privateKey, err := service.GenerateDilithiumKeyPair(ctx)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, publicKey)
	assert.NotNil(t, privateKey)
	assert.Equal(t, 1952, len(publicKey), "Dilithium3 public key should be 1952 bytes")
	assert.Equal(t, 4000, len(privateKey), "Dilithium3 private key should be 4000 bytes")
	assert.NotEqual(t, make([]byte, len(publicKey)), publicKey)
	assert.NotEqual(t, make([]byte, len(privateKey)), privateKey)
}

func TestCirclCryptoService_KyberEncapsulation_RoundTrip(t *testing.T) {
	// Arrange
	service := NewCirclCryptoService(nil)
	ctx := context.Background()

	publicKey, privateKey, err := service.GenerateKyberKeyPair(ctx)
	require.NoError(t, err)

	// Act - Encapsulate
	ciphertext, sharedSecret1, err := service.EncapsulateKyber(ctx, publicKey)
	require.NoError(t, err)
	assert.NotNil(t, ciphertext)
	assert.NotNil(t, sharedSecret1)
	assert.Equal(t, 1568, len(ciphertext), "Kyber-1024 ciphertext should be 1568 bytes")
	assert.Equal(t, 32, len(sharedSecret1), "Shared secret should be 32 bytes")

	// Act - Decapsulate
	sharedSecret2, err := service.DecapsulateKyber(ctx, privateKey, ciphertext)
	require.NoError(t, err)
	assert.NotNil(t, sharedSecret2)

	// Assert - Shared secrets match
	assert.Equal(t, sharedSecret1, sharedSecret2, "Shared secrets should match")
	assert.NotEqual(t, make([]byte, 32), sharedSecret1, "Shared secret should not be all zeros")
}

func TestCirclCryptoService_KyberEncapsulation_MultipleIterations(t *testing.T) {
	// Arrange
	service := NewCirclCryptoService(nil)
	ctx := context.Background()

	publicKey, privateKey, err := service.GenerateKyberKeyPair(ctx)
	require.NoError(t, err)

	// Act - Perform multiple encapsulations
	const iterations = 5
	for i := 0; i < iterations; i++ {
		ct, ss1, err := service.EncapsulateKyber(ctx, publicKey)
		require.NoError(t, err, "Iteration %d encapsulation failed", i)

		ss2, err := service.DecapsulateKyber(ctx, privateKey, ct)
		require.NoError(t, err, "Iteration %d decapsulation failed", i)

		assert.Equal(t, ss1, ss2, "Iteration %d shared secrets don't match", i)
	}
}

func TestCirclCryptoService_DilithiumSignature_RoundTrip(t *testing.T) {
	// Arrange
	service := NewCirclCryptoService(nil)
	ctx := context.Background()

	publicKey, privateKey, err := service.GenerateDilithiumKeyPair(ctx)
	require.NoError(t, err)

	message := []byte("test message for quantum-resistant signature")

	// Act - Sign
	signature, err := service.SignDilithium(ctx, privateKey, message)
	require.NoError(t, err)
	assert.NotNil(t, signature)
	assert.True(t, len(signature) >= 3000 && len(signature) <= 4000,
		"Dilithium3 signature should be around 3293 bytes")

	// Act - Verify
	valid, err := service.VerifyDilithium(ctx, publicKey, message, signature)
	require.NoError(t, err)
	assert.True(t, valid, "Signature should verify")
}

func TestCirclCryptoService_DilithiumSignature_TamperedMessage(t *testing.T) {
	// Arrange
	service := NewCirclCryptoService(nil)
	ctx := context.Background()

	publicKey, privateKey, err := service.GenerateDilithiumKeyPair(ctx)
	require.NoError(t, err)

	message := []byte("original message")

	// Act - Sign original message
	signature, err := service.SignDilithium(ctx, privateKey, message)
	require.NoError(t, err)

	// Act - Verify with tampered message
	tamperedMessage := []byte("tampered message")
	valid, err := service.VerifyDilithium(ctx, publicKey, tamperedMessage, signature)
	require.NoError(t, err)

	// Assert - Should not verify
	assert.False(t, valid, "Tampered message should not verify")
}

func TestCirclCryptoService_DilithiumSignature_CorruptedSignature(t *testing.T) {
	// Arrange
	service := NewCirclCryptoService(nil)
	ctx := context.Background()

	publicKey, privateKey, err := service.GenerateDilithiumKeyPair(ctx)
	require.NoError(t, err)

	message := []byte("test message")

	// Act - Sign
	signature, err := service.SignDilithium(ctx, privateKey, message)
	require.NoError(t, err)

	// Corrupt signature
	corruptedSig := make([]byte, len(signature))
	copy(corruptedSig, signature)
	corruptedSig[0] ^= 0xFF // Flip bits

	// Act - Verify corrupted signature
	valid, err := service.VerifyDilithium(ctx, publicKey, message, corruptedSig)
	require.NoError(t, err)

	// Assert
	assert.False(t, valid, "Corrupted signature should not verify")
}

func TestCirclCryptoService_DilithiumSignature_EmptyMessage(t *testing.T) {
	// Arrange
	service := NewCirclCryptoService(nil)
	ctx := context.Background()

	publicKey, privateKey, err := service.GenerateDilithiumKeyPair(ctx)
	require.NoError(t, err)

	message := []byte{}

	// Act - Sign empty message
	signature, err := service.SignDilithium(ctx, privateKey, message)
	require.NoError(t, err)

	// Act - Verify
	valid, err := service.VerifyDilithium(ctx, publicKey, message, signature)
	require.NoError(t, err)

	// Assert
	assert.True(t, valid, "Empty message signature should verify")
}

func TestCirclCryptoService_DilithiumSignature_LargeMessage(t *testing.T) {
	// Arrange
	service := NewCirclCryptoService(nil)
	ctx := context.Background()

	publicKey, privateKey, err := service.GenerateDilithiumKeyPair(ctx)
	require.NoError(t, err)

	// Create large message (1MB)
	largeMessage := make([]byte, 1024*1024)
	_, err = rand.Read(largeMessage)
	require.NoError(t, err)

	// Act - Sign large message
	signature, err := service.SignDilithium(ctx, privateKey, largeMessage)
	require.NoError(t, err)

	// Verify
	valid, err := service.VerifyDilithium(ctx, publicKey, largeMessage, signature)
	require.NoError(t, err)
	assert.True(t, valid, "Large message signature should verify")
}

func TestCirclCryptoService_ConcurrentKeyGeneration(t *testing.T) {
	// Arrange
	service := NewCirclCryptoService(nil)
	ctx := context.Background()

	const goroutines = 10
	done := make(chan bool, goroutines)

	// Act - Generate keys concurrently
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Kyber keys
			kPub, kPriv, err := service.GenerateKyberKeyPair(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, kPub)
			assert.NotNil(t, kPriv)

			// Dilithium keys
			dPub, dPriv, err := service.GenerateDilithiumKeyPair(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, dPub)
			assert.NotNil(t, dPriv)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestCirclCryptoService_ConcurrentEncapsulation(t *testing.T) {
	// Arrange
	service := NewCirclCryptoService(nil)
	ctx := context.Background()

	publicKey, privateKey, err := service.GenerateKyberKeyPair(ctx)
	require.NoError(t, err)

	const goroutines = 10
	done := make(chan bool, goroutines)

	// Act - Encapsulate/decapsulate concurrently
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			ct, ss1, err := service.EncapsulateKyber(ctx, publicKey)
			assert.NoError(t, err)

			ss2, err := service.DecapsulateKyber(ctx, privateKey, ct)
			assert.NoError(t, err)

			assert.Equal(t, ss1, ss2)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestCirclCryptoService_KeyIndependence(t *testing.T) {
	// Arrange
	service := NewCirclCryptoService(nil)
	ctx := context.Background()

	// Generate two key pairs
	pub1, priv1, err := service.GenerateKyberKeyPair(ctx)
	require.NoError(t, err)

	pub2, priv2, err := service.GenerateKyberKeyPair(ctx)
	require.NoError(t, err)

	// Assert - Keys are different
	assert.NotEqual(t, pub1, pub2, "Public keys should be different")
	assert.NotEqual(t, priv1, priv2, "Private keys should be different")

	// Act - Encapsulate with first public key
	ct, ss1, err := service.EncapsulateKyber(ctx, pub1)
	require.NoError(t, err)

	// Decapsulate with correct private key
	ss2, err := service.DecapsulateKyber(ctx, priv1, ct)
	require.NoError(t, err)
	assert.Equal(t, ss1, ss2, "Correct key pair should recover shared secret")

	// Decapsulate with wrong private key
	ss3, err := service.DecapsulateKyber(ctx, priv2, ct)
	if err == nil {
		assert.NotEqual(t, ss1, ss3, "Wrong private key should produce different shared secret")
	}
}
