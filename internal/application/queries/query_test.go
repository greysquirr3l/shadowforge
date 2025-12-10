// Package queries provides CQRS query infrastructure tests.
package queries_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/greysquirr3l/shadowforge/internal/application/queries"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Query implementation
type testQuery struct {
	name      string
	shouldErr bool
}

func (q testQuery) QueryName() string {
	return q.name
}

func (q testQuery) Validate() error {
	if q.shouldErr {
		return errors.New("validation failed")
	}
	return nil
}

// Test QueryHandler implementation
type testQueryHandler struct {
	executed bool
	result   string
	err      error
	mu       sync.Mutex
}

func (h *testQueryHandler) Handle(ctx context.Context, query testQuery) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.executed = true
	return h.result, h.err
}

func TestQueryBus_NewQueryBus(t *testing.T) {
	bus := queries.NewQueryBus()
	require.NotNil(t, bus)
}

func TestQueryBus_RegisterHandler_Success(t *testing.T) {
	bus := queries.NewQueryBus()
	handler := &testQueryHandler{}

	err := bus.RegisterHandler("test.query", handler)
	assert.NoError(t, err)
}

func TestQueryBus_RegisterHandler_AlreadyRegistered(t *testing.T) {
	bus := queries.NewQueryBus()
	handler := &testQueryHandler{}

	// Register first time - should succeed
	err := bus.RegisterHandler("test.query", handler)
	require.NoError(t, err)

	// Register second time - should fail
	err = bus.RegisterHandler("test.query", handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestQueryBus_Execute_Success(t *testing.T) {
	bus := queries.NewQueryBus()
	expectedResult := "query result"
	handler := &testQueryHandler{result: expectedResult}

	err := bus.RegisterHandler("test.query", handler)
	require.NoError(t, err)

	query := testQuery{name: "test.query", shouldErr: false}
	result, err := queries.Execute[testQuery, string](context.Background(), bus, query)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.True(t, handler.executed)
}

func TestQueryBus_Execute_HandlerNotFound(t *testing.T) {
	bus := queries.NewQueryBus()
	query := testQuery{name: "unknown.query", shouldErr: false}

	result, err := queries.Execute[testQuery, string](context.Background(), bus, query)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no handler registered")
	assert.Empty(t, result)
}

func TestQueryBus_Execute_ValidationFailed(t *testing.T) {
	bus := queries.NewQueryBus()
	handler := &testQueryHandler{result: "should not see this"}

	err := bus.RegisterHandler("test.query", handler)
	require.NoError(t, err)

	query := testQuery{name: "test.query", shouldErr: true}
	result, err := queries.Execute[testQuery, string](context.Background(), bus, query)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query validation failed")
	assert.Empty(t, result)
	assert.False(t, handler.executed)
}

func TestQueryBus_Execute_HandlerError(t *testing.T) {
	bus := queries.NewQueryBus()
	expectedErr := errors.New("handler execution failed")
	handler := &testQueryHandler{result: "", err: expectedErr}

	err := bus.RegisterHandler("test.query", handler)
	require.NoError(t, err)

	query := testQuery{name: "test.query", shouldErr: false}
	result, err := queries.Execute[testQuery, string](context.Background(), bus, query)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, "", result)
	assert.True(t, handler.executed)
}

func TestQueryBus_ConcurrentExecution(t *testing.T) {
	bus := queries.NewQueryBus()
	expectedResult := "concurrent result"
	handler := &testQueryHandler{result: expectedResult}

	err := bus.RegisterHandler("test.query", handler)
	require.NoError(t, err)

	// Execute 100 queries concurrently
	const numGoroutines = 100
	type result struct {
		value string
		err   error
	}
	resultChan := make(chan result, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			query := testQuery{name: "test.query", shouldErr: false}
			value, err := queries.Execute[testQuery, string](context.Background(), bus, query)
			resultChan <- result{value: value, err: err}
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		res := <-resultChan
		assert.NoError(t, res.err)
		assert.Equal(t, expectedResult, res.value)
	}
}

func TestQueryBus_ContextCancellation(t *testing.T) {
	bus := queries.NewQueryBus()
	handler := &testQueryHandler{result: "result"}

	err := bus.RegisterHandler("test.query", handler)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	query := testQuery{name: "test.query", shouldErr: false}
	// Query should still execute (context is passed to handler, not bus)
	result, err := queries.Execute[testQuery, string](ctx, bus, query)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.True(t, handler.executed)
}
