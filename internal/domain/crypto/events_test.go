package crypto_test

import (
	"testing"
	"time"

	"github.com/greysquirr3l/shadowforge/internal/domain/crypto"
	"github.com/stretchr/testify/assert"
)

func TestPayloadEncrypted_Event(t *testing.T) {
	// Arrange
	now := time.Now()
	payloadID, _ := crypto.NewPayloadID("test-payload-123")
	event := crypto.PayloadEncrypted{
		PayloadID:   payloadID,
		Algorithm:   crypto.Kyber1024,
		EncryptedAt: now,
	}

	// Act & Assert - OccurredAt
	assert.Equal(t, now, event.OccurredAt())

	// Act & Assert - EventType
	assert.Equal(t, "crypto.payload_encrypted", event.EventType())

	// Assert implements DomainEvent interface
	var _ crypto.DomainEvent = event
}

func TestPayloadDecrypted_Event(t *testing.T) {
	// Arrange
	now := time.Now()
	payloadID, _ := crypto.NewPayloadID("test-payload-456")
	event := crypto.PayloadDecrypted{
		PayloadID:   payloadID,
		DecryptedAt: now,
	}

	// Act & Assert - OccurredAt
	assert.Equal(t, now, event.OccurredAt())

	// Act & Assert - EventType
	assert.Equal(t, "crypto.payload_decrypted", event.EventType())

	// Assert implements DomainEvent interface
	var _ crypto.DomainEvent = event
}

func TestPayloadSigned_Event(t *testing.T) {
	// Arrange
	now := time.Now()
	payloadID, _ := crypto.NewPayloadID("test-payload-789")
	event := crypto.PayloadSigned{
		PayloadID: payloadID,
		Algorithm: crypto.Dilithium3,
		SignedAt:  now,
	}

	// Act & Assert - OccurredAt
	assert.Equal(t, now, event.OccurredAt())

	// Act & Assert - EventType
	assert.Equal(t, "crypto.payload_signed", event.EventType())

	// Assert implements DomainEvent interface
	var _ crypto.DomainEvent = event
}

func TestSignatureVerified_Event(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"valid signature", true},
		{"invalid signature", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			now := time.Now()
			payloadID, _ := crypto.NewPayloadID("test-payload-sig")
			event := crypto.SignatureVerified{
				PayloadID:  payloadID,
				Valid:      tt.valid,
				VerifiedAt: now,
			}

			// Act & Assert - OccurredAt
			assert.Equal(t, now, event.OccurredAt())

			// Act & Assert - EventType
			assert.Equal(t, "crypto.signature_verified", event.EventType())

			// Assert - Valid field
			assert.Equal(t, tt.valid, event.Valid)

			// Assert implements DomainEvent interface
			var _ crypto.DomainEvent = event
		})
	}
}

func TestKeyPairGenerated_Event(t *testing.T) {
	tests := []struct {
		name      string
		algorithm crypto.PQCAlgorithm
	}{
		{"Kyber1024", crypto.Kyber1024},
		{"Dilithium3", crypto.Dilithium3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			now := time.Now()
			event := crypto.KeyPairGenerated{
				Algorithm:   tt.algorithm,
				GeneratedAt: now,
			}

			// Act & Assert - OccurredAt
			assert.Equal(t, now, event.OccurredAt())

			// Act & Assert - EventType
			assert.Equal(t, "crypto.keypair_generated", event.EventType())

			// Assert - Algorithm field
			assert.Equal(t, tt.algorithm, event.Algorithm)

			// Assert implements DomainEvent interface
			var _ crypto.DomainEvent = event
		})
	}
}

func TestDomainEvents_Timestamps(t *testing.T) {
	// Arrange
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name  string
		event crypto.DomainEvent
		time  time.Time
	}{
		{
			"PayloadEncrypted",
			crypto.PayloadEncrypted{EncryptedAt: past},
			past,
		},
		{
			"PayloadDecrypted",
			crypto.PayloadDecrypted{DecryptedAt: future},
			future,
		},
		{
			"PayloadSigned",
			crypto.PayloadSigned{SignedAt: past},
			past,
		},
		{
			"SignatureVerified",
			crypto.SignatureVerified{VerifiedAt: future},
			future,
		},
		{
			"KeyPairGenerated",
			crypto.KeyPairGenerated{GeneratedAt: past},
			past,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			occurredAt := tt.event.OccurredAt()

			// Assert
			assert.Equal(t, tt.time, occurredAt,
				"Event %s should return correct timestamp", tt.name)
		})
	}
}

func TestDomainEvents_TypeStrings(t *testing.T) {
	tests := []struct {
		event        crypto.DomainEvent
		expectedType string
	}{
		{crypto.PayloadEncrypted{}, "crypto.payload_encrypted"},
		{crypto.PayloadDecrypted{}, "crypto.payload_decrypted"},
		{crypto.PayloadSigned{}, "crypto.payload_signed"},
		{crypto.SignatureVerified{}, "crypto.signature_verified"},
		{crypto.KeyPairGenerated{}, "crypto.keypair_generated"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedType, func(t *testing.T) {
			// Act
			eventType := tt.event.EventType()

			// Assert
			assert.Equal(t, tt.expectedType, eventType,
				"Event should return correct type string")
		})
	}
}
