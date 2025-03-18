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
