// Package crypto provides data transfer objects for cryptographic operations.
package crypto

import (
	"time"

	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

// KeyInfoRequest represents a request to get information about a cryptographic key.
type KeyInfoRequest struct {
	// KeyID is the unique identifier of the key
	KeyID string

	// Algorithm is the cryptographic algorithm of the key
	Algorithm domain_crypto.PQCAlgorithm
}

// KeyInfoResponse represents information about a cryptographic key.
type KeyInfoResponse struct {
	// KeyID is the unique identifier of the key
	KeyID string

	// Algorithm is the cryptographic algorithm
	Algorithm domain_crypto.PQCAlgorithm

	// PublicKey is the public key bytes
	PublicKey []byte

	// CreatedAt is when the key was generated
	CreatedAt time.Time

	// ExpiresAt is when the key expires (optional)
	ExpiresAt *time.Time
}

// ListKeysRequest represents a request to list cryptographic keys.
type ListKeysRequest struct {
	// Algorithm filters keys by algorithm (optional)
	Algorithm *domain_crypto.PQCAlgorithm

	// Limit is the maximum number of keys to return
	Limit int

	// Offset is the number of keys to skip
	Offset int
}

// KeySummary represents a summary of a cryptographic key.
type KeySummary struct {
	// KeyID is the unique identifier
	KeyID string

	// Algorithm is the cryptographic algorithm
	Algorithm domain_crypto.PQCAlgorithm

	// CreatedAt is when the key was generated
	CreatedAt time.Time

	// ExpiresAt is when the key expires (optional)
	ExpiresAt *time.Time

	// IsExpired indicates if the key has expired
	IsExpired bool
}

// ListKeysResponse represents a list of cryptographic keys.
type ListKeysResponse struct {
	// Keys is the list of key summaries
	Keys []KeySummary

	// Count is the total number of keys matching the filter
	Count int

	// Limit is the maximum number of keys returned
	Limit int

	// Offset is the number of keys skipped
	Offset int
}

// GetCapacityRequest represents a request to get encryption capacity.
type GetCapacityRequest struct {
	// MediaSize is the size of the media in bytes
	MediaSize int64

	// Algorithm is the cryptographic algorithm to use
	Algorithm domain_crypto.PQCAlgorithm
}

// GetCapacityResponse represents encryption capacity information.
type GetCapacityResponse struct {
	// MaxPayloadSize is the maximum payload size that can be encrypted
	MaxPayloadSize int64

	// Algorithm is the cryptographic algorithm
	Algorithm domain_crypto.PQCAlgorithm

	// OverheadBytes is the encryption overhead in bytes
	OverheadBytes int64

	// CalculatedAt is when the capacity was calculated
	CalculatedAt time.Time
}
