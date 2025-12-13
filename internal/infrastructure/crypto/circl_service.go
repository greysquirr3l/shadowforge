package crypto

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/argon2"

	"github.com/cloudflare/circl/kem/kyber/kyber1024"
	"github.com/cloudflare/circl/sign/dilithium/mode3"

	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

var (
	// ErrInvalidKeyPair indicates the key pair is malformed or corrupted
	ErrInvalidKeyPair = errors.New("invalid key pair")

	// ErrEncryptionFailed indicates encryption operation failed
	ErrEncryptionFailed = errors.New("encryption failed")

	// ErrDecryptionFailed indicates decryption operation failed
	ErrDecryptionFailed = errors.New("decryption failed")

	// ErrSigningFailed indicates signature generation failed
	ErrSigningFailed = errors.New("signing failed")

	// ErrVerificationFailed indicates signature verification failed
	ErrVerificationFailed = errors.New("verification failed")

	// ErrKeyDerivationFailed indicates key derivation failed
	ErrKeyDerivationFailed = errors.New("key derivation failed")

	// ErrInvalidNonceSize indicates nonce has incorrect size
	ErrInvalidNonceSize = errors.New("invalid nonce size")
)

// CirclCryptoService implements cryptographic operations using Cloudflare's CIRCL library
// for post-quantum cryptography (Kyber-1024 for KEM, Dilithium3 for signatures).
type CirclCryptoService struct {
	logger *logrus.Logger
}

// NewCirclCryptoService creates a new crypto service using CIRCL library.
func NewCirclCryptoService(logger *logrus.Logger) *CirclCryptoService {
	return &CirclCryptoService{
		logger: logger,
	}
}

// GenerateKyberKeyPair generates a Kyber-1024 KEM key pair.
// Returns public key (1568 bytes) and private key (3168 bytes).
func (s *CirclCryptoService) GenerateKyberKeyPair(ctx context.Context) ([]byte, []byte, error) {
	s.logger.Info("Generating Kyber-1024 key pair")

	// Generate Kyber-1024 key pair
	publicKey, privateKey, err := kyber1024.GenerateKeyPair(rand.Reader)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to generate Kyber key pair")
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidKeyPair, err)
	}

	// Pack keys to bytes
	publicKeyBytes, err := publicKey.MarshalBinary()
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to marshal Kyber public key")
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidKeyPair, err)
	}

	privateKeyBytes, err := privateKey.MarshalBinary()
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to marshal Kyber private key")
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidKeyPair, err)
	}

	s.logger.WithFields(logrus.Fields{
		"public_key_size":  len(publicKeyBytes),
		"private_key_size": len(privateKeyBytes),
	}).Info("Kyber-1024 key pair generated successfully")

	return publicKeyBytes, privateKeyBytes, nil
}

// GenerateDilithiumKeyPair generates a Dilithium3 signature key pair.
// Returns public key (1952 bytes) and private key (4000 bytes).
func (s *CirclCryptoService) GenerateDilithiumKeyPair(ctx context.Context) ([]byte, []byte, error) {
	s.logger.Info("Generating Dilithium3 key pair")

	// Generate Dilithium3 key pair
	publicKey, privateKey, err := mode3.GenerateKey(rand.Reader)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to generate Dilithium key pair")
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidKeyPair, err)
	}

	// Pack keys to bytes
	publicKeyBytes, err := publicKey.MarshalBinary()
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to marshal Dilithium public key")
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidKeyPair, err)
	}

	privateKeyBytes, err := privateKey.MarshalBinary()
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to marshal Dilithium private key")
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidKeyPair, err)
	}

	s.logger.WithFields(logrus.Fields{
		"public_key_size":  len(publicKeyBytes),
		"private_key_size": len(privateKeyBytes),
	}).Info("Dilithium3 key pair generated successfully")

	return publicKeyBytes, privateKeyBytes, nil
}

