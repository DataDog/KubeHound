package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals"
)

const (
	ClusterRoleBindingIngestName = "k8s-cluster-role-binding-ingest"
)

type ClusterRoleBindingIngest struct {
	IngestResources
}

var _ ObjectIngest = (*ClusterRoleBindingIngest)(nil)

func (i *ClusterRoleBindingIngest) Name() string {
	return ClusterRoleBindingIngestName
}

func (i *ClusterRoleBindingIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	return globals.ErrNotImplemented
}

func (i *ClusterRoleBindingIngest) Run(ctx context.Context) error {
	return globals.ErrNotImplemented
}

func (i *ClusterRoleBindingIngest) Close(ctx context.Context) error {
	return globals.ErrNotImplemented
}
