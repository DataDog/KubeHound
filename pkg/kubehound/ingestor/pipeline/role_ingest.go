package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
)

const (
	RoleIngestName = "k8s-role-ingest"
)

type RoleIngest struct {
	r *IngestResources
}

var _ ObjectIngest = (*RoleIngest)(nil)

func (i *RoleIngest) Name() string {
	return RoleIngestName
}

func (i *RoleIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	var err error

	i.r, err = CreateResources(ctx, deps)
	if err != nil {
		return err
	}

	return nil
}

// streamCallback is invoked by the collector for each role collected.
// The function ingests an input role into the store database.
func (i *RoleIngest) IngestRole(ctx context.Context, role types.RoleType) error {
	if ok, err := preflight.CheckRole(role); !ok {
		return err
	}

	// Normalize K8s role to store object format
	o, err := i.r.storeConvert.Role(ctx, role)
	if err != nil {
		return err
	}

	// Write to store
	return i.r.writeStore(ctx, o)
}

// completeCallback is invoked by the collector when all roles have been streamed.
// The function flushes all writers and waits for completion.
func (i *RoleIngest) Complete(ctx context.Context) error {
	return i.r.flushWriters(ctx)
}

func (i *RoleIngest) Run(ctx context.Context) error {
	return i.r.collect.StreamRoles(ctx, i)
}

func (i *RoleIngest) Close(ctx context.Context) error {
	return i.r.cleanupAll(ctx)
}
