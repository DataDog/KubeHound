package pipeline

import (
	"context"
	"errors"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

const (
	RoleBindingIngestName = "k8s-role-binding-ingest"
)

type RoleBindingIngest struct {
	vertexIdentity      *vertex.Identity
	vertexPermissionSet *vertex.PermissionSet
	identity            collections.Identity
	rolebinding         collections.RoleBinding
	permissionset       collections.PermissionSet
	r                   *IngestResources
}

var _ ObjectIngest = (*RoleBindingIngest)(nil)

func (i *RoleBindingIngest) Name() string {
	return RoleBindingIngestName
}

func (i *RoleBindingIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	var err error

	i.vertexIdentity = &vertex.Identity{}
	i.vertexPermissionSet = &vertex.PermissionSet{}
	i.identity = collections.Identity{}
	i.rolebinding = collections.RoleBinding{}
	i.permissionset = collections.PermissionSet{}

	i.r, err = CreateResources(ctx, deps,
		WithConverterDB(),
		WithStoreWriter(i.identity),
		WithStoreWriter(i.rolebinding),
		WithStoreWriter(i.permissionset),
		WithGraphWriter(i.vertexIdentity),
		WithGraphWriter(i.vertexPermissionSet))
	if err != nil {
		return err
	}

	return nil
}

// processSubject will handle the ingestion pipeline for a role binding subject belonging to a processed K8s role binding input.
func (i *RoleBindingIngest) processSubject(ctx context.Context, subj *store.BindSubject, parent *store.RoleBinding) error {
	// Normalize K8s bind subject to store identity object format
	sid, err := i.r.storeConvert.Identity(ctx, subj, parent)
	if err != nil {
		return err
	}

	// Check if identity already exists in the database. If so, skip further processing.
	if i.r.identityExists(ctx, sid.Name, sid.Namespace) {
		log.Trace(ctx).Debugf("identity %s/%s already exists, skipping inserts", sid.Namespace, sid.Name)
		return nil
	}

	// Write identity to store
	if err := i.r.writeStore(ctx, i.identity, sid); err != nil {
		return err
	}

	// Transform store model to vertex input
	insert, err := i.r.graphConvert.Identity(sid) //nolint: contextcheck
	if err != nil {
		return err
	}

	// Write to graph
	return i.r.writeVertex(ctx, i.vertexIdentity, insert)
}

// createPermissionSet creates a permission set from an input store role binding.
func (i *RoleBindingIngest) createPermissionSet(ctx context.Context, rb *store.RoleBinding) error {
	// Normalize K8s role binding to store object format
	o, err := i.r.storeConvert.PermissionSet(ctx, rb)
	if err != nil {
		return err
	}

	// Write role binding to store
	if err := i.r.writeStore(ctx, i.permissionset, o); err != nil {
		return err
	}

	// Transform store model to vertex input
	insert, err := i.r.graphConvert.PermissionSet(o) //nolint: contextcheck
	if err != nil {
		return err
	}

	// Write to graph
	return i.r.writeVertex(ctx, i.vertexPermissionSet, insert)
}

// streamCallback is invoked by the collector for each role binding collected.
func (i *RoleBindingIngest) IngestRoleBinding(ctx context.Context, rb types.RoleBindingType) error {
	if ok, err := preflight.CheckRoleBinding(rb); !ok {
		return err
	}

	// Normalize K8s role binding to store object format
	o, err := i.r.storeConvert.RoleBinding(ctx, rb)
	if err != nil {
		if errors.Is(err, converter.ErrDanglingRoleBinding) {
			log.Trace(ctx).Debugf("Role binding dropped (%s::%s): %s", rb.Namespace, rb.Name, err.Error())

			return nil
		}

		return err
	}

	// Write role binding to store
	if err := i.r.writeStore(ctx, i.rolebinding, o); err != nil {
		return err
	}

	// Process subjects as identity objects/vertices
	for _, subj := range o.Subjects {
		s := subj
		err := i.processSubject(ctx, &s, o)
		if err != nil {
			return err
		}
	}

	// Create permission from Rolebinding entry
	err = i.createPermissionSet(ctx, o)
	switch {
	case err == nil:
		// NOP
	case errors.Is(err, converter.ErrRoleCacheMiss):
		fallthrough
	case errors.Is(err, converter.ErrRoleBindProperties):
		log.Trace(ctx).Debugf("Permission set dropped (%s::%s): %v", rb.Namespace, rb.Name, err)
	default:
		return err
	}

	return nil
}

// completeCallback is invoked by the collector when all roles have been streamed.
func (i *RoleBindingIngest) Complete(ctx context.Context) error {
	return i.r.flushWriters(ctx)
}

func (i *RoleBindingIngest) Run(ctx context.Context) error {
	return i.r.collect.StreamRoleBindings(ctx, i)
}

func (i *RoleBindingIngest) Close(ctx context.Context) error {
	return i.r.cleanupAll(ctx)
}
