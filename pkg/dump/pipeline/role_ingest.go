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

type RoleIngestor struct {
	buffer map[string]*rbacv1.RoleList
	writer writer.DumperWriter
}

func ingestRolePath(roleBinding types.RoleType) string {
	return path.Join(roleBinding.Namespace, collector.RolesPath)
}

func NewRoleIngestor(ctx context.Context, dumpWriter writer.DumperWriter) *RoleIngestor {
	return &RoleIngestor{
		buffer: make(map[string]*rbacv1.RoleList),
		writer: dumpWriter,
	}
}

func (d *RoleIngestor) IngestRole(ctx context.Context, role types.RoleType) error {
	if ok, err := preflight.CheckRole(role); !ok {
		return err
	}

	rolePath := ingestRolePath(role)

	return bufferObject[rbacv1.RoleList, types.RoleType](ctx, rolePath, d.buffer, role)
}

// Complete() is invoked by the collector when all k8s assets have been streamed.
// The function flushes all writers and waits for completion.
func (d *RoleIngestor) Complete(ctx context.Context) error {
	return dumpObj[*rbacv1.RoleList](ctx, d.buffer, d.writer)
}
