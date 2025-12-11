package crypto

import (
	"context"
	"crypto"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"

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
)

// CirclCryptoService implements cryptographic operations using Cloudflare's CIRCL library
// for post-quantum cryptography (Kyber-1024 for KEM, Dilithium3 for signatures).
type CirclCryptoService struct {
	logger *slog.Logger
}

// NewCirclCryptoService creates a new crypto service using CIRCL library.
func NewCirclCryptoService(logger *slog.Logger) *CirclCryptoService {
	if logger == nil {
		logger = slog.Default()
	}

	return &CirclCryptoService{
		logger: logger,
	}
}

// GenerateKyberKeyPair generates a Kyber-1024 KEM key pair.
// Returns public key (1568 bytes) and private key (3168 bytes).
func (s *CirclCryptoService) GenerateKyberKeyPair(ctx context.Context) ([]byte, []byte, error) {
	s.logger.InfoContext(ctx, "Generating Kyber-1024 key pair")

	// Generate Kyber-1024 key pair
	publicKey, privateKey, err := kyber1024.GenerateKeyPair(rand.Reader)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate Kyber key pair",
			slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidKeyPair, err)
	}

	// Pack keys to bytes
	publicKeyBytes, err := publicKey.MarshalBinary()
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to marshal Kyber public key",
			slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidKeyPair, err)
	}

	privateKeyBytes, err := privateKey.MarshalBinary()
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to marshal Kyber private key",
			slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidKeyPair, err)
	}

	s.logger.InfoContext(ctx, "Kyber-1024 key pair generated successfully",
		slog.Int("public_key_size", len(publicKeyBytes)),
		slog.Int("private_key_size", len(privateKeyBytes)))

	return publicKeyBytes, privateKeyBytes, nil
}

// GenerateDilithiumKeyPair generates a Dilithium3 signature key pair.
// Returns public key (1952 bytes) and private key (4000 bytes).
func (s *CirclCryptoService) GenerateDilithiumKeyPair(ctx context.Context) ([]byte, []byte, error) {
	s.logger.InfoContext(ctx, "Generating Dilithium3 key pair")

	// Generate Dilithium3 key pair
	publicKey, privateKey, err := mode3.GenerateKey(rand.Reader)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate Dilithium key pair",
			slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidKeyPair, err)
	}

	// Pack keys to bytes
	publicKeyBytes, err := publicKey.MarshalBinary()
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to marshal Dilithium public key",
			slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidKeyPair, err)
	}

	privateKeyBytes, err := privateKey.MarshalBinary()
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to marshal Dilithium private key",
			slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidKeyPair, err)
	}

	s.logger.InfoContext(ctx, "Dilithium3 key pair generated successfully",
		slog.Int("public_key_size", len(publicKeyBytes)),
		slog.Int("private_key_size", len(privateKeyBytes)))

	return publicKeyBytes, privateKeyBytes, nil
}

// EncapsulateKyber performs Kyber-1024 encapsulation to generate a shared secret.
// Returns ciphertext (1568 bytes) and shared secret (32 bytes).
func (s *CirclCryptoService) EncapsulateKyber(ctx context.Context, publicKeyBytes []byte) ([]byte, []byte, error) {
	s.logger.InfoContext(ctx, "Performing Kyber-1024 encapsulation",
		slog.Int("public_key_size", len(publicKeyBytes)))

	// Unmarshal public key
	var publicKey kyber1024.PublicKey
	publicKey.Unpack(publicKeyBytes)

	// Perform encapsulation
	ciphertext := make([]byte, kyber1024.CiphertextSize)
	sharedSecret := make([]byte, kyber1024.SharedKeySize)
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate random seed",
			slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}
	publicKey.EncapsulateTo(ciphertext, sharedSecret, seed)

	s.logger.InfoContext(ctx, "Kyber-1024 encapsulation successful",
		slog.Int("ciphertext_size", len(ciphertext)),
		slog.Int("shared_secret_size", len(sharedSecret)))

	return ciphertext, sharedSecret, nil
}

// DecapsulateKyber performs Kyber-1024 decapsulation to recover the shared secret.
// Returns shared secret (32 bytes).
func (s *CirclCryptoService) DecapsulateKyber(ctx context.Context, privateKeyBytes, ciphertext []byte) ([]byte, error) {
	s.logger.InfoContext(ctx, "Performing Kyber-1024 decapsulation",
		slog.Int("private_key_size", len(privateKeyBytes)),
		slog.Int("ciphertext_size", len(ciphertext)))

	// Unmarshal private key
	var privateKey kyber1024.PrivateKey
	privateKey.Unpack(privateKeyBytes)

	// Perform decapsulation
	sharedSecret := make([]byte, kyber1024.SharedKeySize)
	privateKey.DecapsulateTo(sharedSecret, ciphertext)

	s.logger.InfoContext(ctx, "Kyber-1024 decapsulation successful",
		slog.Int("shared_secret_size", len(sharedSecret)))

	return sharedSecret, nil
}

// SignDilithium creates a Dilithium3 signature for the given message.
// Returns signature (approximately 3293 bytes).
func (s *CirclCryptoService) SignDilithium(ctx context.Context, privateKeyBytes, message []byte) ([]byte, error) {
	s.logger.InfoContext(ctx, "Creating Dilithium3 signature",
		slog.Int("private_key_size", len(privateKeyBytes)),
		slog.Int("message_size", len(message)))

	// Unpack private key
	if len(privateKeyBytes) != mode3.PrivateKeySize {
		s.logger.ErrorContext(ctx, "Invalid Dilithium private key size",
			slog.Int("expected", mode3.PrivateKeySize),
			slog.Int("actual", len(privateKeyBytes)))
		return nil, fmt.Errorf("%w: invalid private key size", ErrSigningFailed)
	}
	var privateKey mode3.PrivateKey
	var privateKeyArr [mode3.PrivateKeySize]byte
	copy(privateKeyArr[:], privateKeyBytes)
	privateKey.Unpack(&privateKeyArr)

	// Sign message using crypto.Hash(0) as opts (required by Dilithium)
	signature, err := privateKey.Sign(rand.Reader, message, crypto.Hash(0))
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to sign message",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", ErrSigningFailed, err)
	}

	s.logger.InfoContext(ctx, "Dilithium3 signature created successfully",
		slog.Int("signature_size", len(signature)))

	return signature, nil
}

// VerifyDilithium verifies a Dilithium3 signature.
// Returns true if signature is valid, false otherwise.
func (s *CirclCryptoService) VerifyDilithium(ctx context.Context, publicKeyBytes, message, signature []byte) (bool, error) {
	s.logger.InfoContext(ctx, "Verifying Dilithium3 signature",
		slog.Int("public_key_size", len(publicKeyBytes)),
		slog.Int("message_size", len(message)),
		slog.Int("signature_size", len(signature)))

	// Unpack public key
	if len(publicKeyBytes) != mode3.PublicKeySize {
		s.logger.ErrorContext(ctx, "Invalid Dilithium public key size",
			slog.Int("expected", mode3.PublicKeySize),
			slog.Int("actual", len(publicKeyBytes)))
		return false, fmt.Errorf("%w: invalid public key size", ErrVerificationFailed)
	}
	var publicKey mode3.PublicKey
	var publicKeyArr [mode3.PublicKeySize]byte
	copy(publicKeyArr[:], publicKeyBytes)
	publicKey.Unpack(&publicKeyArr)

	// Verify signature
	valid := mode3.Verify(&publicKey, message, signature)

	s.logger.InfoContext(ctx, "Dilithium3 signature verification completed",
		slog.Bool("valid", valid))

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
