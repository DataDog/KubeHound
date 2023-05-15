package graphbuilder

import (
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
)

type WorkerPool interface {
	Submit(func()) error
	Stop() error
	WaitForComplete()
}

func NewWorkerPool(cfg *config.KubehoundConfig) (WorkerPool, error) {
	return nil, globals.ErrNotImplemented
}

// Implement https://github.com/alitto/pond
