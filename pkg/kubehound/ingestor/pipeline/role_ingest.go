package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals"
)

const (
	RoleIngestName = "k8s-role-ingest"
)

type RoleIngest struct {
	IngestResources
}

var _ ObjectIngest = (*RoleIngest)(nil)

func (i *RoleIngest) Name() string {
	return RoleIngestName
}

func (i *RoleIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	return globals.ErrNotImplemented
}

func (i *RoleIngest) Run(ctx context.Context) error {
	return globals.ErrNotImplemented
}

func (i *RoleIngest) Close(ctx context.Context) error {
	return globals.ErrNotImplemented
}
