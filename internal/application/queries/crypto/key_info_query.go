// Package crypto provides queries for cryptographic operations.
package crypto

import (
	"context"
	"fmt"

	"github.com/greysquirr3l/shadowforge/internal/application/dto/crypto"
	"github.com/greysquirr3l/shadowforge/internal/application/queries"
	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

// GetKeyInfoQuery is a query to retrieve information about a cryptographic key.
type GetKeyInfoQuery struct {
	request crypto.KeyInfoRequest
}

// NewGetKeyInfoQuery creates a new GetKeyInfoQuery.
func NewGetKeyInfoQuery(request crypto.KeyInfoRequest) *GetKeyInfoQuery {
	return &GetKeyInfoQuery{
		request: request,
	}
}

// QueryName returns the name of this query.
func (q *GetKeyInfoQuery) QueryName() string {
	return "crypto.GetKeyInfo"
}

// Validate validates the query inputs.
func (q *GetKeyInfoQuery) Validate() error {
	if q.request.KeyID == "" {
		return fmt.Errorf("key ID cannot be empty")
	}

	if !q.request.Algorithm.IsValid() {
		return fmt.Errorf("invalid algorithm: %s", q.request.Algorithm.Name())
	}

	return nil
}

// Request returns the key info request.
func (q *GetKeyInfoQuery) Request() crypto.KeyInfoRequest {
	return q.request
}

// GetKeyInfoQueryHandler handles GetKeyInfoQuery.
type GetKeyInfoQueryHandler struct {
	repository domain_crypto.Repository
}

// NewGetKeyInfoQueryHandler creates a new GetKeyInfoQueryHandler.
func NewGetKeyInfoQueryHandler(repository domain_crypto.Repository) *GetKeyInfoQueryHandler {
	return &GetKeyInfoQueryHandler{
		repository: repository,
	}
}

// Handle executes the GetKeyInfoQuery.
func (h *GetKeyInfoQueryHandler) Handle(ctx context.Context, q queries.Query) (crypto.KeyInfoResponse, error) {
	// Type assert to GetKeyInfoQuery
	keyInfoQuery, ok := q.(*GetKeyInfoQuery)
	if !ok {
		return crypto.KeyInfoResponse{}, fmt.Errorf("expected *GetKeyInfoQuery, got %T", q)
	}

	req := keyInfoQuery.Request()

	// Get key pair from repository
	keyPair, err := h.repository.GetKeyPair(ctx, req.KeyID)
	if err != nil {
		return crypto.KeyInfoResponse{}, fmt.Errorf("failed to get key pair: %w", err)
	}

	// Build response
	response := crypto.KeyInfoResponse{
		KeyID:     keyPair.ID,
		Algorithm: keyPair.Algorithm,
		PublicKey: keyPair.PublicKey,
		CreatedAt: keyPair.CreatedAt,
		ExpiresAt: keyPair.ExpiresAt,
	}

	return response, nil
}
