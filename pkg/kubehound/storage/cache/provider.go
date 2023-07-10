package cache

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/services"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
)

// CacheDriver defines the interface for implementations of the cache provider for intermediate caching of K8s relationship data.
type CacheDriver interface {
	services.Dependency

	// Close cleans up any resources used by the CacheDriver implementation. CacheDriver cannot be reused after this call.
	Close(ctx context.Context) error
}

// CacheReader defines the interface for reading data from the cache provider.
//
//go:generate mockery --name CacheReader --output mocks --case underscore --filename cache_reader.go --with-expecter
type CacheReader interface {
	CacheDriver

	// Get fetches an entry from the cache for the provided cache key.
	Get(ctx context.Context, key cachekey.CacheKey) *CacheResult
}

// CacheProvider defines the interface for reading and writing data from the cache provider.
//
//go:generate mockery --name CacheProvider --output mocks --case underscore --filename cache_provider.go --with-expecter
type CacheProvider interface {
	CacheReader

	// BulkWriter creates a new AsyncWriter instance to enable asynchronous bulk inserts.
	BulkWriter(ctx context.Context, opts ...WriterOption) (AsyncWriter, error)
}

// AysncWriter defines the interface for writer clients to queue aysnchronous, batched writes to the cache.
//
//go:generate mockery --name AsyncWriter --output mocks --case underscore --filename cache_writer.go --with-expecter
type AsyncWriter interface {
	// Queue add a model to an asynchronous write queue. Non-blocking.
	Queue(ctx context.Context, key cachekey.CacheKey, value any) error

	// Flush triggers writes of any remaining items in the queue. Blocks until operation completes
	Flush(ctx context.Context) error

	// Close cleans up any resources used by the AsyncWriter implementation. Writer cannot be reused after this call.
	Close(ctx context.Context) error
}

// Factory returns an initialized instance of a cache provider from the provided application config.
func Factory(ctx context.Context, cfg *config.KubehoundConfig) (CacheProvider, error) {
	provider, err := NewMemCacheProvider(ctx)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
