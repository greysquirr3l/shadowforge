// Package commands provides CQRS command infrastructure.
package commands

import "context"

// Command is the interface that all commands must implement.
// Commands represent write operations (state changes) in the application.
type Command interface {
	// CommandName returns the unique name identifier for this command.
	CommandName() string

	// Validate validates the command data before execution.
	Validate() error
}

// CommandHandler is the interface for command handlers.
// Each command handler processes exactly one command type.
type CommandHandler[T Command] interface {
	// Handle processes the command and returns an error if it fails.
	Handle(ctx context.Context, cmd T) error
}

// CommandHandlerWithResult is a command handler that returns a result.
type CommandHandlerWithResult[T Command, R any] interface {
	// Handle processes the command and returns a result or error.
	Handle(ctx context.Context, cmd T) (R, error)
}
