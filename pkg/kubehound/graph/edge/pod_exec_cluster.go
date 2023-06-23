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
	podExecBatchSize = 5
)

func init() {
	Register(PodExecCluster{})
}

// @@DOCLINK: TODO
type PodExecCluster struct {
}

type podExecClusterGroup struct {
	Role primitive.ObjectID `bson:"_id" json:"role"`
}

func (e PodExecCluster) Label() string {
	return "POD_EXEC"
}

func (e PodExecCluster) Name() string {
	return "PodExecCluster"
}

func (e PodExecCluster) BatchSize() int {
	return podExecBatchSize
}

func (e PodExecCluster) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*podExecClusterGroup](ctx, entry)
}

// For each role, attach to all pods in the cluster!
func (e PodExecCluster) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("pec").
			V().HasLabel(vertex.RoleLabel).
			Where(P.Eq("pec")).
			By("storeID").
			By("role").
			As("r").
			V().HasLabel(vertex.PodLabel).
			Has("class", vertex.PodLabel).
			Unfold().
			AddE(e.Label()).
			From(__.Select("r")).
			Barrier().Limit(0)

		return g
	}
}

// Stream finds all roles that are NOT namespaced and have pod/exec or equivalent wildcard permissions.
func (e PodExecCluster) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
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

	return adapter.MongoCursorHandler[podExecClusterGroup](ctx, cur, callback, complete)
}
