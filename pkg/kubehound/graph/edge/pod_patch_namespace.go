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
	Register(PodPatchNamespace{})
}

// @@DOCLINK: TODO
type PodPatchNamespace struct {
}

type podPatchGroup struct {
	Role  primitive.ObjectID   `bson:"_id" json:"role"`
	Nodes []primitive.ObjectID `bson:"nodesInNamespace" json:"nodes"`
}

func (e PodPatchNamespace) Label() string {
	return "POD_PATCH"
}

func (e PodPatchNamespace) Name() string {
	return "PodPatchNamespace"
}

func (e PodPatchNamespace) BatchSize() int {
	return DefaultBatchSize
}

func (e PodPatchNamespace) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*podPatchGroup](ctx, entry)
}

// Traversal expects a list of podPatchGroup serialized as mapstructure for injection into the graph.
// For each podPatchGroup, the traversal will: 1) find the role vertex with matching storeID, 2) find the
// pod vertices for each matching storeID, and 3) add a POD_PATCH edge between the two vertices.
// TODO this only handles the same namespace case. What do we do if its all namespaces to avoid edge count blowing up?
func (e PodPatchNamespace) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("ppg").
			Select("nodes").
			Unfold().
			As("n").
			V().HasLabel(vertex.NodeLabel).
			Where(P.Eq("n")).
			By("storeID").
			By().
			AddE(e.Label()).
			From(
				__.V().HasLabel(vertex.RoleLabel).
					Where(P.Eq("ppg")).
					By("storeID").
					By("role")).
			Barrier().Limit(0)

		return g
	}
}

// Stream finds all roles that are namespaced and have pod/patch or equivalent wildcard permissions and matching nodes.
// Matching nodes are defined as namespaced nodes that share the role namespace or non-namespaced nodes.
func (e PodPatchNamespace) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	roles := adapter.MongoDB(store).Collection(collections.RoleName)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"is_namespaced": true,
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
			"$lookup": bson.M{
				"from": "nodes",
				"let": bson.M{
					"namespace": "$namespace",
				},
				"pipeline": []bson.M{
					{
						"$match": bson.M{"$or": bson.A{
							bson.M{"k8.objectmeta.namespace": "$$namespace"},
							bson.M{"is_namespaced": false},
						}},
					},
					{
						"$project": bson.M{
							"_id": 1,
						},
					},
				},
				"as": "nodesInNamespace",
			},
		},
		{
			"$project": bson.M{
				"_id":              1,
				"nodesInNamespace": "$nodesInNamespace._id",
			},
		},
	}

	cur, err := roles.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[podPatchGroup](ctx, cur, callback, complete)
}
