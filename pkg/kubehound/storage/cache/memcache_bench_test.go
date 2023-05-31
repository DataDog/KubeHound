package cache

import (
	"context"
	"fmt"
	"math"
	"testing"
)

// Chip: Apple M1 Max
// Total Number of Cores: 10 (8 performance and 2 efficiency)
// Memory: 64 GB
// BenchmarkWrite/1024-10              1270           1100644 ns/op
// BenchmarkWrite/2048-10               613           2237067 ns/op
// BenchmarkWrite/4096-10               253           4609347 ns/op
// BenchmarkWrite/8192-10               148           8639079 ns/op
// BenchmarkWrite/16384-10               82          16947682 ns/op
// BenchmarkWrite/32768-10               39          34710559 ns/op
// BenchmarkWrite/65536-10               19          68294311 ns/op
// BenchmarkWrite/131072-10               9         138663819 ns/op
// BenchmarkWrite/262144-10               4         277446823 ns/op
// BenchmarkWrite/524288-10               2         552829500 ns/op
// BenchmarkWrite/1048576-10              1        1098627292 ns/op
func BenchmarkWrite(b *testing.B) {
	ctx := context.Background()

	for k := 10.; k <= 20; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			fakeProvider, _ := NewCacheProvider(ctx)
			fakeCacheWriter, _ := fakeProvider.BulkWriter(ctx)
			for i := 0; i < b.N*n; i++ {

				containerKey := ContainerKey(fmt.Sprintf("%ftestPod%d", k, i), fmt.Sprintf("%ftestContainer%d", k, i))
				fakeCacheWriter.Queue(ctx, containerKey, fmt.Sprintf("%ftestContainerID%d", k, i))
			}
			fakeProvider.Close(ctx)
		})
	}
}

// Chip: Apple M1 Max
// Total Number of Cores: 10 (8 performance and 2 efficiency)
// Memory: 64 GB
// BenchmarkRead/1024-10           1000000000               0.0002529 ns/op
// BenchmarkRead/2048-10           1000000000               0.0003826 ns/op
// BenchmarkRead/4096-10           1000000000               0.0008309 ns/op
// BenchmarkRead/8192-10           1000000000               0.001593 ns/op
// BenchmarkRead/16384-10          1000000000               0.003414 ns/op
// BenchmarkRead/32768-10          1000000000               0.007371 ns/op
// BenchmarkRead/65536-10          1000000000               0.01987 ns/op
// BenchmarkRead/131072-10         1000000000               0.03713 ns/op
// BenchmarkRead/262144-10         1000000000               0.07691 ns/op
// BenchmarkRead/524288-10         1000000000               0.1851 ns/op
// BenchmarkRead/1048576-10        1000000000               0.3632 ns/op
func BenchmarkRead(b *testing.B) {
	ctx := context.Background()

	for k := 10.; k <= 20; k++ {
		n := int(math.Pow(2, k))
		fakeProvider, fakeCache := fakeCacheBuilder(ctx, b.N*n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for key, _ := range fakeCache {
				fakeProvider.Get(ctx, key)
			}
		})
	}
}
