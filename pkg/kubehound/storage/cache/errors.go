package cache

import (
	"errors"
)

var (
	errCacheEntryOverwrite = errors.New("cache entry already exists in test and set operation")
)

// OverwriteError is a custom error type used by the cache implementation is an overwrite occurs when the WithTest option is set.
type OverwriteError struct {
	existingEntry *CacheResult
	err           error
}

// NewOverwriteError returns a new overwrite error, specifying the existing value in the cache that generated the error.
func NewOverwriteError(existing *CacheResult) error {
	return &OverwriteError{
		existingEntry: existing,
		err:           errCacheEntryOverwrite,
	}
}

// Error implements the standard error interface.
func (e OverwriteError) Error() string {
	return e.err.Error()
}

// Existing returns the existing value in the cache that generated the error.
func (e OverwriteError) Existing() *CacheResult {
	return e.existingEntry
}
