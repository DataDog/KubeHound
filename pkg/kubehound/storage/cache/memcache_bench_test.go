package cache

import (
	"context"
	"fmt"
	"math"
	"testing"
)

// go test -bench=. -benchmem
// Chip: Apple M1 Max
// Total Number of Cores: 10 (8 performance and 2 efficiency)
// Memory: 64 GB
// BenchmarkWrite/1024-10              1081           1457681 ns/op          502334 B/op      14370 allocs/op
// BenchmarkWrite/2048-10               609           2168746 ns/op          970973 B/op      28733 allocs/op
// BenchmarkWrite/4096-10               292           4378407 ns/op         1965000 B/op      57473 allocs/op
// BenchmarkWrite/8192-10               142           8629822 ns/op         3960456 B/op     114951 allocs/op
// BenchmarkWrite/16384-10               82          21557271 ns/op         7627943 B/op     229879 allocs/op
// BenchmarkWrite/32768-10               28          42693875 ns/op        17067977 B/op     460096 allocs/op
// BenchmarkWrite/65536-10               14          87483321 ns/op        34138948 B/op     920203 allocs/op
// BenchmarkWrite/131072-10               8         159673495 ns/op        65375701 B/op    1839719 allocs/op
// BenchmarkWrite/262144-10               4         300941146 ns/op        130748070 B/op   3679430 allocs/op
// BenchmarkWrite/524288-10               2         583498562 ns/op        261488468 B/op   7358834 allocs/op
// BenchmarkWrite/1048576-10              1        1194777000 ns/op        522988328 B/op  14717714 allocs/op
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

// go test -bench=. -benchmem
// Chip: Apple M1 Max
// Total Number of Cores: 10 (8 performance and 2 efficiency)
// Memory: 64 GB
// BenchmarkRead/1024-10           1000000000               0.0002132 ns/op       0 B/op          0 allocs/op
// BenchmarkRead/2048-10           1000000000               0.0004069 ns/op       0 B/op          0 allocs/op
// BenchmarkRead/4096-10           1000000000               0.0009266 ns/op       0 B/op          0 allocs/op
// BenchmarkRead/8192-10           1000000000               0.001837 ns/op        0 B/op          0 allocs/op
// BenchmarkRead/16384-10          1000000000               0.003899 ns/op        0 B/op          0 allocs/op
// BenchmarkRead/32768-10          1000000000               0.009364 ns/op        0 B/op          0 allocs/op
// BenchmarkRead/65536-10          1000000000               0.01708 ns/op         0 B/op          0 allocs/op
// BenchmarkRead/131072-10         1000000000               0.04346 ns/op         0 B/op          0 allocs/op
// BenchmarkRead/262144-10         1000000000               0.08424 ns/op         0 B/op          0 allocs/op
// BenchmarkRead/524288-10         1000000000               0.1908 ns/op          0 B/op          0 allocs/op
// BenchmarkRead/1048576-10        1000000000               0.3739 ns/op          0 B/op          0 allocs/op
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
