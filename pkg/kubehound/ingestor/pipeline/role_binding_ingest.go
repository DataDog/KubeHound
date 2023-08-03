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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	RoleBindingIngestName = "k8s-role-binding-ingest"
)

type RoleBindingIngest struct {
	vertex        *vertex.Identity
	identity      collections.Identity
	rolebinding   collections.RoleBinding
	permissionset collections.PermissionSet
	r             *IngestResources
}

var _ ObjectIngest = (*RoleBindingIngest)(nil)

func (i *RoleBindingIngest) Name() string {
	return RoleBindingIngestName
}

func (i *RoleBindingIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	var err error

	i.vertex = &vertex.Identity{}
	i.identity = collections.Identity{}
	i.rolebinding = collections.RoleBinding{}
	i.permissionset = collections.PermissionSet{}

	i.r, err = CreateResources(ctx, deps,
		WithCacheWriter(cache.WithTest()),
		WithConverterCache(),
		WithStoreWriter(i.identity),
		WithStoreWriter(i.rolebinding),
		WithStoreWriter(i.permissionset),
		WithGraphWriter(i.vertex),
		WithCacheReader())
	if err != nil {
		return err
	}

	return nil
}

// processSubject will handle the ingestion pipeline for a role binding subject belonging to a processed K8s role binding input.
// We create identities via indirectly accessing role binding subjects rather than direct access (e.g k get serviceAccounts -A -o json)
// as this is the only way to discover non-serviceaccount users. However, this can create duplicate entries so lookup in cache before
// writing to the store.
// See reference: https://stackoverflow.com/questions/69932281/kubectl-command-to-return-a-list-of-all-user-accounts-from-kubernetes
func (i *RoleBindingIngest) processSubject(ctx context.Context, subj *store.BindSubject, parent *store.RoleBinding) error {
	// Normalize K8s bind subject to store identity object format
	sid, err := i.r.storeConvert.Identity(ctx, subj, parent)
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

// createPermissions
// Stats around permission in our cluster for some table corner calculation
// gizmo.us1.staging.dog: rb:4491 / crb:676 / r:1374 / cr:721
// apm3.us1.prod.dog: rb: 1851 /crb:171 / r:525 / cr:191
// daffy.us1.prod.dog: rb:1504 / crb:196 / r:127 / cr:219

// Workflow
// The rolebindings are being processed after the roles (cf pipeline order, file:///pkg/kubehound/ingestor/pipeline_ingestor.go )
// First save into the cache the role and clusterroles
// 	* create a specific save func to dump the whole object
//  * do some calculation to estimate the size to make sure we dont blow up our RAM

// RBAC rules and limitation
// * Roles and RoleBindings must exist in the same namespace.
// * RoleBindings can exist in separate namespaces to Service Accounts.
// * RoleBindings can link ClusterRoles, but they only grant access to the namespace of the RoleBinding.
// * ClusterRoleBindings link accounts to ClusterRoles and grant access across all resources.
// * ClusterRoleBindings can not reference Roles.

func (i *RoleBindingIngest) createPermissionSet(ctx context.Context, rb types.RoleBindingType, rbid primitive.ObjectID) error {

	// Get Role from cache
	role, err := i.r.cacheReader.Get(ctx, cachekey.Role(rb.RoleRef.Name, rb.Namespace)).Role()
	if err != nil {
		return err
	}

	// Normalize K8s role binding to store object format
	o, err := i.r.storeConvert.PermissionSet(ctx, role, rbid)
	if err != nil {
		return err
	}

	// Roles and RoleBindings must exist in the same namespace.
	if rb.Namespace != role.Namespace {
		log.I.Warnf("The role namespace does not match the rolebinding: r::%s/cr::%s", rb.Namespace, role.Namespace)
		return nil
	}

	// RoleBindings can exist in separate namespaces to Service Accounts.
	// Will be threated in the ROLE_GRANT edege

	// RoleBindings can link ClusterRoles, but they only grant access to the namespace of the RoleBinding.
	if !role.IsNamespaced {
		log.I.Warnf("Cluster role detected, downgrading it:%t ", o.IsNamespaced)
		o.IsNamespaced = true
		o.Namespace = rb.Namespace
	}
	log.I.Warnf("wait what ? o.IsNamespaced:%t ", o.IsNamespaced)

	// Async write role binding to store
	if err := i.r.writeStore(ctx, i.permissionset, o); err != nil {
		return err
	}

	return nil
}

// streamCallback is invoked by the collector for each role binding collected.
// The function ingests an input role binding object into the store/graph and then ingests
// all child objects (identites, etc) through their own ingestion pipeline.
func (i *RoleBindingIngest) IngestRoleBinding(ctx context.Context, rb types.RoleBindingType) error {
	if ok, err := preflight.CheckRoleBinding(rb); !ok {
		return err
	}

	// Normalize K8s role binding to store object format
	o, err := i.r.storeConvert.RoleBinding(ctx, rb)
	if err != nil {
		if err == converter.ErrDanglingRoleBinding {
			log.I.Warnf("%s : %s", err.Error(), rb.Name)
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
		err := i.processSubject(ctx, &s, o)
		if err != nil {
			return err
		}
	}

	// Create permission from Rolebinding entry
	//i.createPermissionSet(ctx, rb, o.Id)

	return nil
}

// completeCallback is invoked by the collector when all roles have been streamed.
// The function flushes all writers and waits for completion.
func (i *RoleBindingIngest) Complete(ctx context.Context) error {
	return i.r.flushWriters(ctx)
}

func (i *RoleBindingIngest) Run(ctx context.Context) error {
	return i.r.collect.StreamRoleBindings(ctx, i)
}

func (i *RoleBindingIngest) Close(ctx context.Context) error {
	return i.r.cleanupAll(ctx)
}
