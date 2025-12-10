package crypto

import (
	"fmt"

	"github.com/google/uuid"
)

// PayloadID is a value object representing a unique identifier for a crypto payload.
type PayloadID struct {
	value string
}

// NewPayloadID creates a new PayloadID with validation.
func NewPayloadID(id string) (PayloadID, error) {
	if id == "" {
		return PayloadID{}, ErrInvalidPayloadID
	}
	return PayloadID{value: id}, nil
}

// GeneratePayloadID creates a new random PayloadID using UUID v4.
func GeneratePayloadID() PayloadID {
	return PayloadID{value: uuid.New().String()}
}

// GenerateKeyPairID creates a new random ID for a KeyPair using UUID v4.
func GenerateKeyPairID() string {
	return uuid.New().String()
}

// String returns the string representation of the PayloadID.
func (p PayloadID) String() string {
	return p.value
}

// IsZero checks if the PayloadID is the zero value.
func (p PayloadID) IsZero() bool {
	return p.value == ""
}

// Equals checks if two PayloadIDs are equal.
func (p PayloadID) Equals(other PayloadID) bool {
	return p.value == other.value
}

// PQCAlgorithm represents a post-quantum cryptography algorithm.
type PQCAlgorithm struct {
	name    string
	keySize int
}

// Post-quantum algorithm constants
var (
	Kyber1024   = PQCAlgorithm{name: "kyber1024", keySize: Kyber1024KeySize}
	Dilithium3  = PQCAlgorithm{name: "dilithium3", keySize: Dilithium3KeySize}
	UnknownAlgo = PQCAlgorithm{name: "unknown", keySize: 0}
)

// PQC key size constants (from NIST standards)
const (
	Kyber1024KeySize   = 1568 // Kyber-1024 public key size in bytes
	Dilithium3KeySize  = 1952 // Dilithium3 public key size in bytes
	Dilithium3SignSize = 3293 // Dilithium3 signature size in bytes
)

// NewPQCAlgorithm creates a new PQCAlgorithm from a string name.
func NewPQCAlgorithm(name string) (PQCAlgorithm, error) {
	switch name {
	case "kyber1024":
		return Kyber1024, nil
	case "dilithium3":
		return Dilithium3, nil
	default:
		return UnknownAlgo, fmt.Errorf("%w: %s", ErrInvalidAlgorithm, name)
	}
}

// Name returns the algorithm name.
func (a PQCAlgorithm) Name() string {
	return a.name
}

// KeySize returns the expected key size for this algorithm.
func (a PQCAlgorithm) KeySize() int {
	return a.keySize
}

// IsValid checks if the algorithm is valid (not unknown).
func (a PQCAlgorithm) IsValid() bool {
	return a.name != "unknown" && a.keySize > 0
}

// String returns the string representation of the algorithm.
func (a PQCAlgorithm) String() string {
	return a.name
}

// Equals checks if two PQCAlgorithms are equal.
func (a PQCAlgorithm) Equals(other PQCAlgorithm) bool {
	return a.name == other.name
}
