// Package commands provides command error types.
package commands

import "errors"

var (
// ErrCommandValidationFailed indicates command validation failed.
ErrCommandValidationFailed = errors.New("command validation failed")

// ErrHandlerNotFound indicates no handler registered for command.
ErrHandlerNotFound = errors.New("command handler not found")

// ErrHandlerAlreadyRegistered indicates handler already registered.
ErrHandlerAlreadyRegistered = errors.New("command handler already registered")

// ErrCommandExecutionFailed indicates command execution failed.
ErrCommandExecutionFailed = errors.New("command execution failed")
)
