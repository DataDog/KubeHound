package storedb

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IndexBuilder struct {
	db *mongo.Database
}

func NewIndexBuilder(db *mongo.Database) (*IndexBuilder, error) {
	return &IndexBuilder{
		db: db,
	}, nil
}

func (ib *IndexBuilder) BuildAll(ctx context.Context) error {
	if err := ib.buildContainerIndices(ctx); err != nil {
		return fmt.Errorf("build container indices: %w", err)
	}

	if err := ib.buildEndpointIndices(ctx); err != nil {
		return fmt.Errorf("build endpoint indices: %w", err)
	}

	if err := ib.buildContainerIndices(ctx); err != nil {
		return fmt.Errorf("build container indices: %w", err)
	}

	if err := ib.buildContainerIndices(ctx); err != nil {
		return fmt.Errorf("build container indices: %w", err)
	}

	if err := ib.buildContainerIndices(ctx); err != nil {
		return fmt.Errorf("build container indices: %w", err)
	}

	if err := ib.buildContainerIndices(ctx); err != nil {
		return fmt.Errorf("build container indices: %w", err)
	}

	if err := ib.buildContainerIndices(ctx); err != nil {
		return fmt.Errorf("build container indices: %w", err)
	}

	if err := ib.buildContainerIndices(ctx); err != nil {
		return fmt.Errorf("build container indices: %w", err)
	}

	return nil
}

func (ib *IndexBuilder) buildContainerIndices(ctx context.Context) error {
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
			Keys:    bson.M{"k8.securitycontext.privileged": 1},
			Options: options.Index().SetName("byPrivileged"),
		},
		{
			Keys:    bson.M{"k8.securitycontext.capabilities.add": 1},
			Options: options.Index().SetName("ByCapabilities"),
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
				{"inherited.namespace", 1},
				{"inherited.pod_name", 1},
				{"k8.ports", 1},
			},
			Options: options.Index().SetName("bySharedNode"),
		},
	}

	_, err := containers.Indexes().CreateMany(ctx, indices)

	return err
}
