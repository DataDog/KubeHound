package worker

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
)

// WorkerPool provides a worker pool for parallelised processing tasks.
type WorkerPool interface {
	// Submit sends a work item to the worker pool to be processed.
	Submit(workFunc func() error)

	// Start starts the worker pool and returns a derived context which is canceled the first time
	// a function submitted to the group returns a non-nil error or the first time Wait returns,
	// whichever occurs first.
	Start(parent context.Context) (context.Context, error)

	// WaitForComplete blocks until either all the tasks submitted to this group have completed,
	// one of them returned a non-nil error or the context associated to this group was canceled.
	WaitForComplete() error
}

// PoolFactory creates a new worker pool instance from the provided config.
func PoolFactory(cfg *config.KubehoundConfig) (WorkerPool, error) {
	return newPond(cfg), nil
}
