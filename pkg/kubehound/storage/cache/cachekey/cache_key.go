package cachekey

// CacheKey defines a generic, provider agnostic abstraction of a cache key.
type CacheKey interface {
	// Shard returns the shard (aka cache namespace) to which the cache key belongs.
	Shard() string

	// Key returns the value to be used in the cachec lookup operation.
	Key() string
}
