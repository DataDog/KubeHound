package cache

import (
	"context"
	"fmt"
	"math"
	"testing"
)

func BenchmarkWrite(b *testing.B) {
	ctx := context.Background()

	fakeProvider, _ := NewCacheProvider(ctx)
	fakeCacheWriter, _ := fakeProvider.BulkWriter(ctx)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				containerKey := ContainerKey(fmt.Sprintf("testPod%d", n), fmt.Sprintf("testContainer%d", n))
				fakeCacheWriter.Queue(ctx, containerKey, fmt.Sprintf("testContainerID%d", n))
			}
		})
	}
}

func BenchmarkRead(b *testing.B) {
	ctx := context.Background()

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		fakeProvider, fakeCache := fakeCacheBuilder(ctx, b.N*n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for key, _ := range fakeCache {
				fakeProvider.Get(ctx, key)
			}
		})
	}
}
