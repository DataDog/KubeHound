package storedb

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IndexBuilder handles the creation of indices for the store collections.
type IndexBuilder struct {
	db *mongo.Database
}

// NewIndexBuilder creates a new index builder instance for the provided DB.
func NewIndexBuilder(db *mongo.Database) (*IndexBuilder, error) {
	return &IndexBuilder{
		db: db,
	}, nil
}

// BuildAll builds all the store indices.
func (ib *IndexBuilder) BuildAll(ctx context.Context) error {
	if err := ib.containers(ctx); err != nil {
		return fmt.Errorf("build container indices: %w", err)
	}

	if err := ib.endpoints(ctx); err != nil {
		return fmt.Errorf("build endpoint indices: %w", err)
	}

	if err := ib.identities(ctx); err != nil {
		return fmt.Errorf("build identity indices: %w", err)
	}

	if err := ib.nodes(ctx); err != nil {
		return fmt.Errorf("build node indices: %w", err)
	}

	if err := ib.permissionsets(ctx); err != nil {
		return fmt.Errorf("build permission set indices: %w", err)
	}

	if err := ib.pods(ctx); err != nil {
		return fmt.Errorf("build pod indices: %w", err)
	}

	if err := ib.volumes(ctx); err != nil {
		return fmt.Errorf("build volume indices: %w", err)
	}

	return nil
}

