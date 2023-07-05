package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

const (
	ClusterRoleBindingIngestName = "k8s-cluster-role-binding-ingest"
)

type ClusterRoleBindingIngest struct {
	vertex      vertex.Identity
	identity    collections.Identity
	rolebinding collections.RoleBinding
	r           *IngestResources
}

var _ ObjectIngest = (*ClusterRoleBindingIngest)(nil)

func (i *ClusterRoleBindingIngest) Name() string {
	return ClusterRoleBindingIngestName
}

func (i *ClusterRoleBindingIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	var err error

	i.vertex = vertex.Identity{}
	i.identity = collections.Identity{}
	i.rolebinding = collections.RoleBinding{}

	i.r, err = CreateResources(ctx, deps,
		WithCacheWriter(cache.WithTest()),
		WithConverterCache(),
		WithStoreWriter(i.identity),
		WithStoreWriter(i.rolebinding),
		WithGraphWriter(i.vertex))
	if err != nil {
		return err
	}

	return nil
}

// processSubject will handle the ingestion pipeline for a cluster role binding subject belonging to a processed K8s cluster role binding input.
// We create identities via indirectly accessing role binding subjects rather than direct access (e.g k get serviceAccounts -A -o json)
// as this is the only way to discover non-serviceaccount users. However, this can create duplicate entries so lookup in cache before
// writing to the store.
// See reference: https://stackoverflow.com/questions/69932281/kubectl-command-to-return-a-list-of-all-user-accounts-from-kubernetes
func (i *ClusterRoleBindingIngest) processSubject(ctx context.Context, subj *store.BindSubject) error {
	// Normalize K8s bind subject to store identity object format
	sid, err := i.r.storeConvert.Identity(ctx, subj)
	if err != nil {
		return err
	}

	// Async write to cache. If entry is already present skip further processing.
	ck := cachekey.Identity(sid.Name, sid.Namespace)
	err = i.r.writeCache(ctx, ck, sid.Id.Hex())
	switch err {
	case cache.ErrCacheEntryOverwrite:
		log.I.Debugf("identity cache entry %#v already exists, skipping inserts", ck)
		return nil
	case nil:
		// NOP
	default:
		return err
	}

	// Async write identity to store
	if err := i.r.writeStore(ctx, i.identity, sid); err != nil {
		return err
	}

	// Transform store model to vertex input
	insert, err := i.r.graphConvert.Identity(sid)
	if err != nil {
		return err
	}

	// Aysnc write to graph
	if err := i.r.writeVertex(ctx, i.vertex, insert); err != nil {
		return err
	}

	return nil
}

// streamCallback is invoked by the collector for each cluster role binding collected.
// The function ingests an input cluster role binding object into the store/graph and then ingests
// all child objects (identites, etc) through their own ingestion pipeline.
func (i *ClusterRoleBindingIngest) IngestClusterRoleBinding(ctx context.Context, crb types.ClusterRoleBindingType) error {
	if ok, err := preflight.CheckClusterRoleBinding(crb); !ok {
		return err
	}

	// Normalize K8s cluster role binding to store object format
	o, err := i.r.storeConvert.ClusterRoleBinding(ctx, crb)
	if err != nil {
		if err == converter.ErrDanglingRoleBinding {
			log.I.Debugf("%s : %s", err.Error(), crb.Name)
			return nil
		}

		return err
	}

	// Async write role binding to store
	if err := i.r.writeStore(ctx, i.rolebinding, o); err != nil {
		return err
	}

	// Rolebinding itself has no graph component. However, the role binding subjects must be processed and
	// included in the store & graph as identity objects/vertices.
	for _, subj := range o.Subjects {
		s := subj
		err := i.processSubject(ctx, &s)
		if err != nil {
			return err
		}
	}

	return nil
}

// completeCallback is invoked by the collector when all roles have been streamed.
// The function flushes all writers and waits for completion.
func (i *ClusterRoleBindingIngest) Complete(ctx context.Context) error {
	return i.r.flushWriters(ctx)
}

func (i *ClusterRoleBindingIngest) Run(ctx context.Context) error {
	return i.r.collect.StreamClusterRoleBindings(ctx, i)
}

func (i *ClusterRoleBindingIngest) Close(ctx context.Context) error {
	return i.r.cleanupAll(ctx)
}
