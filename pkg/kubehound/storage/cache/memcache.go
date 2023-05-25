package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type MemCacheProvider struct {
	data map[string]string
	mu   *sync.RWMutex
}

func NewCacheProvider(ctx context.Context) (*MemCacheProvider, error) {

	cacheProvider := &MemCacheProvider{
		data: make(map[string]string),
	}

	return cacheProvider, nil
}

func (mp *MemCacheProvider) Name() string {
	return "MemCacheProvider"
}

func (m *MemCacheProvider) GetKeyName(cacheKey CacheKey) string {
	return fmt.Sprintf("%s##%s", cacheKey.Namespace(), cacheKey.Key())
}

func (m *MemCacheProvider) Get(key CacheKey) (string, error) {
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

func (m *MemCacheAsyncWriter) Queue(ctx context.Context, key CacheKey, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	keyId := m.GetKeyName(key)
	_, ok := m.data[keyId]
	if ok {
		return errors.New("entry already present in cache")
	}
	m.data[keyId] = value
	return nil
}

func (m *MemCacheAsyncWriter) Flush(ctx context.Context) error {
	return nil
}

func (m *MemCacheAsyncWriter) Close(ctx context.Context) error {
	return nil
}
