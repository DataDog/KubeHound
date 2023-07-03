package edge

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	// Register(PodExecNamespace{})
}

// @@DOCLINK: TODO
type PodExecNamespace struct {
}

type podExecNamespaceGroup struct {
	Role primitive.ObjectID   `bson:"_id" json:"role"`
	Pods []primitive.ObjectID `bson:"podsInNamespace" json:"pods"`
}

func (e PodExecNamespace) Label() string {
	return "POD_EXEC"
}

func (e PodExecNamespace) Name() string {
	return "PodExecNamespace"
}

func (e PodExecNamespace) BatchSize() int {
	return BatchSizeClusterImpact
}

func (e PodExecNamespace) Processor(ctx context.Context, oic *converter.ObjectIdConverter, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*podExecNamespaceGroup](ctx, entry)
}

// Traversal expects a list of podExecNamespaceGroup serialized as mapstructure for injection into the graph.
// For each podExecNamespaceGroup, the traversal will: 1) find the role vertex with matching storeID, 2) find the
// pod vertices for each matching storeID, and 3) add a POD_EXEC edge between the vertices.
func (e PodExecNamespace) Traversal() types.EdgeTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("peg").
			Select("pods").
			Unfold().
			As("p").
			V().
			HasLabel(vertex.PodLabel).
			Has("class", vertex.PodLabel).
			Has("storeID", __.Where(P.Eq("p"))).
			AddE(e.Label()).
			From(
				__.V().
					HasLabel(vertex.RoleLabel).
					Has("critical", false). // Not out edges from critical assets
					Has("storeID", __.Where(P.Eq("peg")).By().By("role"))).
			Barrier().Limit(0)

		return g
	}
}

// Stream finds all roles that are namespaced and have pod/exec or equivalent wildcard permissions and matching pods.
// Matching pods are defined as all pods that share the role namespace or non-namespaced pods.
func (e PodExecNamespace) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
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
								bson.M{"verbs": "exec"},
								bson.M{"verbs": "*"},
							}},
						},
					},
				},
			},
		},
		{
			"$lookup": bson.M{
				"as":   "podsInNamespace",
				"from": "pods",
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
				"_id":             1,
				"podsInNamespace": "$podsInNamespace._id",
			},
		},
	}

	cur, err := roles.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[podExecNamespaceGroup](ctx, cur, callback, complete)
}
