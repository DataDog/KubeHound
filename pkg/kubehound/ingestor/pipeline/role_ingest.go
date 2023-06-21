package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
)

const (
	RoleIngestName = "k8s-role-ingest"
)

type RoleIngest struct {
	vertex     vertex.Role
	collection collections.Role
	r          *IngestResources
}

var _ ObjectIngest = (*RoleIngest)(nil)

func (i *RoleIngest) Name() string {
	return RoleIngestName
}

func (i *RoleIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	var err error

	i.vertex = vertex.Role{}
	i.collection = collections.Role{}
	i.r, err = CreateResources(ctx, deps,
		WithCacheWriter(),
		WithStoreWriter(i.collection),
		WithGraphWriter(i.vertex))

	if err != nil {
		return err
	}

	return nil
}

// streamCallback is invoked by the collector for each role collected.
// The function ingests an input role into the cache/store/graph databases asynchronously.
func (i *RoleIngest) IngestRole(ctx context.Context, role types.RoleType) error {
	// Normalize K8s role to store object format
	o, err := i.r.storeConvert.Role(ctx, role)
	if err != nil {
		return err
	}

	// Async write to store
	if err := i.r.writeStore(ctx, i.collection, o); err != nil {
		return err
	}

	// Async write to cache
	if err := i.r.writeCache(ctx, cachekey.Role(o.Name, o.Namespace), o.Id.Hex()); err != nil {
		return err
	}

	// Transform store model to vertex input
	insert, err := i.r.graphConvert.Role(o)
	if err != nil {
		return err
	}

	// Aysnc write to graph
	if err := i.r.writeVertex(ctx, i.vertex, insert); err != nil {
		return err
	}

	return nil
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
