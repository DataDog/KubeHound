package pipeline

import (
	"context"
	"path"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	rbacv1 "k8s.io/api/rbac/v1"
)

type RoleBindingIngestor struct {
	buffer map[string]*rbacv1.RoleBindingList
	writer writer.DumperWriter
}

func ingestRoleBindingPath(roleBinding types.RoleBindingType) string {
	return path.Join(roleBinding.Namespace, collector.RoleBindingsPath)
}

func NewRoleBindingIngestor(ctx context.Context, dumpWriter writer.DumperWriter) *RoleBindingIngestor {
	return &RoleBindingIngestor{
		buffer: make(map[string]*rbacv1.RoleBindingList),
		writer: dumpWriter,
	}
}

func (d *RoleBindingIngestor) IngestRoleBinding(ctx context.Context, roleBinding types.RoleBindingType) error {
	if ok, err := preflight.CheckRoleBinding(roleBinding); !ok {
		return err
	}

	roleBindingPath := ingestRoleBindingPath(roleBinding)

	return bufferObject[rbacv1.RoleBindingList, types.RoleBindingType](ctx, roleBindingPath, d.buffer, roleBinding)
}

// Complete() is invoked by the collector when all k8s assets have been streamed.
// The function flushes all writers and waits for completion.
func (d *RoleBindingIngestor) Complete(ctx context.Context) error {
	return dumpObj[*rbacv1.RoleBindingList](ctx, d.buffer, d.writer)
}
