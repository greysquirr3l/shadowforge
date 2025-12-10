// Package commands provides CQRS command infrastructure tests.
package commands_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/greysquirr3l/shadowforge/internal/application/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Command implementation
type testCommand struct {
	name      string
	shouldErr bool
}

func (c testCommand) CommandName() string {
	return c.name
}

func (c testCommand) Validate() error {
	if c.shouldErr {
		return errors.New("validation failed")
	}
	return nil
}

// Test CommandHandler implementation
type testCommandHandler struct {
	executed bool
	err      error
	mu       sync.Mutex
}

func (h *testCommandHandler) Handle(ctx context.Context, cmd commands.Command) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.executed = true
	return h.err
}

// Test CommandHandlerWithResult implementation
type testCommandWithResultHandler struct {
	executed bool
	result   string
	err      error
}

func (h *testCommandWithResultHandler) Handle(ctx context.Context, cmd testCommand) (string, error) {
	h.executed = true
	return h.result, h.err
}

func TestCommandBus_NewCommandBus(t *testing.T) {
	bus := commands.NewCommandBus()
	require.NotNil(t, bus)
}

func TestCommandBus_RegisterHandler_Success(t *testing.T) {
	bus := commands.NewCommandBus()
	handler := &testCommandHandler{}

	err := bus.RegisterHandler("test.command", handler)
	assert.NoError(t, err)
}

func TestCommandBus_RegisterHandler_AlreadyRegistered(t *testing.T) {
	bus := commands.NewCommandBus()
	handler := &testCommandHandler{}

	// Register first time - should succeed
	err := bus.RegisterHandler("test.command", handler)
	require.NoError(t, err)

	// Register second time - should fail
	err = bus.RegisterHandler("test.command", handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestCommandBus_Execute_Success(t *testing.T) {
	bus := commands.NewCommandBus()
	handler := &testCommandHandler{}

	err := bus.RegisterHandler("test.command", handler)
	require.NoError(t, err)

	cmd := testCommand{name: "test.command", shouldErr: false}
	err = bus.Execute(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, handler.executed, "Handler should have been executed")
}

func TestCommandBus_Execute_HandlerNotFound(t *testing.T) {
	bus := commands.NewCommandBus()
	cmd := testCommand{name: "unknown.command", shouldErr: false}

	err := bus.Execute(context.Background(), cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no handler registered")
}

func TestCommandBus_Execute_ValidationFailed(t *testing.T) {
	bus := commands.NewCommandBus()
	handler := &testCommandHandler{}

	err := bus.RegisterHandler("test.command", handler)
	require.NoError(t, err)

	cmd := testCommand{name: "test.command", shouldErr: true}
	err = bus.Execute(context.Background(), cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command validation failed")
	assert.False(t, handler.executed, "Handler should not have been executed")
}

func TestCommandBus_Execute_HandlerError(t *testing.T) {
	bus := commands.NewCommandBus()
	expectedErr := errors.New("handler execution failed")
	handler := &testCommandHandler{err: expectedErr}

	err := bus.RegisterHandler("test.command", handler)
	require.NoError(t, err)

	cmd := testCommand{name: "test.command", shouldErr: false}
	err = bus.Execute(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.True(t, handler.executed)
}

func TestCommandBus_ExecuteWithResult_Success(t *testing.T) {
	bus := commands.NewCommandBus()
	expectedResult := "success result"
	handler := &testCommandWithResultHandler{result: expectedResult}

	err := bus.RegisterHandler("test.command", handler)
	require.NoError(t, err)

	cmd := testCommand{name: "test.command", shouldErr: false}
	result, err := commands.ExecuteWithResult[testCommand, string](context.Background(), bus, cmd)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.True(t, handler.executed)
}

func TestCommandBus_ExecuteWithResult_HandlerNotFound(t *testing.T) {
	bus := commands.NewCommandBus()
	cmd := testCommand{name: "unknown.command", shouldErr: false}

	result, err := commands.ExecuteWithResult[testCommand, string](context.Background(), bus, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no handler registered")
	assert.Empty(t, result)
}

func TestCommandBus_ExecuteWithResult_ValidationFailed(t *testing.T) {
	bus := commands.NewCommandBus()
	handler := &testCommandWithResultHandler{result: "should not see this"}

	err := bus.RegisterHandler("test.command", handler)
	require.NoError(t, err)

	cmd := testCommand{name: "test.command", shouldErr: true}
	result, err := commands.ExecuteWithResult[testCommand, string](context.Background(), bus, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command validation failed")
	assert.Empty(t, result)
	assert.False(t, handler.executed)
}

func TestCommandBus_ExecuteWithResult_HandlerError(t *testing.T) {
	bus := commands.NewCommandBus()
	expectedErr := errors.New("handler execution failed")
	handler := &testCommandWithResultHandler{result: "", err: expectedErr}

	err := bus.RegisterHandler("test.command", handler)
	require.NoError(t, err)

	cmd := testCommand{name: "test.command", shouldErr: false}
	result, err := commands.ExecuteWithResult[testCommand, string](context.Background(), bus, cmd)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, "", result)
	assert.True(t, handler.executed)
}

func TestCommandBus_ConcurrentExecution(t *testing.T) {
	bus := commands.NewCommandBus()
	handler := &testCommandHandler{}

	err := bus.RegisterHandler("test.command", handler)
	require.NoError(t, err)

	// Execute 100 commands concurrently
	const numGoroutines = 100
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			cmd := testCommand{name: "test.command", shouldErr: false}
			errChan <- bus.Execute(context.Background(), cmd)
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		assert.NoError(t, err)
	}
}

func TestCommandBus_ContextCancellation(t *testing.T) {
	bus := commands.NewCommandBus()
	handler := &testCommandHandler{}

	err := bus.RegisterHandler("test.command", handler)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cmd := testCommand{name: "test.command", shouldErr: false}
	// Command should still execute (context is passed to handler, not bus)
	err = bus.Execute(ctx, cmd)
	assert.NoError(t, err)
	assert.True(t, handler.executed)
}
