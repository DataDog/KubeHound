package storedb

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	mongotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go.mongodb.org/mongo-driver/mongo"
)

var _ Provider = (*MongoProvider)(nil)

type MongoProvider struct {
	client *mongo.Client
	db     *mongo.Database
	tags   []string
}

func NewMongoProvider(ctx context.Context, url string, connectionTimeout time.Duration) (*MongoProvider, error) {
	opts := options.Client()
	opts.Monitor = mongotrace.NewMonitor()
	opts.ApplyURI(url + fmt.Sprintf("/?connectTimeoutMS=%d", connectionTimeout))

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	db := client.Database(MongoDatabaseName)

	return &MongoProvider{
		client: client,
		db:     db,
		tags:   []string{telemetry.TagTypeMongodb},
	}, nil
}

func buildIndices(ctx context.Context, db *mongo.Database) error {
	// Containers
	containers := db.Collection(collections.ContainerName)
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
	if err != nil {
		return err
	}

	// Pods
	pods := db.Collection(collections.PodName)
	indices = []mongo.IndexModel{
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
	}

	_, err = pods.Indexes().CreateMany(ctx, indices)
	if err != nil {
		return err
	}

	// Volumes
	volumes := db.Collection(collections.VolumeName)
	indices = []mongo.IndexModel{
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
		// Mixed index to support the EXPLOIT_HOST_TRAVERSE edge
		{
			Keys: bson.D{
				{"node_id", 1},
				{"projected_id", 1},
				{"type", 1},
			},
			Options: options.Index().SetName("bySharedNode"),
		},
		// Mixed index to support the EXPLOIT_HOST_* edge
		{
			Keys: bson.D{
				{"source", 1},
				{"readonly", 1},
				{"type", 1},
			},
			Options: options.Index().SetName("byMountProperties"),
		},
	}

	_, err = volumes.Indexes().CreateMany(ctx, indices)
	if err != nil {
		return err
	}

	// PermissionSets
	permissions := db.Collection(collections.PermissionSetName)
	indices = []mongo.IndexModel{
		// {
		// 	Keys:    bson.M{"role_id": 1},
		// 	Options: options.Index().SetName("byRoleId"),
		// },
		// {
		// 	Keys:    bson.M{"role_name": 1},
		// 	Options: options.Index().SetName("byRole"),
		// },
		// {
		// 	Keys:    bson.M{"role_binding_id": 1},
		// 	Options: options.Index().SetName("byRoleBindingId"),
		// },
		// {
		// 	Keys:    bson.M{"role_binding_name": 1},
		// 	Options: options.Index().SetName("byRoleBinding"),
		// },
		{
			Keys:    bson.M{"is_namespaced": 1},
			Options: options.Index().SetName("byNamespaceSet"),
		},
		{
			Keys:    bson.M{"namespace": 1},
			Options: options.Index().SetName("byNamespace"),
		},
		{
			Keys:    bson.M{"rules.apigroups": 1},
			Options: options.Index().SetName("byApiGroup"),
		},
		{
			Keys:    bson.M{"rules.resources": 1},
			Options: options.Index().SetName("byResources"),
		},
		{
			Keys:    bson.M{"rules.verbs": 1},
			Options: options.Index().SetName("byVerbs"),
		},
		{
			Keys:    bson.M{"rules.resourcenames": 1},
			Options: options.Index().SetName("byResourceNames"),
		},
		// {
		// 	Keys: bson.D{
		// 		{"rules.apigroups", 1},
		// 		{"rules.resources", 1},
		// 		{"rules.verbs", 1},
		// 		{"rules.resourcenames", 1},
		// 	},
		// 	Options: options.Index().SetName("bySinglePermission"),
		// },
	}

	_, err = permissions.Indexes().CreateMany(ctx, indices)
	if err != nil {
		return err
	}

	// Endpoints
	endpoints := db.Collection(collections.EndpointName)
	indices = []mongo.IndexModel{
		// {
		// 	Keys:    bson.M{"container_id": 1},
		// 	Options: options.Index().SetName("byContainer"),
		// },
		// {
		// 	Keys:    bson.M{"pod_name": 1},
		// 	Options: options.Index().SetName("byPodName"),
		// },
		// {
		// 	Keys:    bson.M{"pod_namespace": 1},
		// 	Options: options.Index().SetName("byPodNamespace"),
		// },
		{
			Keys:    bson.M{"has_slice": 1},
			Options: options.Index().SetName("bySliceSet"),
		},
		// {
		// 	Keys:    bson.M{"port": 1},
		// 	Options: options.Index().SetName("byPort"),
		// },
		// {
		// 	Keys:    bson.M{"exposure": 1},
		// 	Options: options.Index().SetName("byExposure"),
		// },
	}

	_, err = endpoints.Indexes().CreateMany(ctx, indices)
	if err != nil {
		return err
	}

	// Identities
	identities := db.Collection(collections.IdentityName)
	indices = []mongo.IndexModel{
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
	}

	_, err = identities.Indexes().CreateMany(ctx, indices)
	if err != nil {
		return err
	}

	// Nodes
	nodes := db.Collection(collections.NodeName)
	indices = []mongo.IndexModel{
		{
			Keys:    bson.M{"user_id": 1},
			Options: options.Index().SetName("ByUserId"),
		},
	}

	_, err = nodes.Indexes().CreateMany(ctx, indices)
	if err != nil {
		return err
	}

	return nil
}

func (mp *MongoProvider) Prepare(ctx context.Context) error {
	collections, err := mp.db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("listing mongo DB collections: %w", err)
	}

	for _, collectionName := range collections {
		err = mp.db.Collection(collectionName).Drop(ctx)
		if err != nil {
			return fmt.Errorf("deleting mongo DB collection %s: %w", collectionName, err)
		}
	}

	if err := buildIndices(ctx, mp.db); err != nil {
		return err
	}

	return nil
}

func (mp *MongoProvider) Raw() any {
	return mp.client
}

func (mp *MongoProvider) Name() string {
	return "MongoProvider"
}

func (mp *MongoProvider) HealthCheck(ctx context.Context) (bool, error) {
	err := mp.client.Ping(ctx, nil)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (mp *MongoProvider) Close(ctx context.Context) error {
	return mp.client.Disconnect(ctx)
}

func (mp *MongoProvider) BulkWriter(ctx context.Context, collection collections.Collection, opts ...WriterOption) (AsyncWriter, error) {
	writer := NewMongoAsyncWriter(ctx, mp, collection, opts...)

	return writer, nil
}