// EncapsulateKyber performs Kyber-1024 encapsulation to generate a shared secret.
// Returns ciphertext (1568 bytes) and shared secret (32 bytes).
func (s *CirclCryptoService) EncapsulateKyber(ctx context.Context, publicKeyBytes []byte) ([]byte, []byte, error) {
	s.logger.WithFields(logrus.Fields{
		"public_key_size": len(publicKeyBytes),
	}).Info("Performing Kyber-1024 encapsulation")

	// Unmarshal public key
	var publicKey kyber1024.PublicKey
	publicKey.Unpack(publicKeyBytes)

	// Perform encapsulation
	ciphertext := make([]byte, kyber1024.CiphertextSize)
	sharedSecret := make([]byte, kyber1024.SharedKeySize)
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to generate random seed")
		return nil, nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}
	publicKey.EncapsulateTo(ciphertext, sharedSecret, seed)

	s.logger.WithFields(logrus.Fields{
		"ciphertext_size":    len(ciphertext),
		"shared_secret_size": len(sharedSecret),
	}).Info("Kyber-1024 encapsulation successful")

	return ciphertext, sharedSecret, nil
}

// DecapsulateKyber performs Kyber-1024 decapsulation to recover the shared secret.
// Returns shared secret (32 bytes).
func (s *CirclCryptoService) DecapsulateKyber(ctx context.Context, privateKeyBytes, ciphertext []byte) ([]byte, error) {
	s.logger.WithFields(logrus.Fields{
		"private_key_size": len(privateKeyBytes),
		"ciphertext_size":  len(ciphertext),
	}).Info("Performing Kyber-1024 decapsulation")

	// Unmarshal private key
	var privateKey kyber1024.PrivateKey
	privateKey.Unpack(privateKeyBytes)

	// Perform decapsulation
	sharedSecret := make([]byte, kyber1024.SharedKeySize)
	privateKey.DecapsulateTo(sharedSecret, ciphertext)

	s.logger.WithFields(logrus.Fields{
		"shared_secret_size": len(sharedSecret),
	}).Info("Kyber-1024 decapsulation successful")

	return sharedSecret, nil
}

// SignDilithium creates a Dilithium3 signature for the given message.
// Returns signature (approximately 3293 bytes).
func (s *CirclCryptoService) SignDilithium(ctx context.Context, privateKeyBytes, message []byte) ([]byte, error) {
	s.logger.WithFields(logrus.Fields{
		"private_key_size": len(privateKeyBytes),
		"message_size":     len(message),
	}).Info("Creating Dilithium3 signature")

	// Unpack private key
	if len(privateKeyBytes) != mode3.PrivateKeySize {
		s.logger.WithFields(logrus.Fields{
			"expected": mode3.PrivateKeySize,
			"actual":   len(privateKeyBytes),
		}).Error("Invalid Dilithium private key size")
		return nil, fmt.Errorf("%w: invalid private key size", ErrSigningFailed)
	}
	var privateKey mode3.PrivateKey
	var privateKeyArr [mode3.PrivateKeySize]byte
	copy(privateKeyArr[:], privateKeyBytes)
	privateKey.Unpack(&privateKeyArr)

	// Sign message using crypto.Hash(0) as opts (required by Dilithium)
	signature, err := privateKey.Sign(rand.Reader, message, crypto.Hash(0))
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to sign message")
		return nil, fmt.Errorf("%w: %v", ErrSigningFailed, err)
	}

	s.logger.WithFields(logrus.Fields{
		"signature_size": len(signature),
	}).Info("Dilithium3 signature created successfully")

	return signature, nil
}

