package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	rbacv1 "k8s.io/api/rbac/v1"
)

type ClusterRoleBindingIngestor struct {
	buffer map[string]*rbacv1.ClusterRoleBindingList
	writer writer.DumperWriter
}

func NewClusterRoleBindingIngestor(ctx context.Context, dumpWriter writer.DumperWriter) *ClusterRoleBindingIngestor {
	return &ClusterRoleBindingIngestor{
		buffer: make(map[string]*rbacv1.ClusterRoleBindingList),
		writer: dumpWriter,
	}
}

func (d *ClusterRoleBindingIngestor) IngestClusterRoleBinding(ctx context.Context, clusterRoleBinding types.ClusterRoleBindingType) error {
	if ok, err := preflight.CheckClusterRoleBinding(clusterRoleBinding); !ok {
		return err
	}

	return bufferObject[rbacv1.ClusterRoleBindingList, types.ClusterRoleBindingType](ctx, collector.ClusterRoleBindingsPath, d.buffer, clusterRoleBinding)
}

// Complete() is invoked by the collector when all k8s assets have been streamed.
// The function flushes all writers and waits for completion.
func (d *ClusterRoleBindingIngestor) Complete(ctx context.Context) error {
	return dumpObj[*rbacv1.ClusterRoleBindingList](ctx, d.buffer, d.writer)
}
