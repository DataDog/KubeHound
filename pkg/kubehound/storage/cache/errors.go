package cache

import "errors"

var (
	errCacheEntryOverwrite = errors.New("cache entry already exists in test and set operation")
)

type OverwriteError struct {
	existingEntry *CacheResult
	err           error
}

func NewOverwriteError(existing *CacheResult) error {
	return &OverwriteError{
		existingEntry: existing,
		err:           errCacheEntryOverwrite,
	}
}

func (e OverwriteError) Error() string {
	return e.err.Error()
}

func (e OverwriteError) Existing() *CacheResult {
	return e.existingEntry
}
