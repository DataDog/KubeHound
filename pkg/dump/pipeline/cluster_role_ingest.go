package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	rbacv1 "k8s.io/api/rbac/v1"
)

type ClusterRoleIngestor struct {
	buffer map[string]*rbacv1.ClusterRoleList
	writer writer.DumperWriter
}

func NewClusterRoleIngestor(ctx context.Context, dumpWriter writer.DumperWriter) *ClusterRoleIngestor {
	return &ClusterRoleIngestor{
		buffer: make(map[string]*rbacv1.ClusterRoleList),
		writer: dumpWriter,
	}
}

func (d *ClusterRoleIngestor) IngestClusterRole(ctx context.Context, clusterRole types.ClusterRoleType) error {
	if ok, err := preflight.CheckClusterRole(clusterRole); !ok {
		return err
	}

	clusterRolePath := collector.ClusterRolesPath

	return bufferObject[rbacv1.ClusterRoleList, types.ClusterRoleType](ctx, clusterRolePath, d.buffer, clusterRole)
}

// Complete() is invoked by the collector when all k8s assets have been streamed.
// The function flushes all writers and waits for completion.
func (d *ClusterRoleIngestor) Complete(ctx context.Context) error {
	return dumpObj[*rbacv1.ClusterRoleList](ctx, d.buffer, d.writer)
}