// VerifyDilithium verifies a Dilithium3 signature.
// Returns true if signature is valid, false otherwise.
func (s *CirclCryptoService) VerifyDilithium(ctx context.Context, publicKeyBytes, message, signature []byte) (bool, error) {
	s.logger.WithFields(logrus.Fields{
		"public_key_size": len(publicKeyBytes),
		"message_size":    len(message),
		"signature_size":  len(signature),
	}).Info("Verifying Dilithium3 signature")

	// Unpack public key
	if len(publicKeyBytes) != mode3.PublicKeySize {
		s.logger.WithFields(logrus.Fields{
			"expected": mode3.PublicKeySize,
			"actual":   len(publicKeyBytes),
		}).Error("Invalid Dilithium public key size")
		return false, fmt.Errorf("%w: invalid public key size", ErrVerificationFailed)
	}
	var publicKey mode3.PublicKey
	var publicKeyArr [mode3.PublicKeySize]byte
	copy(publicKeyArr[:], publicKeyBytes)
	publicKey.Unpack(&publicKeyArr)

	// Verify signature
	valid := mode3.Verify(&publicKey, message, signature)

	s.logger.WithFields(logrus.Fields{
		"valid": valid,
	}).Info("Dilithium3 signature verification completed")

	return valid, nil
}

// GetAlgorithmInfo returns information about supported PQC algorithms.
func (s *CirclCryptoService) GetAlgorithmInfo(algorithm domain_crypto.PQCAlgorithm) (AlgorithmInfo, error) {
	switch algorithm {
	case domain_crypto.Kyber1024:
		return AlgorithmInfo{
			Name:             "Kyber-1024",
			Type:             "KEM",
			PublicKeySize:    kyber1024.PublicKeySize,
			PrivateKeySize:   kyber1024.PrivateKeySize,
			CiphertextSize:   kyber1024.CiphertextSize,
			SharedSecretSize: kyber1024.SharedKeySize,
			SecurityLevel:    256, // bits
			NISTLevel:        5,
		}, nil

	case domain_crypto.Dilithium3:
		return AlgorithmInfo{
			Name:           "Dilithium3",
			Type:           "Signature",
			PublicKeySize:  mode3.PublicKeySize,
			PrivateKeySize: mode3.PrivateKeySize,
			SignatureSize:  mode3.SignatureSize,
			SecurityLevel:  192, // bits
			NISTLevel:      3,
		}, nil

	default:
		return AlgorithmInfo{}, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// AlgorithmInfo contains information about a PQC algorithm.
type AlgorithmInfo struct {
	Name             string
	Type             string // "KEM" or "Signature"
	PublicKeySize    int
	PrivateKeySize   int
	CiphertextSize   int // For KEM algorithms
	SharedSecretSize int // For KEM algorithms
	SignatureSize    int // For signature algorithms
	SecurityLevel    int // Security level in bits
	NISTLevel        int // NIST security level (1-5)
}

// ============================================================================
// DOMAIN INTERFACE IMPLEMENTATION
// These methods implement the domain CryptoService interface using low-level
// primitives defined above as helpers.
// ============================================================================

// GenerateKeyPair generates a post-quantum cryptographic key pair.
// This implements the domain interface by generating both Kyber and Dilithium keys.
func (s *CirclCryptoService) GenerateKeyPair(ctx context.Context, algorithm domain_crypto.PQCAlgorithm) (*domain_crypto.KeyPair, error) {
	s.logger.WithFields(logrus.Fields{
		"algorithm": algorithm.Name(),
	}).Info("Generating key pair")

	var kyberPub, kyberPriv, dilithiumPub, dilithiumPriv []byte
	var err error

	// Generate Kyber key pair for encryption
	kyberPub, kyberPriv, err = s.GenerateKyberKeyPair(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Kyber key pair: %w", err)
	}

	// Generate Dilithium key pair for signatures
	dilithiumPub, dilithiumPriv, err = s.GenerateDilithiumKeyPair(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Dilithium key pair: %w", err)
	}

	// Combine keys (Kyber for encryption, Dilithium for signing)
	// In a real implementation, you might store these separately or use a composite format
	// For now, we'll concatenate Kyber + Dilithium keys
	publicKey := append(kyberPub, dilithiumPub...)
	privateKey := append(kyberPriv, dilithiumPriv...)

	keyPair := &domain_crypto.KeyPair{
		ID:         generateKeyPairID(),
		Algorithm:  algorithm,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		CreatedAt:  time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"key_id":           keyPair.ID,
		"public_key_size":  len(publicKey),
		"private_key_size": len(privateKey),
	}).Info("Key pair generated successfully")

	return keyPair, nil
}

// Encrypt encrypts data using post-quantum KEM (Kyber) + AES-GCM.
func (s *CirclCryptoService) Encrypt(ctx context.Context, data []byte, publicKey []byte, algorithm domain_crypto.PQCAlgorithm) (*domain_crypto.CryptoPayload, error) {
	s.logger.WithFields(logrus.Fields{
		"algorithm": algorithm.Name(),
		"data_size": len(data),
	}).Info("Encrypting payload")

	// Extract Kyber public key (first 1568 bytes)
	if len(publicKey) < kyber1024.PublicKeySize {
		return nil, fmt.Errorf("%w: public key too short", ErrInvalidKeyPair)
	}
	kyberPubKey := publicKey[:kyber1024.PublicKeySize]

	// Perform Kyber encapsulation to get shared secret
	ciphertext, sharedSecret, err := s.EncapsulateKyber(ctx, kyberPubKey)
	if err != nil {
		return nil, fmt.Errorf("Kyber encapsulation failed: %w", err)
	}

	// Use shared secret as AES-256 key (first 32 bytes)
	block, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create cipher: %v", ErrEncryptionFailed, err)
	}

	// Create GCM mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create GCM: %v", ErrEncryptionFailed, err)
	}

	// Generate nonce
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("%w: failed to generate nonce: %v", ErrEncryptionFailed, err)
	}

	// Encrypt data
	encryptedData := aesgcm.Seal(nil, nonce, data, nil)

	// Store Kyber ciphertext + encrypted data
	finalData := append(ciphertext, encryptedData...)

	// Create payload
	payloadID := domain_crypto.GeneratePayloadID()
	payload, err := domain_crypto.NewCryptoPayload(payloadID, finalData, algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to create crypto payload: %w", err)
	}
	payload.Nonce = nonce

	s.logger.WithFields(logrus.Fields{
		"payload_id":     payloadID.String(),
		"encrypted_size": len(finalData),
	}).Info("Payload encrypted successfully")

	return payload, nil
}

