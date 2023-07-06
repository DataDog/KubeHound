package worker

import (
	"context"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/alitto/pond"
)

const (
	DefaultPoolSize     = 5
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
