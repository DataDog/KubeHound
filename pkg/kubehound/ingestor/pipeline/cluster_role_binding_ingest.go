package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
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
		WithConverterCache(),
		WithStoreWriter(i.identity),
		WithStoreWriter(i.rolebinding),
		WithGraphWriter(i.vertex))
	if err != nil {
		return err
	}

	return nil
}

// processSubject will handle the ingestion pipeline for a role binding subject belonging to a processed K8s role binding input.
func (i *ClusterRoleBindingIngest) processSubject(ctx context.Context, subj *store.BindSubject) error {
	// Normalize K8s bind subject to store identity object format
	sid, err := i.r.storeConvert.Identity(ctx, subj)
	if err != nil {
		return err
	}

	// Async write identity to store
	if err := i.r.storeWriter(i.identity).Queue(ctx, sid); err != nil {
		return err
	}

	// Transform store model to vertex input
	v, err := i.r.graphConvert.Identity(sid)
	if err != nil {
		return err
	}

	// Aysnc write to graph
	if err := i.r.graphWriter(i.vertex).Queue(ctx, v); err != nil {
		return err
	}

	return nil
}

// streamCallback is invoked by the collector for each cluster role binding collected.
// The function ingests an input cluster role binding object into the store/graph and then ingests
// all child objects (identites, etc) through their own ingestion pipeline.
func (i *ClusterRoleBindingIngest) streamCallback(ctx context.Context, crb types.ClusterRoleBindingType) error {
	// Normalize K8s cluster role binding to store object format
	// TODO We can get cache misses here if bindings remain with no corresponding role which happens is staging!
	o, err := i.r.storeConvert.ClusterRoleBinding(ctx, crb)
	if err != nil {
		return err
	}

	// Async write role binding to store
	if err := i.r.storeWriter(i.rolebinding).Queue(ctx, o); err != nil {
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
func (i *ClusterRoleBindingIngest) completeCallback(ctx context.Context) error {
	return i.r.flushWriters(ctx)
}

func (i *ClusterRoleBindingIngest) Run(ctx context.Context) error {
	return i.r.collect.StreamClusterRoleBindings(ctx, i.streamCallback, i.completeCallback)
}

func (i *ClusterRoleBindingIngest) Close(ctx context.Context) error {
	return i.r.cleanupAll(ctx)
}
