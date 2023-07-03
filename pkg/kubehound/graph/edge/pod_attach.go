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
	//Register(PodAttach{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2880668080/POD+ATTACH
type PodAttach struct {
}

type podAttachGroup struct {
	Node primitive.ObjectID   `bson:"_id" json:"node"`
	Pods []primitive.ObjectID `bson:"pods" json:"pods"`
}

func (e PodAttach) Label() string {
	return "POD_ATTACH"
}

func (e PodAttach) Name() string {
	return "PodAttach"
}

func (e PodAttach) BatchSize() int {
	return BatchSizeSmall
}

func (e PodAttach) Processor(ctx context.Context, oic *converter.ObjectIdConverter, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*podAttachGroup](ctx, entry)
}

// Traversal expects a list of podAttachGroup serialized as mapstructure for injection into the graph.
// For each podAttachGroup, the traversal will: 1) find the pod vertex with matching storeID, 2) find the
// node vertex with matching storeID, and 3) add a POD_ATTACH edge between the two vertices.
func (e PodAttach) Traversal() types.EdgeTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("pa").
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
					HasLabel(vertex.NodeLabel).
					Has("class", vertex.NodeLabel).
					Has("storeID", __.Where(P.Eq("pa")).By().By("node"))).
			Barrier().Limit(0)

		return g
	}
}

func (e PodAttach) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	pods := adapter.MongoDB(store).Collection(collections.PodName)
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id": "$node_id",
			"pods": bson.M{
				"$push": "$_id",
			},
		},
		},
	}

	cur, err := pods.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[podAttachGroup](ctx, cur, callback, complete)
}