// Decrypt decrypts a CryptoPayload using post-quantum KEM (Kyber) + AES-GCM.
func (s *CirclCryptoService) Decrypt(ctx context.Context, payload *domain_crypto.CryptoPayload, privateKey []byte) ([]byte, error) {
	s.logger.WithFields(logrus.Fields{
		"payload_id": payload.ID.String(),
		"algorithm":  payload.Algorithm.Name(),
	}).Info("Decrypting payload")

	// Extract Kyber private key (first 3168 bytes)
	if len(privateKey) < kyber1024.PrivateKeySize {
		return nil, fmt.Errorf("%w: private key too short", ErrInvalidKeyPair)
	}
	kyberPrivKey := privateKey[:kyber1024.PrivateKeySize]

	// Extract Kyber ciphertext and encrypted data
	if len(payload.Data) < kyber1024.CiphertextSize {
		return nil, fmt.Errorf("%w: payload data too short", ErrDecryptionFailed)
	}
	kyberCiphertext := payload.Data[:kyber1024.CiphertextSize]
	encryptedData := payload.Data[kyber1024.CiphertextSize:]

	// Decapsulate to recover shared secret
	sharedSecret, err := s.DecapsulateKyber(ctx, kyberPrivKey, kyberCiphertext)
	if err != nil {
		return nil, fmt.Errorf("Kyber decapsulation failed: %w", err)
	}

	// Create AES cipher with shared secret
	block, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create cipher: %v", ErrDecryptionFailed, err)
	}

	// Create GCM mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create GCM: %v", ErrDecryptionFailed, err)
	}

	// Validate nonce size
	if len(payload.Nonce) != aesgcm.NonceSize() {
		return nil, fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidNonceSize, aesgcm.NonceSize(), len(payload.Nonce))
	}

	// Decrypt data
	plaintext, err := aesgcm.Open(nil, payload.Nonce, encryptedData, nil)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("AES-GCM decryption failed")
		return nil, fmt.Errorf("%w: AES-GCM decryption failed: %v", ErrDecryptionFailed, err)
	}

	s.logger.WithFields(logrus.Fields{
		"payload_id":     payload.ID.String(),
		"plaintext_size": len(plaintext),
	}).Info("Payload decrypted successfully")

	return plaintext, nil
}

