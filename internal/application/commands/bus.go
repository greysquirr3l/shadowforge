// Package commands provides command bus infrastructure.
package commands

import (
"context"
"fmt"
"sync"
)

// CommandBus coordinates command execution.
// It maintains a registry of command handlers and routes commands to them.
type CommandBus struct {
handlers map[string]interface{}
mu       sync.RWMutex
}

// NewCommandBus creates a new command bus.
func NewCommandBus() *CommandBus {
return &CommandBus{
handlers: make(map[string]interface{}),
}
}

// RegisterHandler registers a command handler.
func (b *CommandBus) RegisterHandler(commandName string, handler interface{}) error {
b.mu.Lock()
defer b.mu.Unlock()

if _, exists := b.handlers[commandName]; exists {
return fmt.Errorf("handler already registered for command: %s", commandName)
}

b.handlers[commandName] = handler
return nil
}

// Execute executes a command without returning a result.
func (b *CommandBus) Execute(ctx context.Context, cmd Command) error {
b.mu.RLock()
handlerInterface, exists := b.handlers[cmd.CommandName()]
b.mu.RUnlock()

if !exists {
return fmt.Errorf("no handler registered for command: %s", cmd.CommandName())
}

// Validate command before execution
if err := cmd.Validate(); err != nil {
return fmt.Errorf("command validation failed: %w", err)
}

// Type assertion to CommandHandler
handler, ok := handlerInterface.(CommandHandler[Command])
if !ok {
return fmt.Errorf("invalid handler type for command: %s", cmd.CommandName())
}

return handler.Handle(ctx, cmd)
}

// ExecuteWithResult executes a command and returns a result.
func ExecuteWithResult[T Command, R any](ctx context.Context, b *CommandBus, cmd T) (R, error) {
var zero R

b.mu.RLock()
handlerInterface, exists := b.handlers[cmd.CommandName()]
b.mu.RUnlock()

if !exists {
return zero, fmt.Errorf("no handler registered for command: %s", cmd.CommandName())
}

// Validate command before execution
if err := cmd.Validate(); err != nil {
return zero, fmt.Errorf("command validation failed: %w", err)
}

// Type assertion to CommandHandlerWithResult
handler, ok := handlerInterface.(CommandHandlerWithResult[T, R])
if !ok {
return zero, fmt.Errorf("invalid handler type for command: %s", cmd.CommandName())
}

return handler.Handle(ctx, cmd)
}
