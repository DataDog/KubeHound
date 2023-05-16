package worker

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
)

// WorkerPool provides a worker pool for parallelised processing tasks.
type WorkerPool interface {
	// Submit submits a work item to the queue to be consumed by the next available worker.
	Submit(workFunc func() error)

	// TODO
	Start(parent context.Context) (context.Context, error)

	// Wait blocks until either all the work items have completed, one of them returned a
	// non-nil error or the context associated to this pool was canceled.
	WaitForComplete() error
}

// PoolFactory creates a new worker pool instance from the provided config.
func PoolFactory(cfg *config.KubehoundConfig) (WorkerPool, error) {
	return newPond(cfg), globals.ErrNotImplemented
}
