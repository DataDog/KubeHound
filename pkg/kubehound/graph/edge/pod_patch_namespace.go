package edge

import (
	"context"
	"os"

	"github.com/DataDog/KubeHound/pkg/globals"
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
	if _, ok := os.LookupEnv(globals.SwitchNamespacedNodes); ok {
		// Only required if nodes are namespaced. Otherwise will create overly large queries and cause timeouts
		Register(PodPatchNamespace{})
	}
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
	return BatchSizeClusterImpact
}

func (e PodPatchNamespace) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*podPatchGroup](ctx, entry)
}

// Traversal expects a list of podPatchGroup serialized as mapstructure for injection into the graph.
// For each podPatchGroup, the traversal will: 1) find the role vertex with matching storeID, 2) find the
// node vertices for each matching storeID, and 3) add a POD_PATCH edge between the vertices.
func (e PodPatchNamespace) Traversal() Traversal {
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
			Has("class", vertex.NodeLabel).
			Unfold().
			AddE(e.Label()).
			From(__.Select("r")).
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
				"as":   "nodesInNamespace",
				"from": "nodes",
				"let": bson.M{
					"roleNamespace": "$namespace",
				},
				"pipeline": []bson.M{
					{
						"$match": bson.M{"$or": bson.A{
							bson.M{"$expr": bson.M{
								"$eq": bson.A{
									"$k8.objectmeta.namespace", "$$roleNamespace",
								},
							}},
							bson.M{"is_namespaced": false},
						}},
					},
					{
						"$project": bson.M{
							"_id": 1,
						},
					},
				},
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
