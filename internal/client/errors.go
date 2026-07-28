// Package client implements a Go HTTP client for the Cloud-Hypervisor REST API.
//
// The client supports both Unix domain socket (UDS) and TCP connections,
// typed error handling for CH API status codes, and configurable timeouts.
package client

import (
	"errors"
	"fmt"
)

// Sentinel errors returned by the client when the CH API responds with
// specific HTTP status codes. Callers can use errors.Is to distinguish them.
var (
	// ErrNotFound is returned when the CH API responds with HTTP 404.
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidState is returned when the CH API responds with HTTP 405,
	// indicating the VM is not in the expected state for the operation.
	ErrInvalidState = errors.New("invalid vm state")
)

// Error is a structured error that wraps an HTTP status code and the response
// body snippet returned by the Cloud-Hypervisor API.
type Error struct {
	StatusCode int
	Body       string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("unexpected status %d: %s", e.StatusCode, e.Body)
}

// Unwrap allows errors.Is / errors.As to traverse through sentinel errors
// when an Error wraps one (e.g., by using fmt.Errorf("%w: %w", sentinel, err)).
func (e *Error) Unwrap() error {
	// No parent sentinel — Error is the leaf in the chain.
	// Sentinels like ErrNotFound are embedded in the error message
	// via fmt.Errorf with %w by the calling code.
	return nil
}
