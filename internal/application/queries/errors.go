// Package queries provides query error types.
package queries

import "errors"

var (
	// ErrQueryValidationFailed indicates query validation failed.
	ErrQueryValidationFailed = errors.New("query validation failed")

	// ErrHandlerNotFound indicates no handler registered for query.
	ErrHandlerNotFound = errors.New("query handler not found")

	// ErrHandlerAlreadyRegistered indicates handler already registered.
	ErrHandlerAlreadyRegistered = errors.New("query handler already registered")

	// ErrQueryExecutionFailed indicates query execution failed.
	ErrQueryExecutionFailed = errors.New("query execution failed")
)
