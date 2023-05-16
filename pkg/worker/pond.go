package worker

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/alitto/pond"
)

type PondWorkerPool struct {
}

func newPond(cfg *config.KubehoundConfig) {

	pool := pond.New(10, 1000)

	return
}

func (wp *PondWorkerPool) Start(ctx context.Context) error {

}

func (wp *PondWorkerPool) Submit(workFunc func()) error {

}

func (wp *PondWorkerPool) Stop() error {

}

func (wp *PondWorkerPool) WaitForComplete() error {

}