// Sign creates a digital signature using Dilithium.
func (s *CirclCryptoService) Sign(ctx context.Context, data []byte, privateKey []byte, algorithm domain_crypto.PQCAlgorithm) ([]byte, error) {
	s.logger.WithFields(logrus.Fields{
		"algorithm": algorithm.Name(),
		"data_size": len(data),
	}).Info("Signing data")

	// Extract Dilithium private key (skip Kyber key)
	if len(privateKey) < kyber1024.PrivateKeySize+mode3.PrivateKeySize {
		return nil, fmt.Errorf("%w: private key too short for Dilithium", ErrInvalidKeyPair)
	}
	dilithiumPrivKey := privateKey[kyber1024.PrivateKeySize : kyber1024.PrivateKeySize+mode3.PrivateKeySize]

	// Sign using Dilithium
	signature, err := s.SignDilithium(ctx, dilithiumPrivKey, data)
	if err != nil {
		return nil, fmt.Errorf("Dilithium signing failed: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"signature_size": len(signature),
	}).Info("Data signed successfully")

	return signature, nil
}

// Verify verifies a digital signature using Dilithium.
func (s *CirclCryptoService) Verify(ctx context.Context, data []byte, signature []byte, publicKey []byte, algorithm domain_crypto.PQCAlgorithm) (bool, error) {
	s.logger.WithFields(logrus.Fields{
		"algorithm":      algorithm.Name(),
		"data_size":      len(data),
		"signature_size": len(signature),
	}).Info("Verifying signature")

	// Extract Dilithium public key (skip Kyber key)
	if len(publicKey) < kyber1024.PublicKeySize+mode3.PublicKeySize {
		return false, fmt.Errorf("%w: public key too short for Dilithium", ErrInvalidKeyPair)
	}
	dilithiumPubKey := publicKey[kyber1024.PublicKeySize : kyber1024.PublicKeySize+mode3.PublicKeySize]

	// Verify using Dilithium
	valid, err := s.VerifyDilithium(ctx, dilithiumPubKey, data, signature)
	if err != nil {
		return false, fmt.Errorf("Dilithium verification failed: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"valid": valid,
	}).Info("Signature verification completed")

	return valid, nil
}

// DeriveKey derives a cryptographic key from a password using Argon2id.
func (s *CirclCryptoService) DeriveKey(ctx context.Context, password string, salt []byte) ([]byte, error) {
	s.logger.WithFields(logrus.Fields{
		"salt_size": len(salt),
	}).Info("Deriving key from password")

	if len(password) == 0 {
		return nil, fmt.Errorf("%w: password cannot be empty", ErrKeyDerivationFailed)
	}

	if len(salt) < 16 {
		return nil, fmt.Errorf("%w: salt must be at least 16 bytes", ErrKeyDerivationFailed)
	}

	// Argon2id parameters (conservative for security)
	const (
		time    = 3         // Number of iterations
		memory  = 64 * 1024 // 64 MB
		threads = 4         // Number of threads
		keyLen  = 32        // 256-bit key
	)

	// Derive key using Argon2id
	key := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	s.logger.WithFields(logrus.Fields{
		"key_size": len(key),
	}).Info("Key derived successfully")

	return key, nil
}

// generateKeyPairID generates a unique ID for a key pair.
func generateKeyPairID() string {
	// Simple timestamp-based ID for now
	// In production, use UUID or similar
	return fmt.Sprintf("keypair-%d", time.Now().UnixNano())
}
