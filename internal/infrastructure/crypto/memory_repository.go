package crypto

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

var (
	// ErrPayloadNotFound is returned when a payload is not found in the repository.
	ErrPayloadNotFound = errors.New("payload not found")

	// ErrKeyPairNotFound is returned when a key pair is not found in the repository.
	ErrKeyPairNotFound = errors.New("key pair not found")
)

// MemoryRepository is an in-memory implementation of CryptoRepository.
// Thread-safe using sync.RWMutex for concurrent access.
type MemoryRepository struct {
	payloads map[string]*domain_crypto.CryptoPayload
	keyPairs map[domain_crypto.PQCAlgorithm]*domain_crypto.KeyPair
	mu       sync.RWMutex
	logger   *slog.Logger
}

// NewMemoryRepository creates a new in-memory repository.
func NewMemoryRepository(logger *slog.Logger) *MemoryRepository {
	if logger == nil {
		logger = slog.Default()
	}

	return &MemoryRepository{
		payloads: make(map[string]*domain_crypto.CryptoPayload),
		keyPairs: make(map[domain_crypto.PQCAlgorithm]*domain_crypto.KeyPair),
		logger:   logger,
	}
}

// SavePayload persists an encrypted payload.
func (r *MemoryRepository) SavePayload(ctx context.Context, payload *domain_crypto.CryptoPayload) error {
	if payload == nil {
		return errors.New("payload cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.payloads[payload.ID.String()] = payload

	r.logger.InfoContext(ctx, "Payload saved to memory repository",
		slog.String("payload_id", payload.ID.String()),
		slog.String("algorithm", payload.Algorithm.String()))

	return nil
}

// GetPayload retrieves an encrypted payload by ID.
func (r *MemoryRepository) GetPayload(ctx context.Context, id domain_crypto.PayloadID) (*domain_crypto.CryptoPayload, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	payload, exists := r.payloads[id.String()]
	if !exists {
		r.logger.WarnContext(ctx, "Payload not found in repository",
			slog.String("payload_id", id.String()))
		return nil, fmt.Errorf("%w: %s", ErrPayloadNotFound, id.String())
	}

	r.logger.InfoContext(ctx, "Payload retrieved from memory repository",
		slog.String("payload_id", id.String()))

	return payload, nil
}

// DeletePayload removes an encrypted payload.
func (r *MemoryRepository) DeletePayload(ctx context.Context, id domain_crypto.PayloadID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.payloads[id.String()]; !exists {
		r.logger.WarnContext(ctx, "Attempted to delete non-existent payload",
			slog.String("payload_id", id.String()))
		return fmt.Errorf("%w: %s", ErrPayloadNotFound, id.String())
	}

	delete(r.payloads, id.String())

	r.logger.InfoContext(ctx, "Payload deleted from memory repository",
		slog.String("payload_id", id.String()))

	return nil
}

// SaveKeyPair persists a cryptographic key pair.
func (r *MemoryRepository) SaveKeyPair(ctx context.Context, keyPair *domain_crypto.KeyPair) error {
	if keyPair == nil {
		return errors.New("key pair cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.keyPairs[keyPair.Algorithm] = keyPair

	r.logger.InfoContext(ctx, "Key pair saved to memory repository",
		slog.String("algorithm", keyPair.Algorithm.String()))

	return nil
}

// GetKeyPair retrieves a key pair by algorithm type.
func (r *MemoryRepository) GetKeyPair(ctx context.Context, algorithm domain_crypto.PQCAlgorithm) (*domain_crypto.KeyPair, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keyPair, exists := r.keyPairs[algorithm]
	if !exists {
		r.logger.WarnContext(ctx, "Key pair not found in repository",
			slog.String("algorithm", algorithm.String()))
		return nil, fmt.Errorf("%w: %s", ErrKeyPairNotFound, algorithm)
	}

	r.logger.InfoContext(ctx, "Key pair retrieved from memory repository",
		slog.String("algorithm", algorithm.String()))

	return keyPair, nil
}

// DeleteKeyPair removes a key pair.
func (r *MemoryRepository) DeleteKeyPair(ctx context.Context, algorithm domain_crypto.PQCAlgorithm) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.keyPairs[algorithm]; !exists {
		r.logger.WarnContext(ctx, "Attempted to delete non-existent key pair",
			slog.String("algorithm", algorithm.String()))
		return fmt.Errorf("%w: %s", ErrKeyPairNotFound, algorithm)
	}

	delete(r.keyPairs, algorithm)

	r.logger.InfoContext(ctx, "Key pair deleted from memory repository",
		slog.String("algorithm", algorithm.String()))

	return nil
}

// Clear removes all data from the repository (useful for testing).
func (r *MemoryRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.payloads = make(map[string]*domain_crypto.CryptoPayload)
	r.keyPairs = make(map[domain_crypto.PQCAlgorithm]*domain_crypto.KeyPair)

	r.logger.Info("Memory repository cleared")
}

// PayloadCount returns the number of stored payloads (useful for testing).
func (r *MemoryRepository) PayloadCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.payloads)
}

// KeyPairCount returns the number of stored key pairs (useful for testing).
func (r *MemoryRepository) KeyPairCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.keyPairs)
}
