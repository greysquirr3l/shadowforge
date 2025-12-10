package crypto

import (
	"context"
)

// CryptoService defines the domain service interface for cryptographic operations.
// This is the core interface for encryption, decryption, signing, and verification.
type CryptoService interface {
	// GenerateKeyPair creates a new post-quantum cryptographic key pair.
	GenerateKeyPair(ctx context.Context, algorithm PQCAlgorithm) (*KeyPair, error)

	// Encrypt encrypts data using the specified algorithm and public key.
	Encrypt(ctx context.Context, data []byte, publicKey []byte, algorithm PQCAlgorithm) (*CryptoPayload, error)

	// Decrypt decrypts a payload using the corresponding private key.
	Decrypt(ctx context.Context, payload *CryptoPayload, privateKey []byte) ([]byte, error)

	// Sign creates a digital signature for the given data.
	Sign(ctx context.Context, data []byte, privateKey []byte, algorithm PQCAlgorithm) ([]byte, error)

	// Verify verifies a digital signature.
	Verify(ctx context.Context, data []byte, signature []byte, publicKey []byte, algorithm PQCAlgorithm) (bool, error)

	// DeriveKey derives a cryptographic key from a password using Argon2id.
	DeriveKey(ctx context.Context, password string, salt []byte) ([]byte, error)
}
