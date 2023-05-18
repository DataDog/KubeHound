package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals"
)

const (
	PodIngestName = "k8s-pod-ingest"
)

type PodIngest struct {
	BaseObjectIngest
}

var _ ObjectIngest = (*PodIngest)(nil)

func (i PodIngest) Name() string {
	return PodIngestName
}

func (i PodIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	return globals.ErrNotImplemented
}

func (i PodIngest) Run(ctx context.Context) error {
	return globals.ErrNotImplemented
}

func (i PodIngest) Close(ctx context.Context) error {
	return globals.ErrNotImplemented
}
