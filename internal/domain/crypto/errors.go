package crypto

import "errors"

// Domain errors for the Cryptography bounded context.
// These are sentinel errors used to identify expected error conditions.
var (
	// Payload errors
	ErrInvalidPayloadID  = errors.New("invalid payload ID")
	ErrEmptyPayload      = errors.New("payload data cannot be empty")
	ErrPayloadExpired    = errors.New("payload has expired")
	ErrInvalidExpiration = errors.New("expiration time must be after creation time")

	// Algorithm errors
	ErrInvalidAlgorithm  = errors.New("invalid or unsupported PQC algorithm")
	ErrAlgorithmMismatch = errors.New("algorithm mismatch")

	// Key errors
	ErrEmptyKey        = errors.New("key cannot be empty")
	ErrEmptyPublicKey  = errors.New("public key cannot be empty")
	ErrEmptyPrivateKey = errors.New("private key cannot be empty")
	ErrInvalidKeySize  = errors.New("invalid key size for algorithm")

	// Signature errors
	ErrEmptySignature    = errors.New("signature cannot be empty")
	ErrInvalidSignature  = errors.New("signature verification failed")
	ErrSignatureMismatch = errors.New("signature does not match payload")

	// Encryption/Decryption errors
	ErrEncryptionFailed = errors.New("encryption operation failed")
	ErrDecryptionFailed = errors.New("decryption operation failed")
	ErrCorruptedData    = errors.New("data corruption detected")
	ErrInvalidNonce     = errors.New("invalid or missing nonce")

	// Key derivation errors
	ErrKeyDerivationFailed = errors.New("key derivation failed")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrInvalidSalt         = errors.New("invalid salt")
)
