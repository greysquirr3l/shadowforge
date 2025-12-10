package crypto

import (
	"context"
	"fmt"
	"time"

	dto "github.com/greysquirr3l/shadowforge/internal/application/dto/crypto"
	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

// ListKeysQuery retrieves a paginated list of cryptographic keys with optional filtering
type ListKeysQuery struct {
	request dto.ListKeysRequest
}

// NewListKeysQuery creates a new ListKeysQuery instance
func NewListKeysQuery(req dto.ListKeysRequest) *ListKeysQuery {
	return &ListKeysQuery{request: req}
}

// QueryName returns the unique name of this query
func (q *ListKeysQuery) QueryName() string {
	return "crypto.ListKeys"
}

// Validate ensures the query parameters are valid
func (q *ListKeysQuery) Validate() error {
	if q.request.Limit <= 0 {
		return fmt.Errorf("limit must be greater than 0, got %d", q.request.Limit)
	}

	if q.request.Offset < 0 {
		return fmt.Errorf("offset must be non-negative, got %d", q.request.Offset)
	}

	// Validate algorithm if provided (optional filter)
	if q.request.Algorithm != nil {
		if !q.request.Algorithm.IsValid() {
			return fmt.Errorf("invalid algorithm filter: %s", q.request.Algorithm.Name())
		}
	}

	return nil
}

// Request returns the underlying request object
func (q *ListKeysQuery) Request() dto.ListKeysRequest {
	return q.request
}

// ListKeysQueryHandler handles the execution of ListKeysQuery
type ListKeysQueryHandler struct {
	repository domain_crypto.Repository
}

// NewListKeysQueryHandler creates a new handler for ListKeysQuery
func NewListKeysQueryHandler(repo domain_crypto.Repository) *ListKeysQueryHandler {
	return &ListKeysQueryHandler{
		repository: repo,
	}
}

// Handle executes the query and returns the result
func (h *ListKeysQueryHandler) Handle(ctx context.Context, query interface{}) (interface{}, error) {
	// Type assert to ListKeysQuery
	listQuery, ok := query.(*ListKeysQuery)
	if !ok {
		return nil, fmt.Errorf("expected *ListKeysQuery, got %T", query)
	}

	// Get request parameters
	req := listQuery.Request()

	// Retrieve keys from repository with optional algorithm filter
	keyPairs, totalCount, err := h.repository.ListKeyPairs(
		ctx,
		req.Algorithm, // nil means no filter
		req.Limit,
		req.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list key pairs: %w", err)
	}

	// Build key summaries from KeyPair entities
	keySummaries := make([]dto.KeySummary, 0, len(keyPairs))
	now := time.Now()

	for _, keyPair := range keyPairs {
		// Calculate IsExpired flag
		isExpired := false
		if keyPair.ExpiresAt != nil && keyPair.ExpiresAt.Before(now) {
			isExpired = true
		}

		summary := dto.KeySummary{
			KeyID:     keyPair.ID,
			Algorithm: keyPair.Algorithm,
			CreatedAt: keyPair.CreatedAt,
			ExpiresAt: keyPair.ExpiresAt,
			IsExpired: isExpired,
		}

		keySummaries = append(keySummaries, summary)
	}

	// Build response with pagination metadata
	response := dto.ListKeysResponse{
		Keys:   keySummaries,
		Count:  totalCount,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	return response, nil
}
