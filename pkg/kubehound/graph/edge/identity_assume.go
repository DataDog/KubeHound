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

	// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2880373371/IDENTITY+ASSUME
	IdentityAssumeLabel = "IDENTITY_ASSUME"
)

func init() {
	Register(IdentityAssume{})
}

type IdentityAssume struct {
}

type identityGroup struct {
	Container primitive.ObjectID `bson:"container_id" json:"container"`
	Identity  primitive.ObjectID `bson:"identity_id" json:"identity"`
}

func (e IdentityAssume) Label() string {
	return IdentityAssumeLabel
}

func (e IdentityAssume) BatchSize() int {
	return DefaultBatchSize
}

func (e IdentityAssume) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*identityGroup](ctx, entry)
}

func (e IdentityAssume) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("ig").
			V().HasLabel(vertex.ContainerLabel).
			Where(P.Eq("ig")).
			By("storeID").
			By("container").
			AddE(e.Label()).
			To(
				__.V().HasLabel(vertex.IdentityLabel).
					Where(P.Eq("ig")).
					By("storeID").
					By("identity")).
			Barrier().Limit(0)

		return g
	}
}

func (e IdentityAssume) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	containers := adapter.MongoDB(store).Collection(collections.ContainerName)
	pipeline := bson.A{
		bson.M{
			"$lookup": bson.M{
				"from":         collections.IdentityName,
				"localField":   "inherited.service_account",
				"foreignField": "name",
				"as":           "identity",
			},
		},
		bson.M{
			"$match": bson.M{
				"identity": bson.M{
					"$ne": bson.A{},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"container_id": "$_id",
				"identity_id":  bson.M{"$arrayElemAt": []interface{}{"$identity._id", 0}},
				"_id":          0,
			},
		},
	}
	cur, err := containers.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[identityGroup](ctx, cur, callback, complete)
}
