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
	// Use a small batch size here as each role will generate a significant number of edges
	return 5
}

func (e PodPatchCluster) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*podPatchClusterGroup](ctx, entry)
}

func (e PodPatchCluster) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("ppc").
			V().HasLabel(vertex.RoleLabel).
			Where(P.Eq("ppc")).
			By("storeID").
			By("role").
			As("r").
			V().HasLabel(vertex.NodeLabel).
			Has("class", vertex.NodeLabel).
			Unfold().
			AddE(e.Label()).
			From(__.Select("r")).
			Barrier().Limit(0)

		return g
	}
}

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
