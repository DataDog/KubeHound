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
	_, ok := m.data[keyId]
	if ok {
		if m.opts.Test {
			return ErrCacheEntryOverwrite
		}

		_ = statsd.Incr(telemetry.MetricCacheDuplicateEntry, []string{keyId}, 1)
		log.Trace(ctx).Warnf("overwriting cache entry: [key]%s", keyId)
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
