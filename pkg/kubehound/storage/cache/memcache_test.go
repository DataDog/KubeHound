//nolint:containedctx
package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
)

func fakeCacheBuilder(ctx context.Context, cacheSize int) (*MemCacheProvider, map[cachekey.CacheKey]string) {
	fakeProvider, _ := NewMemCacheProvider(ctx)

	fakeCache := make(map[cachekey.CacheKey]string, cacheSize)

	for i := 1; i <= cacheSize; i++ {
		fakeCache[cachekey.Container(fmt.Sprintf("Pod%d", i), fmt.Sprintf("container%d", i), "test")] = fmt.Sprintf("value%d", i)
	}

	fakeCacheWriter, _ := fakeProvider.BulkWriter(ctx)
	for key, val := range fakeCache {
		_ = fakeCacheWriter.Queue(ctx, key, val)
	}

	return fakeProvider, fakeCache
}

func TestMemCacheProvider_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fakeProvider, fakeCache := fakeCacheBuilder(ctx, 3)

	type fields struct {
		data map[string]any
		mu   *sync.RWMutex
	}
	type args struct {
		ctx       context.Context
		fakeCache map[cachekey.CacheKey]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test retrieving element from cache",
			fields: fields{
				data: fakeProvider.data,
				mu:   fakeProvider.mu,
			},
			args: args{
				fakeCache: fakeCache,
				ctx:       ctx,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &MemCacheProvider{
				data: tt.fields.data,
				mu:   tt.fields.mu,
			}

			for key, val := range tt.args.fakeCache {
				got, err := m.Get(tt.args.ctx, key).Text()
				if (err != nil) != tt.wantErr {
					t.Errorf("MemCacheProvider.Get() error = %v, wantErr %v", err, tt.wantErr)

					return
				}
				if got != val {
					t.Errorf("MemCacheProvider.Get() = %v, want %v", got, val)
				}
			}
		})
	}
}

func TestMemCacheAsyncWriter_Queue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Standard write
	fakeProvider1, _ := NewMemCacheProvider(ctx)
	fakeCache1 := map[cachekey.CacheKey]string{
		cachekey.Container("testPod1", "container1", "test"): "qwerty",
		cachekey.Container("testPod2", "container2", "test"): "asdfgh",
		cachekey.Container("testPod3", "container3", "test"): "zxcvb",
	}

	// Testing for collision in cache
	fakeProvider2, fakeCache2 := fakeCacheBuilder(ctx, 3)

	type fields struct {
		MemCacheProvider MemCacheProvider
		Opts             *writerOptions
	}
	type args struct {
		ctx       context.Context
		fakeCache map[cachekey.CacheKey]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test retrieving element from cache",
			fields: fields{
				MemCacheProvider: *fakeProvider1,
				Opts:             &writerOptions{},
			},
			args: args{
				fakeCache: fakeCache1,
				ctx:       ctx,
			},
			wantErr: false,
		},
		{
			name: "Already present in cache",
			fields: fields{
				MemCacheProvider: *fakeProvider2,
				Opts:             &writerOptions{},
			},
			args: args{
				fakeCache: fakeCache2,
				ctx:       ctx,
			},
			wantErr: false,
		},
		{
			name: "Already present in cache",
			fields: fields{
				MemCacheProvider: *fakeProvider2,
				Opts:             &writerOptions{Test: true},
			},
			args: args{
				fakeCache: fakeCache2,
				ctx:       ctx,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &MemCacheAsyncWriter{
				data: tt.fields.MemCacheProvider.data,
				mu:   tt.fields.MemCacheProvider.mu,
				opts: tt.fields.Opts,
			}

			for key, val := range tt.args.fakeCache {
				if err := m.Queue(tt.args.ctx, key, val); (err != nil) != tt.wantErr {
					t.Errorf("MemCacheAsyncWriter.Queue() error = %v, wantErr %v", err, tt.wantErr)
				}

				got, err := tt.fields.MemCacheProvider.Get(tt.args.ctx, key).Text()
				if err != nil {
					t.Errorf("MemCacheProvider.Get() error = %v", err)

					return
				}
				if got != val {
					t.Errorf("MemCacheProvider.Get() = %v, want %v", got, val)
				}
			}
		})
	}
}