// containers builds the store indices for the containers collection.
func (ib *IndexBuilder) containers(ctx context.Context) error {
	containers := ib.db.Collection(collections.ContainerName)
	indices := []mongo.IndexModel{
		{
			Keys:    bson.M{"pod_id": 1},
			Options: options.Index().SetName("byPod"),
		},
		{
			Keys:    bson.M{"node_id": 1},
			Options: options.Index().SetName("byNode"),
		},
		{
			Keys:    bson.M{"inherited.pod_name": 1},
			Options: options.Index().SetName("byPodName"),
		},
		{
			Keys: bson.D{
				{Key: "k8.securitycontext.privileged", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byPrivileged"),
		},
		{
			Keys: bson.D{
				{Key: "k8.securitycontext.capabilities.add", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("ByCapabilities"),
		},
		{
			Keys: bson.D{
				{Key: "inherited.host_pid", Value: 1},
				{Key: "k8.securitycontext.capabilities.add", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("BySharedPID"),
		},
		{
			Keys:    bson.M{"inherited.host_pid": 1},
			Options: options.Index().SetName("byHostPid"),
		},
		{
			Keys:    bson.M{"inherited.service_account": 1},
			Options: options.Index().SetName("bySA"),
		},
		{
			Keys:    bson.M{"inherited.namespace": 1},
			Options: options.Index().SetName("byNamespace"),
		},
		{
			Keys: bson.D{
				{Key: "inherited.namespace", Value: 1},
				{Key: "inherited.pod_name", Value: 1},
				{Key: "k8.ports", Value: 1},
			},
			Options: options.Index().SetName("bySharedNode"),
		},
		{
			Keys: bson.D{
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byRun"),
		},
		{
			Keys: bson.D{
				{Key: "k8.securitycontext.runasuser", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byRunAsUser"),
		},
	}

	_, err := containers.Indexes().CreateMany(ctx, indices)

	return err
}

// endpoints builds the store indices for the endpoints collection.
func (ib *IndexBuilder) endpoints(ctx context.Context) error {
	endpoints := ib.db.Collection(collections.EndpointName)
	indices := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "has_slice", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("bySliceSet"),
		},
		{
			Keys: bson.D{
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byRun"),
		},
	}

	_, err := endpoints.Indexes().CreateMany(ctx, indices)

	return err
}

// identities builds the store indices for the identities collection.
func (ib *IndexBuilder) identities(ctx context.Context) error {
	identities := ib.db.Collection(collections.IdentityName)
	indices := []mongo.IndexModel{
		{
			Keys:    bson.M{"namespace": 1},
			Options: options.Index().SetName("byNamespace"),
		},
		{
			Keys:    bson.M{"is_namespaced": 1},
			Options: options.Index().SetName("byNamespaceSet"),
		},
		{
			Keys:    bson.M{"type": 1},
			Options: options.Index().SetName("byType"),
		},
		{
			Keys:    bson.M{"name": 1},
			Options: options.Index().SetName("byName"),
		},
		{
			Keys: bson.D{
				{Key: "name", Value: 1},
				{Key: "namespace", Value: 1},
				{Key: "type", Value: 1},
			},
			Options: options.Index().SetName("byLookupFields"),
		},
		{
			Keys: bson.D{
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byRun"),
		},
	}

	_, err := identities.Indexes().CreateMany(ctx, indices)

	return err
}

// nodes builds the store indices for the nodes collection.
func (ib *IndexBuilder) nodes(ctx context.Context) error {
	nodes := ib.db.Collection(collections.NodeName)
	indices := []mongo.IndexModel{
		{

			Keys: bson.D{
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
				{Key: "user_id", Value: 1},
			},
			Options: options.Index().SetName("ByUserId"),
		},
		{
			Keys: bson.D{
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byRun"),
		},
	}

	_, err := nodes.Indexes().CreateMany(ctx, indices)

	return err
}

// permissionsets builds the store indices for the permissionsets collection.
func (ib *IndexBuilder) permissionsets(ctx context.Context) error {
	permissions := ib.db.Collection(collections.PermissionSetName)
	indices := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "is_namespaced", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byNamespaceSet"),
		},
		{
			Keys: bson.D{
				{Key: "namespace", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byNamespace"),
		},
		{
			Keys: bson.D{
				{Key: "rules.apigroups", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byApiGroup"),
		},
		{
			Keys: bson.D{
				{Key: "rules.resources", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byResources"),
		},
		{
			Keys: bson.D{
				{Key: "rules.verbs", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byVerbs"),
		},
		{
			Keys: bson.D{
				{Key: "rules.resourcenames", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byResourceNames"),
		},
		{
			Keys: bson.D{
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byRun"),
		},
	}

	_, err := permissions.Indexes().CreateMany(ctx, indices)

	return err
}

// pods builds the store indices for the pods collection.
func (ib *IndexBuilder) pods(ctx context.Context) error {
	pods := ib.db.Collection(collections.PodName)
	indices := []mongo.IndexModel{
		{
			Keys:    bson.M{"node_id": 1},
			Options: options.Index().SetName("byNode"),
		},
		{
			Keys:    bson.M{"is_namespaced": 1},
			Options: options.Index().SetName("byNamespaceSet"),
		},
		{
			Keys:    bson.M{"k8.objectmeta.namespace": 1},
			Options: options.Index().SetName("byNamespace"),
		},
		{
			Keys: bson.D{
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byRun"),
		},
		{
			Keys: bson.D{
				{Key: "k8.spec.shareprocessnamespace", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byShareProessNamespace"),
		},
	}

	_, err := pods.Indexes().CreateMany(ctx, indices)

	return err
}

// volumes builds the store indices for the volumes collection.
func (ib *IndexBuilder) volumes(ctx context.Context) error {
	volumes := ib.db.Collection(collections.VolumeName)
	indices := []mongo.IndexModel{
		{
			Keys:    bson.M{"pod_id": 1},
			Options: options.Index().SetName("byPod"),
		},
		{
			Keys:    bson.M{"node_id": 1},
			Options: options.Index().SetName("byNode"),
		},
		{
			Keys:    bson.M{"container_id": 1},
			Options: options.Index().SetName("byContainer"),
		},
		{
			Keys:    bson.M{"projected_id": 1},
			Options: options.Index().SetName("byProjected"),
		},
		{
			Keys:    bson.M{"type": 1},
			Options: options.Index().SetName("byType"),
		},
		{
			Keys:    bson.M{"source": 1},
			Options: options.Index().SetName("bySource"),
		},
		{
			Keys:    bson.M{"readonly": 1},
			Options: options.Index().SetName("byReadOnly"),
		},
		{
			Keys: bson.D{
				{Key: "node_id", Value: 1},
				{Key: "projected_id", Value: 1},
				{Key: "type", Value: 1},
			},
			Options: options.Index().SetName("bySharedNode"),
		},
		{
			Keys: bson.D{
				{Key: "source", Value: 1},
				{Key: "type", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byMountProperties"),
		},
		{
			Keys: bson.D{
				{Key: "source", Value: 1},
				{Key: "readonly", Value: 1},
				{Key: "type", Value: 1},
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byMountPropertiesEx"),
		},
		{
			Keys: bson.D{
				{Key: "runtime.runID", Value: 1},
				{Key: "runtime.cluster.name", Value: 1},
			},
			Options: options.Index().SetName("byRun"),
		},
	}

	_, err := volumes.Indexes().CreateMany(ctx, indices)

	return err
}
