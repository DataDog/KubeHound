package graphdb

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/stretchr/testify/assert"
)

func microBatcherTestInstance(t *testing.T) (*microBatcher, *atomic.Int32) {
	t.Helper()

	var (
		writerFuncCalledCount atomic.Int32
	)

	underTest := newMicroBatcher(log.DefaultLogger(), 5, 1,
		func(_ context.Context, _ []any) error {
			writerFuncCalledCount.Add(1)

			return nil
		})

	return underTest, &writerFuncCalledCount
}

func TestMicroBatcher_AfterBatchSize(t *testing.T) {
	t.Parallel()

	underTest, writerFuncCalledCount := microBatcherTestInstance(t)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	underTest.Start(ctx)

	for i := 0; i < 10; i++ {
		assert.NoError(t, underTest.Enqueue(ctx, i))
	}

	assert.NoError(t, underTest.Flush(ctx))

	assert.Equal(t, int32(2), writerFuncCalledCount.Load())
}

func TestMicroBatcher_AfterFlush(t *testing.T) {
	t.Parallel()

	underTest, writerFuncCalledCount := microBatcherTestInstance(t)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	underTest.Start(ctx)

	for i := 0; i < 11; i++ {
		assert.NoError(t, underTest.Enqueue(ctx, i))
	}

	assert.NoError(t, underTest.Flush(ctx))

	assert.Equal(t, int32(3), writerFuncCalledCount.Load())
}
