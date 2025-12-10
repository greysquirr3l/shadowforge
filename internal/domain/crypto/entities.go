// Package crypto implements the Cryptography bounded context for Shadowforge.
// It handles post-quantum cryptographic operations using Kyber-1024 and Dilithium3.
package crypto

import (
	"time"
)

// CryptoPayload is the root aggregate for encrypted data in the Cryptography context.
// It represents a complete encrypted payload with metadata and cryptographic bindings.
type CryptoPayload struct {
	ID          PayloadID
	Data        []byte
	Algorithm   PQCAlgorithm
	Nonce       []byte
	Signature   []byte
	PublicKey   []byte
	EncryptedAt time.Time
	ExpiresAt   *time.Time
}

// NewCryptoPayload creates a new CryptoPayload with validation.
func NewCryptoPayload(id PayloadID, data []byte, algorithm PQCAlgorithm) (*CryptoPayload, error) {
	if len(data) == 0 {
		return nil, ErrEmptyPayload
	}

	if !algorithm.IsValid() {
		return nil, ErrInvalidAlgorithm
	}

	return &CryptoPayload{
		ID:          id,
		Data:        data,
		Algorithm:   algorithm,
		EncryptedAt: time.Now(),
	}, nil
}

// SetSignature attaches a digital signature to the payload.
func (p *CryptoPayload) SetSignature(signature, publicKey []byte) error {
	if len(signature) == 0 {
		return ErrEmptySignature
	}
	if len(publicKey) == 0 {
		return ErrEmptyPublicKey
	}

	p.Signature = signature
	p.PublicKey = publicKey
	return nil
}

// SetExpiration sets when this payload should expire.
func (p *CryptoPayload) SetExpiration(expiresAt time.Time) error {
	if expiresAt.Before(p.EncryptedAt) {
		return ErrInvalidExpiration
	}

	p.ExpiresAt = &expiresAt
	return nil
}

// IsExpired checks if the payload has expired.
func (p *CryptoPayload) IsExpired() bool {
	if p.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*p.ExpiresAt)
}

// Validate checks the integrity of the CryptoPayload.
func (p *CryptoPayload) Validate() error {
	if p.ID.IsZero() {
		return ErrInvalidPayloadID
	}

	if len(p.Data) == 0 {
		return ErrEmptyPayload
	}

	if !p.Algorithm.IsValid() {
		return ErrInvalidAlgorithm
	}

	if p.IsExpired() {
		return ErrPayloadExpired
	}

	return nil
}

// KeyPair represents a post-quantum cryptographic key pair.
// KeyPair represents a post-quantum cryptographic key pair.
type KeyPair struct {
	ID         string
	Algorithm  PQCAlgorithm
	PublicKey  []byte
	PrivateKey []byte
	CreatedAt  time.Time
	ExpiresAt  *time.Time
}

// NewKeyPair creates a new KeyPair with validation.
func NewKeyPair(algorithm PQCAlgorithm, publicKey, privateKey []byte) (*KeyPair, error) {
	if !algorithm.IsValid() {
		return nil, ErrInvalidAlgorithm
	}

	if len(publicKey) == 0 || len(privateKey) == 0 {
		return nil, ErrEmptyKey
	}

	return &KeyPair{
		ID:         GenerateKeyPairID(),
		Algorithm:  algorithm,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		CreatedAt:  time.Now(),
	}, nil
}

// Destroy securely zeros the private key material.
// SECURITY: This MUST be called when the key pair is no longer needed.
func (k *KeyPair) Destroy() {
	for i := range k.PrivateKey {
		k.PrivateKey[i] = 0
	}
	// Prevent compiler optimization
	_ = k.PrivateKey
}
