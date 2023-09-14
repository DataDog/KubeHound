package worker

import (
	"context"
	"time"

	"github.com/alitto/pond"
)

const (
	DefaultWaitTimeout = 30 * time.Second
)

// PondWorkerPool implements a worker pool based on https://github.com/alitto/pond
type PondWorkerPool struct {
	pool  *pond.WorkerPool
	group *pond.TaskGroupWithContext
}

// newPond creates a new PondWorkerPool instance.
// This function should not be called directly, but invoked via the factory method.
func newPond(size int, capacity int) WorkerPool {
	return &PondWorkerPool{
		pool: pond.New(size, capacity, pond.Strategy(pond.Eager())),
	}
}

func (wp *PondWorkerPool) Start(parent context.Context) (context.Context, error) {
	group, ctx := wp.pool.GroupContext(parent)
	wp.group = group
	return ctx, nil
}

func (wp *PondWorkerPool) Submit(workFunc func() error) {
	wp.group.Submit(workFunc)
}

func (wp *PondWorkerPool) WaitForComplete() error {
	return wp.group.Wait()
}
