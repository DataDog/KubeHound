package worker

import (
	"context"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/alitto/pond"
)

const (
	DefaultPoolSize     = 10
	DefaultPoolCapacity = 100
	DefaultWaitTimeout  = 30 * time.Second
)

// PondWorkerPool implements a worker pool based on https://github.com/alitto/pond
type PondWorkerPool struct {
	pool  *pond.WorkerPool
	group *pond.TaskGroupWithContext
}

// newPond creates a new PondWorkerPool instance.
// This function should not be called directly, but invoked via the factory method.
func newPond(cfg *config.KubehoundConfig) WorkerPool {
	// TODO override defaults from configuration
	return &PondWorkerPool{
		pool: pond.New(DefaultPoolSize, DefaultPoolCapacity, pond.Strategy(pond.Eager())),
	}
}

// Start starts the worker pool and returns a derived context which is canceled the first time
// a function submitted to the group returns a non-nil error or the first time Wait returns,
// whichever occurs first.
func (wp *PondWorkerPool) Start(parent context.Context) (context.Context, error) {
	group, ctx := wp.pool.GroupContext(parent)
	wp.group = group
	return ctx, nil
}

// Submit sends a work item to the worker pool to be processed.
func (wp *PondWorkerPool) Submit(workFunc func() error) {
	wp.group.Submit(workFunc)
}

// Stop stops this pool and waits until either all tasks in the queue are completed
// or the default deadline is reached, whichever comes first.
func (wp *PondWorkerPool) Stop() error {
	wp.pool.StopAndWaitFor(DefaultWaitTimeout)
	return nil
}

// WaitForComplete blocks until either all the tasks submitted to this group have completed,
// one of them returned a non-nil error or the context associated to this group was canceled.
func (wp *PondWorkerPool) WaitForComplete() error {
	return wp.group.Wait()
}
