/*
Package repository provides error definitions commonly used to interact with data repositories.

This package defines error types to standardize error handling for common repository-related
operations, such as retrieving, creating, or updating records. These errors can be used to
signal specific conditions like missing data or uniqueness violations in a consistent manner.

Errors:
  - ErrNotFound: Indicates that a requested record could not be located.
  - ErrAlreadyExists: Signals an attempt to create a record that conflicts with an existing one.
*/
package repository

import "errors"

var (
	// ErrNotFound is returned when a requested record is not found in the repository.
	// This error signals that an item with the specified identifier or criteria
	// does not exist in the current data store.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists is returned when attempting to create a record that already exists.
	// This error helps enforce uniqueness constraints within the repository and prevent
	// duplicate entries.
	ErrAlreadyExists = errors.New("already exists")
)
