// Package crypto provides DTOs for cryptographic operations.
package crypto

import (
	"time"

	"github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

// EncryptionRequest represents a request to encrypt data.
type EncryptionRequest struct {
	// Data is the plaintext data to encrypt
	Data []byte

	// Algorithm specifies the encryption algorithm to use
	Algorithm crypto.PQCAlgorithm

	// RecipientPublicKey is the public key of the recipient (for asymmetric encryption)
	RecipientPublicKey []byte

	// SenderPrivateKey is the private key of the sender (for signing)
	SenderPrivateKey []byte
}

// EncryptionResponse represents the response from an encryption operation.
type EncryptionResponse struct {
	// EncryptedData is the encrypted ciphertext
	EncryptedData []byte

	// Signature is the digital signature of the encrypted data
	Signature []byte

	// Nonce is the nonce used for encryption (if applicable)
	Nonce []byte

	// Algorithm is the algorithm used
	Algorithm crypto.PQCAlgorithm

	// EncryptedAt is the timestamp of encryption
	EncryptedAt time.Time
}

// DecryptionRequest represents a request to decrypt data.
type DecryptionRequest struct {
	// EncryptedData is the ciphertext to decrypt
	EncryptedData []byte

	// RecipientPrivateKey is the private key of the recipient
	RecipientPrivateKey []byte

	// SenderPublicKey is the public key of the sender (for signature verification)
	SenderPublicKey []byte

	// Signature is the digital signature to verify
	Signature []byte

	// Nonce is the nonce used during encryption (if applicable)
	Nonce []byte

	// Algorithm is the algorithm used for encryption
	Algorithm crypto.PQCAlgorithm
}

// DecryptionResponse represents the response from a decryption operation.
type DecryptionResponse struct {
	// Data is the decrypted plaintext
	Data []byte

	// SignatureValid indicates whether the signature was valid
	SignatureValid bool

	// DecryptedAt is the timestamp of decryption
	DecryptedAt time.Time
}

// KeyPairRequest represents a request to generate a key pair.
type KeyPairRequest struct {
	// Algorithm specifies which algorithm to generate keys for
	Algorithm crypto.PQCAlgorithm

	// KeySize specifies the key size (if applicable)
	KeySize int
}

// KeyPairResponse represents a generated key pair.
type KeyPairResponse struct {
	// PublicKey is the generated public key
	PublicKey []byte

	// PrivateKey is the generated private key
	PrivateKey []byte

	// Algorithm is the algorithm for this key pair
	Algorithm crypto.PQCAlgorithm

	// GeneratedAt is the timestamp of key generation
	GeneratedAt time.Time
}
