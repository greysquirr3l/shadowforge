package crypto

import (
	"context"
	"fmt"
	"time"

	dto "github.com/greysquirr3l/shadowforge/internal/application/dto/crypto"
	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

// GetCapacityQuery calculates the encryption capacity and overhead for a given media size
type GetCapacityQuery struct {
	request dto.GetCapacityRequest
}

// NewGetCapacityQuery creates a new GetCapacityQuery instance
func NewGetCapacityQuery(req dto.GetCapacityRequest) *GetCapacityQuery {
	return &GetCapacityQuery{request: req}
}

// QueryName returns the unique name of this query
func (q *GetCapacityQuery) QueryName() string {
	return "crypto.GetCapacity"
}

// Validate ensures the query parameters are valid
func (q *GetCapacityQuery) Validate() error {
	if q.request.MediaSize <= 0 {
		return fmt.Errorf("media size must be greater than 0, got %d", q.request.MediaSize)
	}

	if !q.request.Algorithm.IsValid() {
		return fmt.Errorf("invalid algorithm: %s", q.request.Algorithm.Name())
	}

	return nil
}

// Request returns the underlying request object
func (q *GetCapacityQuery) Request() dto.GetCapacityRequest {
	return q.request
}

// GetCapacityQueryHandler handles the execution of GetCapacityQuery
type GetCapacityQueryHandler struct {
	// Note: This handler calculates overhead based on algorithm constants
	// In a full implementation, this might use a CryptoService for calculations
}

// NewGetCapacityQueryHandler creates a new handler for GetCapacityQuery
func NewGetCapacityQueryHandler() *GetCapacityQueryHandler {
	return &GetCapacityQueryHandler{}
}

// Handle executes the query and returns the result
func (h *GetCapacityQueryHandler) Handle(ctx context.Context, query interface{}) (interface{}, error) {
	// Type assert to GetCapacityQuery
	capacityQuery, ok := query.(*GetCapacityQuery)
	if !ok {
		return nil, fmt.Errorf("expected *GetCapacityQuery, got %T", query)
	}

	// Get request parameters
	req := capacityQuery.Request()

	// Calculate overhead based on algorithm
	overheadBytes := h.calculateOverhead(req.Algorithm)

	// Calculate maximum payload size
	// Payload = MediaSize - Overhead - SafetyMargin
	const safetyMarginBytes = 64 // Safety margin for headers, padding, etc.
	maxPayloadSize := req.MediaSize - overheadBytes - safetyMarginBytes

	if maxPayloadSize < 0 {
		maxPayloadSize = 0
	}

	// Build response
	response := dto.GetCapacityResponse{
		MaxPayloadSize: maxPayloadSize,
		Algorithm:      req.Algorithm,
		OverheadBytes:  overheadBytes,
		CalculatedAt:   time.Now(),
	}

	return response, nil
}

// calculateOverhead returns the cryptographic overhead in bytes for the given algorithm
func (h *GetCapacityQueryHandler) calculateOverhead(algorithm domain_crypto.PQCAlgorithm) int64 {
	switch algorithm {
	case domain_crypto.Kyber1024:
		// Kyber-1024 overhead:
		// - Ciphertext: ~1568 bytes
		// - Shared secret: 32 bytes (used for AES-256-GCM)
		// - AES-GCM nonce: 12 bytes
		// - AES-GCM tag: 16 bytes
		// Total: ~1628 bytes
		return 1628

	case domain_crypto.Dilithium3:
		// Dilithium3 overhead (signature):
		// - Signature: ~3293 bytes
		// - Public key: ~1952 bytes (if included)
		// Total: ~5245 bytes
		return 5245

	default:
		// Unknown algorithm - use conservative estimate
		// Combined Kyber + Dilithium overhead
		return 6873
	}
}
