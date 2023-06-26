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

const (
	// Use a small batch size here as each role will generate a significant number of edges
	podCreateBatchSize = 5
)

func init() {
	Register(PodCreateCluster{})
}

// @@DOCLINK: TODO
type PodCreateCluster struct {
}

type podCreateClusterGroup struct {
	Role primitive.ObjectID `bson:"_id" json:"role"`
}

func (e PodCreateCluster) Label() string {
	return "POD_CREATE"
}

func (e PodCreateCluster) Name() string {
	return "PodCreateCluster"
}

func (e PodCreateCluster) BatchSize() int {
	return podCreateBatchSize
}

func (e PodCreateCluster) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*podCreateClusterGroup](ctx, entry)
}

// Traversal expects a list of podCreateClusterGroup serialized as mapstructure for injection into the graph.
// For each podCreateClusterGroup, the traversal will: 1) find the role vertex with matching storeID, 2) find ALL
// matching nodes in the cluster 3) add a POD_CREATE edge between the vertices.
func (e PodCreateCluster) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("pcc").
			V().HasLabel(vertex.RoleLabel).
			Has("critical", false). // Not out edges from critical assets
			Has("storeID", __.Where(P.Eq("pcc")).By().By("role")).
			As("r").
			V().
			HasLabel(vertex.NodeLabel).
			Unfold().
			AddE(e.Label()).
			From(__.Select("r")).
			Barrier().Limit(0)

		return g
	}
}

// Stream finds all roles that are NOT namespaced and have pod/create or equivalent wildcard permissions.
func (e PodCreateCluster) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	roles := adapter.MongoDB(store).Collection(collections.RoleName)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"is_namespaced": false,
				"rules": bson.M{
					"$elemMatch": bson.M{
						"$and": bson.A{
							bson.M{"$or": bson.A{
								bson.M{"resources": "pods"},
								bson.M{"resources": "pods/*"},
								bson.M{"resources": "*"},
							}},
							bson.M{"$or": bson.A{
								bson.M{"verbs": "create"},
								bson.M{"verbs": "*"},
							}},
						},
					},
				},
			},
		},
		{
			"$project": bson.M{
				"_id": 1,
			},
		},
	}

	cur, err := roles.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[podCreateClusterGroup](ctx, cur, callback, complete)
}
