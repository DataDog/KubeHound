package cache

import (
	"context"
	"sync"
	"testing"
)

func fakeCacheBuilder(ctx context.Context) (*MemCacheProvider, map[CacheKey]string) {
	fakeProvider, _ := NewCacheProvider(ctx)

	fakeCache := map[CacheKey]string{
		ContainerKey("testPod1", "container1"): "qwerty",
		ContainerKey("testPod2", "container2"): "asdfgh",
		ContainerKey("testPod3", "container3"): "zxcvb",
	}

	fakeCacheWriter, _ := fakeProvider.BulkWriter(ctx)
	for key, val := range fakeCache {
		fakeCacheWriter.Queue(context.TODO(), key, val)
	}

	return fakeProvider, fakeCache
}

func TestMemCacheProvider_Get(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	fakeProvider, fakeCache := fakeCacheBuilder(ctx)

	type fields struct {
		data map[string]string
		mu   *sync.RWMutex
	}
	type args struct {
		ctx       context.Context
		fakeCache map[CacheKey]string
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
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			m := &MemCacheProvider{
				data: tt.fields.data,
				mu:   tt.fields.mu,
			}

			for key, val := range tt.args.fakeCache {
				got, err := m.Get(tt.args.ctx, key)
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

	ctx := context.TODO()

	// Standard write
	fakeProvider1, _ := NewCacheProvider(ctx)
	fakeCache1 := map[CacheKey]string{
		ContainerKey("testPod1", "container1"): "qwerty",
		ContainerKey("testPod2", "container2"): "asdfgh",
		ContainerKey("testPod3", "container3"): "zxcvb",
	}

	// Testing for collision in cache
	fakeProvider2, fakeCache2 := fakeCacheBuilder(ctx)

	type fields struct {
		MemCacheProvider MemCacheProvider
	}
	type args struct {
		ctx       context.Context
		fakeCache map[CacheKey]string
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
			},
			args: args{
				fakeCache: fakeCache2,
				ctx:       ctx,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m := &MemCacheAsyncWriter{
				MemCacheProvider: tt.fields.MemCacheProvider,
			}

			for key, val := range tt.args.fakeCache {

				if err := m.Queue(tt.args.ctx, key, val); (err != nil) != tt.wantErr {
					t.Errorf("MemCacheAsyncWriter.Queue() error = %v, wantErr %v", err, tt.wantErr)
				}

				got, err := m.Get(tt.args.ctx, key)
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
