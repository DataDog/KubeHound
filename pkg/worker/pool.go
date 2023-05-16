package worker

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
)

// WorkerPool provides a worker pool for parallelised processing tasks.
type WorkerPool interface {
	// Submit submits a work item to the queue to be consumed by the next available worker.
	Submit(workFunc func()) error

	// TODO
	Start(ctx context.Context) error

	// Stop stops any further work and blocks until all workers have completed shutdown.
	Stop() error

	// WaitForComplete blocks until all work items are completed.
	WaitForComplete() error
}

// PoolFactory creates a new worker pool instance from the provided config.
// TODO Implement https://github.com/alitto/pond
func PoolFactory(cfg *config.KubehoundConfig) (WorkerPool, error) {
	return nil, globals.ErrNotImplemented
}

https://github.com/alitto/pond#submitting-a-group-of-tasks-associated-to-a-context-since-v180
DO this