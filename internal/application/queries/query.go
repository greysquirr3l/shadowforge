// Package queries provides CQRS query infrastructure.
package queries

import "context"

// Query is the interface that all queries must implement.
// Queries represent read operations (no state changes) in the application.
type Query interface {
// QueryName returns the unique name identifier for this query.
QueryName() string

// Validate validates the query data before execution.
Validate() error
}

// QueryHandler is the interface for query handlers.
// Each query handler processes exactly one query type and returns a result.
type QueryHandler[T Query, R any] interface {
// Handle processes the query and returns a result or error.
Handle(ctx context.Context, query T) (R, error)
}
