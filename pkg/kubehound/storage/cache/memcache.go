package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

type MemCacheProvider struct {
	data map[string]string
	mu   *sync.RWMutex
}

func NewCacheProvider(ctx context.Context) (*MemCacheProvider, error) {

	var mu sync.RWMutex
	cacheProvider := &MemCacheProvider{
		data: make(map[string]string),
		mu:   &mu,
	}

	return cacheProvider, nil
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

func (m *MemCacheProvider) GetKeyName(cacheKey cachekey.CacheKey) string {
	return fmt.Sprintf("%s##%s", cacheKey.Namespace(), cacheKey.Key())
}

func (m *MemCacheProvider) Get(ctx context.Context, key cachekey.CacheKey) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var err error
	data, ok := m.data[m.GetKeyName(key)]
	if !ok {
		err = errors.New("entry not found in cache")
	}

	return data, err
}

func (m *MemCacheProvider) BulkWriter(ctx context.Context) (AsyncWriter, error) {
	memCacheWriter := &MemCacheAsyncWriter{}
	memCacheWriter.data = m.data
	memCacheWriter.mu = m.mu
	return memCacheWriter, nil
}

type MemCacheAsyncWriter struct {
	MemCacheProvider
}

func (m *MemCacheAsyncWriter) Queue(ctx context.Context, key cachekey.CacheKey, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	keyId := m.GetKeyName(key)
	_, ok := m.data[keyId]
	if ok {
		log.I.Warnf("overwriting cache entry: [key]%s", keyId)
	}
	m.data[keyId] = value
	return nil
}

func (m *MemCacheAsyncWriter) Flush(ctx context.Context) error {
	m.Close(ctx)
	return nil
}

func (m *MemCacheAsyncWriter) Close(ctx context.Context) error {
	m.MemCacheProvider.Close(ctx)
	return nil
}
