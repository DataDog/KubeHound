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
	Register(PodExec{})
}

// @@DOCLINK: TODO
type PodExec struct {
}

type podExecGroup struct {
	Role primitive.ObjectID `bson:"_id" json:"role"`
}

func (e PodExec) Label() string {
	return "POD_EXEC"
}

func (e PodExec) Name() string {
	return "PodExec"
}

func (e PodExec) BatchSize() int {
	return BatchSizeClusterImpact
}

func (e PodExec) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*podExecGroup](ctx, entry)
}

// Traversal expects a list of podExecGroup serialized as mapstructure for injection into the graph.
// For each podExecGroup, the traversal will: 1) find the role vertex with matching storeID, 2) find ALL
// matching pods in the cluster 3) add a POD_EXEC edge between the vertices.
func (e PodExec) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("pec").
			V().
			HasLabel(vertex.RoleLabel).
			Has("critical", false). // Not out edges from critical assets
			Has("storeID", __.Where(P.Eq("pec")).By().By("role")).
			As("r").
			V().
			HasLabel(vertex.PodLabel).
			Has("class", vertex.PodLabel).
			Unfold().
			AddE(e.Label()).
			From(__.Select("r")).
			Barrier().Limit(0)

		return g
	}
}

// Stream finds all roles that are NOT namespaced and have pod/exec or equivalent wildcard permissions.
func (e PodExec) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
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
								bson.M{"verbs": "exec"},
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

	return adapter.MongoCursorHandler[podExecGroup](ctx, cur, callback, complete)
}