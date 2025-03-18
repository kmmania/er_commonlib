/*
Package controller provides error definitions for handling common error scenarios in controller operations.

This package defines error types that can be used by controllers to signal specific conditions
like missing resources, conflicts, or invalid inputs. These errors facilitate consistent error
handling and communication between application layers.

Errors:
  - ErrNotFound: Indicates that a requested resource could not be located.
  - ErrAlreadyExists: Signals an attempt to create a resource that conflicts with an existing one.
  - ErrInvalidInput: Used to indicate issues with the provided input, such as invalid or incomplete data.
*/
package controller

import "errors"

var (
	// ErrNotFound is returned by the Controller when a requested TestAMetadata
	// object cannot be found in the repository.
	//
	// This error can be used by higher layers to distinguish cases where a resource
	// is missing.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists is returned by the Controller when an attempt to add a new
	// TestAMetadata object fails due to an existing entry with the same identifier.
	//
	// This error can be used to enforce uniqueness constraints and signal conflicts.
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidInput is returned by the Controller when the input provided to an operation
	// is invalid or does not meet the required criteria.
	//
	// This error can be used to signal issues such as missing required fields, invalid formats,
	// or values that are out of acceptable ranges.
	ErrInvalidInput = errors.New("invalid input")
)
