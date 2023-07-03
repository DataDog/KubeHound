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

func (e IdentityAssume) Name() string {
	return "IdentityAssume"
}

func (e IdentityAssume) BatchSize() int {
	return BatchSizeDefault
}

func (e IdentityAssume) Processor(ctx context.Context, oic *converter.ObjectIdConverter, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*identityGroup](ctx, entry)
}

// Traversal expects a list of identityGroup serialized as mapstructure for injection into the graph.
// For each identityGroup, the traversal will: 1) find the container with matching storeID, 2) find the
// identity vertex with matching storeID, and 3) add a IDENTITY_ASSUME edge between the two vertices.
func (e IdentityAssume) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("ig").
			V().
			HasLabel(vertex.ContainerLabel).
			Has("class", vertex.ContainerLabel).
			Has("storeID", __.Where(P.Eq("ig")).By().By("container")).
			AddE(e.Label()).
			To(
				__.V().
					HasLabel(vertex.IdentityLabel).
					Has("class", vertex.IdentityLabel).
					Has("storeID", __.Where(P.Eq("ig")).By().By("identity"))).
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
