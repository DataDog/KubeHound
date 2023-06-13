package cache

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

type MemCacheAsyncWriter struct {
	data map[string]string
	mu   *sync.RWMutex
	opts *writerOptions
}

func (m *MemCacheAsyncWriter) Queue(ctx context.Context, key cachekey.CacheKey, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	keyId := computeKey(key)
	_, ok := m.data[keyId]
	if ok {
		if m.opts.Test {
			return ErrCacheEntryOverwrite
		} else {
			log.I.Warnf("overwriting cache entry: [key]%s", keyId)
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
