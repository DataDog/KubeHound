package edge

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	Register(shareProcessNamespace{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2880275294/SHARED+PS+NAMESPACE
type shareProcessNamespace struct {
}

// We want to map all the containers that shares the same process namespace
type shareProcessNamespaceContainers struct {
	Pod        primitive.ObjectID   `bson:"_id" json:"pod"`
	Containers []primitive.ObjectID `bson:"containers" json:"containers"`
}

func (e shareProcessNamespace) BatchSize() int {
	return DefaultBatchSize
}

func (e shareProcessNamespace) Label() string {
	return "SHARED_PS_NAMESPACE"
}

func (e shareProcessNamespace) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		return source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("sharedpns").
			V().HasLabel(vertex.PodLabel).
			Has("shareProcessNamespace", true).
			Has(
				"storeID", __.Select("sharedpns").Select("pod")).As("pod").
			V().HasLabel("Container").
			Has(
				"storeID", __.Select("sharedpns").Select("container")).As("container").
			MergeE(e.Label()).
			From("pod").
			To("container")
	}
}

func (e shareProcessNamespace) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*shareProcessNamespaceContainers](ctx, entry)
}

func (e shareProcessNamespace) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	containers := adapter.MongoDB(store).Collection(collections.PodName)
	pipeline := []bson.M{
		// find pods that have shareProcessNamespace set
		{
			"$match": bson.M{"shareProcessNamespace": true},
		},
		// Gather pods ID and their related containers
		{
			"$group": bson.M{
				"_id": "$pod_id",
				"containers": bson.M{
					"$push": "$_id",
				},
			},
		},
	}

	cur, err := containers.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[shareProcessNamespaceContainers](ctx, cur, callback, complete)
}
