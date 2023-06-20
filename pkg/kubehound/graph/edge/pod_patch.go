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
	Register(PodPatch{})
}

type PodPatch struct {
}

type podPatchGroup struct {
	Role primitive.ObjectID   `bson:"_id" json:"role"`
	Pods []primitive.ObjectID `bson:"podsInNamespace" json:"pods"`
}

func (e PodPatch) Label() string {
	return "POD_PATCH"
}

func (e PodPatch) BatchSize() int {
	return DefaultBatchSize
}

func (e PodPatch) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*podPatchGroup](ctx, entry)
}

// TODO this only handles the same namespace case. What do we do if its all namespaces to avoid edge count blowing up?
func (e PodPatch) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("ppg").
			Select("pods").
			Unfold().
			As("p").
			V().HasLabel(vertex.PodLabel).
			Where(P.Eq("p")).
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

func (e PodPatch) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	roles := adapter.MongoDB(store).Collection(collections.RoleName)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"rules": bson.M{
					"$elemMatch": bson.M{
						"$or": bson.A{
							bson.M{"resources": "pods"},
							bson.M{"resources": "pods/*"},
							bson.M{"resources": "*"},
						},
						"verbs": "patch",
					},
				},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "pods",
				"localField":   "namespace",
				"foreignField": "k8.objectmeta.namespace",
				"as":           "podsInNamespace",
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

	return adapter.MongoCursorHandler[podPatchGroup](ctx, cur, callback, complete)
}
