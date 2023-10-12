package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
)

const (
	ClusterRoleIngestName = "k8s-cluster-role-ingest"
)

type ClusterRoleIngest struct {
	collection collections.Role
	r          *IngestResources
}

var _ ObjectIngest = (*ClusterRoleIngest)(nil)

func (i *ClusterRoleIngest) Name() string {
	return ClusterRoleIngestName
}

func (i *ClusterRoleIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	var err error

	i.collection = collections.Role{}

	i.r, err = CreateResources(ctx, deps,
		WithCacheWriter(),
		WithStoreWriter(i.collection))
	if err != nil {
		return err
	}

	return nil
}

// streamCallback is invoked by the collector for each cluster role collected.
// The function ingests an input cluster role into the cache/store/graph databases asynchronously.
func (i *ClusterRoleIngest) IngestClusterRole(ctx context.Context, role types.ClusterRoleType) error {
	if ok, err := preflight.CheckClusterRole(role); !ok {
		return err
	}

	// Normalize K8s cluster role to store object format. Cluster roles are treated as
	// role within our model (with IsNamespaced flag set to false).
	o, err := i.r.storeConvert.ClusterRole(ctx, role)
	if err != nil {
		return err
	}

	// Async write to store
	if err := i.r.writeStore(ctx, i.collection, o); err != nil {
		return err
	}

	// Async write to cache
	return i.r.writeCache(ctx, cachekey.Role(o.Name, o.Namespace), *o)
}

// completeCallback is invoked by the collector when all cluster roles have been streamed.
// The function flushes all writers and waits for completion.
func (i *ClusterRoleIngest) Complete(ctx context.Context) error {
	return i.r.flushWriters(ctx)
}

func (i *ClusterRoleIngest) Run(ctx context.Context) error {
	return i.r.collect.StreamClusterRoles(ctx, i)
}

func (i *ClusterRoleIngest) Close(ctx context.Context) error {
	return i.r.cleanupAll(ctx)
}
