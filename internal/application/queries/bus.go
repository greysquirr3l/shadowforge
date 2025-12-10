// Package queries provides query bus infrastructure.
package queries

import (
"context"
"fmt"
"sync"
)

// QueryBus coordinates query execution.
// It maintains a registry of query handlers and routes queries to them.
type QueryBus struct {
handlers map[string]interface{}
mu       sync.RWMutex
}

// NewQueryBus creates a new query bus.
func NewQueryBus() *QueryBus {
return &QueryBus{
handlers: make(map[string]interface{}),
}
}

// RegisterHandler registers a query handler.
func (b *QueryBus) RegisterHandler(queryName string, handler interface{}) error {
b.mu.Lock()
defer b.mu.Unlock()

if _, exists := b.handlers[queryName]; exists {
return fmt.Errorf("handler already registered for query: %s", queryName)
}

b.handlers[queryName] = handler
return nil
}

// Execute executes a query and returns a result.
func Execute[T Query, R any](ctx context.Context, b *QueryBus, query T) (R, error) {
var zero R

b.mu.RLock()
handlerInterface, exists := b.handlers[query.QueryName()]
b.mu.RUnlock()

if !exists {
return zero, fmt.Errorf("no handler registered for query: %s", query.QueryName())
}

// Validate query before execution
if err := query.Validate(); err != nil {
return zero, fmt.Errorf("query validation failed: %w", err)
}

// Type assertion to QueryHandler
handler, ok := handlerInterface.(QueryHandler[T, R])
if !ok {
return zero, fmt.Errorf("invalid handler type for query: %s", query.QueryName())
}

return handler.Handle(ctx, query)
}
