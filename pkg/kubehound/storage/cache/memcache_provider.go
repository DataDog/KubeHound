package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
)

type MemCacheProvider struct {
	data map[string]string
	mu   *sync.RWMutex
}

// NewMemCacheProvider returns a new cache provider based on a simple in-memory map.
func NewMemCacheProvider(ctx context.Context) (*MemCacheProvider, error) {
	var mu sync.RWMutex
	cacheProvider := &MemCacheProvider{
		data: make(map[string]string),
		mu:   &mu,
	}

	return cacheProvider, nil
}

// computeKey transforms the cachekey input into a string value to use as a key in the underlying map.
func computeKey(cacheKey cachekey.CacheKey) string {
	return fmt.Sprintf("%s##%s", cacheKey.Shard(), cacheKey.Key())
}

func (mp *MemCacheProvider) Name() string {
	return "MemCacheProvider"
}

func (m *MemCacheProvider) Close(ctx context.Context) error {
	m.data = make(map[string]string)
	return nil
}

func (m *MemCacheProvider) HealthCheck(ctx context.Context) (bool, error) {
	return true, nil
}

func (m *MemCacheProvider) Get(ctx context.Context, key cachekey.CacheKey) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var err error
	data, ok := m.data[computeKey(key)]
	if !ok {
		_ = statsd.Incr(MetricCacheMiss, []string{}, 1)
		err = errors.New("entry not found in cache")
	} else {
		_ = statsd.Incr(MetricCacheHit, []string{}, 1)
	}

	return data, err
}

func (m *MemCacheProvider) BulkWriter(ctx context.Context, opts ...WriterOption) (AsyncWriter, error) {
	memCacheWriter := &MemCacheAsyncWriter{}
	memCacheWriter.data = m.data
	memCacheWriter.mu = m.mu

	wOpts := &writerOptions{}
	for _, o := range opts {
		o(wOpts)
	}

	memCacheWriter.opts = wOpts

	return memCacheWriter, nil
}
