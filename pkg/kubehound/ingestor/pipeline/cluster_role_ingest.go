package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals"
)

const (
	ClusterRoleIngestName = "k8s-cluster-role-ingest"
)

type ClusterRoleIngest struct {
	IngestResources
}

var _ ObjectIngest = (*ClusterRoleIngest)(nil)

func (i *ClusterRoleIngest) Name() string {
	return ClusterRoleIngestName
}

func (i *ClusterRoleIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	return globals.ErrNotImplemented
}

func (i *ClusterRoleIngest) Run(ctx context.Context) error {
	return globals.ErrNotImplemented
}

func (i *ClusterRoleIngest) Close(ctx context.Context) error {
	return globals.ErrNotImplemented
}
