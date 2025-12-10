package crypto

import (
	"time"
)

// DomainEvent represents something that happened in the Crypto domain.
type DomainEvent interface {
	// OccurredAt returns when the event occurred.
	OccurredAt() time.Time

	// EventType returns the type of the event.
	EventType() string
}

// PayloadEncrypted is emitted when a payload is successfully encrypted.
type PayloadEncrypted struct {
	PayloadID   PayloadID
	Algorithm   PQCAlgorithm
	EncryptedAt time.Time
}

func (e PayloadEncrypted) OccurredAt() time.Time {
	return e.EncryptedAt
}

func (e PayloadEncrypted) EventType() string {
	return "crypto.payload_encrypted"
}

// PayloadDecrypted is emitted when a payload is successfully decrypted.
type PayloadDecrypted struct {
	PayloadID   PayloadID
	DecryptedAt time.Time
}

func (e PayloadDecrypted) OccurredAt() time.Time {
	return e.DecryptedAt
}

func (e PayloadDecrypted) EventType() string {
	return "crypto.payload_decrypted"
}

// PayloadSigned is emitted when a payload is digitally signed.
type PayloadSigned struct {
	PayloadID PayloadID
	Algorithm PQCAlgorithm
	SignedAt  time.Time
}

func (e PayloadSigned) OccurredAt() time.Time {
	return e.SignedAt
}

func (e PayloadSigned) EventType() string {
	return "crypto.payload_signed"
}

// SignatureVerified is emitted when a signature is verified.
type SignatureVerified struct {
	PayloadID  PayloadID
	Valid      bool
	VerifiedAt time.Time
}

func (e SignatureVerified) OccurredAt() time.Time {
	return e.VerifiedAt
}

func (e SignatureVerified) EventType() string {
	return "crypto.signature_verified"
}

// KeyPairGenerated is emitted when a new key pair is generated.
type KeyPairGenerated struct {
	Algorithm   PQCAlgorithm
	GeneratedAt time.Time
}

func (e KeyPairGenerated) OccurredAt() time.Time {
	return e.GeneratedAt
}

func (e KeyPairGenerated) EventType() string {
	return "crypto.keypair_generated"
}
