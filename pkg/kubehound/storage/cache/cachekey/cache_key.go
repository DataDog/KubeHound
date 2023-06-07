package cachekey

// CacheKey defines a generic, provider agnostic abstraction of a cache key.
type CacheKey interface {
	// Namespace returns the namespace to which the cache key belongs.
	Namespace() string

	// Key returns the value to be used in the cachec lookup operation.
	Key() string
}
