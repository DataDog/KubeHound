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
	podPatchBatchSize = 5
)

func init() {
	Register(PodPatchCluster{})
}

// @@DOCLINK: TODO
type PodPatchCluster struct {
}

type podPatchClusterGroup struct {
	Role primitive.ObjectID `bson:"_id" json:"role"`
}

func (e PodPatchCluster) Label() string {
	return "POD_PATCH"
}

func (e PodPatchCluster) Name() string {
	return "PodPatchCluster"
}

func (e PodPatchCluster) BatchSize() int {
	return podPatchBatchSize
}

func (e PodPatchCluster) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*podPatchClusterGroup](ctx, entry)
}

// Traversal expects a list of podPatchClusterGroup serialized as mapstructure for injection into the graph.
// For each podPatchClusterGroup, the traversal will: 1) find the role vertex with matching storeID, 2) find ALL
// matching nodes in the cluster 3) add a POD_PATCH edge between the vertices.
func (e PodPatchCluster) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("ppc").
			V().
			HasLabel(vertex.RoleLabel).
			Has("critical", false). // Not out edges from critical assets
			Has("storeID", __.Where(P.Eq("ppc")).By().By("role")).
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

// Stream finds all roles that are NOT namespaced and have pod/patch or equivalent wildcard permissions.
func (e PodPatchCluster) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
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
								bson.M{"verbs": "patch"},
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

	return adapter.MongoCursorHandler[podPatchClusterGroup](ctx, cur, callback, complete)
}
