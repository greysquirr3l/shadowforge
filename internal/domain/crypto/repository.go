package crypto

import (
	"context"
)

// CryptoRepository defines the persistence interface for crypto domain objects.
// Implementation is in the infrastructure layer.
type CryptoRepository interface {
	// SavePayload persists an encrypted payload.
	SavePayload(ctx context.Context, payload *CryptoPayload) error

	// GetPayload retrieves an encrypted payload by ID.
	GetPayload(ctx context.Context, id PayloadID) (*CryptoPayload, error)

	// DeletePayload removes an encrypted payload.
	DeletePayload(ctx context.Context, id PayloadID) error

	// SaveKeyPair persists a cryptographic key pair.
	SaveKeyPair(ctx context.Context, keyPair *KeyPair) error

	// GetKeyPair retrieves a key pair by algorithm type.
	GetKeyPair(ctx context.Context, algorithm PQCAlgorithm) (*KeyPair, error)

	// DeleteKeyPair removes a key pair.
	DeleteKeyPair(ctx context.Context, algorithm PQCAlgorithm) error
}
