package cache

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
)

type MemCacheAsyncWriter struct {
	data map[string]any
	mu   *sync.RWMutex
	opts *writerOptions
}

func (m *MemCacheAsyncWriter) Queue(ctx context.Context, key cachekey.CacheKey, value any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_ = statsd.Incr(telemetry.MetricCacheWrite, []string{}, 1)
	keyId := computeKey(key)
	entry, ok := m.data[keyId]
	if ok {
		if m.opts.Test {
			// if test & set behaviour is specified, return an error containing the existing value in the cache
			return NewOverwriteError(&CacheResult{Value: entry})
		}

		if !m.opts.ExpectOverwrite {
			// if overwrite is expected (e.g fast tracking of existence regardless of value), suppress metrics and logs
			_ = statsd.Incr(telemetry.MetricCacheDuplicateEntry, []string{keyId}, 1)
			log.Trace(ctx).Warnf("overwriting cache entry key=%s old=%#v new=%#v", keyId, entry, value)
		}
	}

	m.data[keyId] = value
	return nil
}

func (m *MemCacheAsyncWriter) Flush(ctx context.Context) error {
	return nil
}

func (m *MemCacheAsyncWriter) Close(ctx context.Context) error {
	// Underlying data map is owned by the proviuder object
	return nil
}
