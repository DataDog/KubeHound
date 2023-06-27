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
	Register(ContainerAttach{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2883354625/CONTAINER+ATTACH
type ContainerAttach struct {
}

type containerAttachGroup struct {
	Pod        primitive.ObjectID   `bson:"_id" json:"pod"`
	Containers []primitive.ObjectID `bson:"containers" json:"containers"`
}

func (e ContainerAttach) Label() string {
	return "CONTAINER_ATTACH"
}

func (e ContainerAttach) Name() string {
	return "ContainerAttacher"
}

func (e ContainerAttach) BatchSize() int {
	return DefaultBatchSize
}

func (e ContainerAttach) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*containerAttachGroup](ctx, entry)
}

// Traversal expects a list of containerAttachGroup serialized as mapstructure for injection into the graph.
// For each containerAttachGroup, the traversal will: 1) find the pod vertex with matching storeID, 2) find the
// container vertices with matching storeIDs, and 3) add a CONTAINER_ATTACH edge between the pod and container vertices.
func (e ContainerAttach) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("ca").
			Select("containers").
			Unfold().
			As("c").
			V().
			HasLabel(vertex.ContainerLabel).
			Has("storeID", __.Where(P.Eq("c"))).
			AddE(e.Label()).
			From(
				__.V().
					HasLabel(vertex.PodLabel).
					Has("storeID", __.Where(P.Eq("ca")).By().By("pod"))).
			Barrier().Limit(0)

		return g
	}
}

func (e ContainerAttach) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	containers := adapter.MongoDB(store).Collection(collections.ContainerName)
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id": "$pod_id",
			"containers": bson.M{
				"$push": "$_id",
			},
		},
		},
	}

	cur, err := containers.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[containerAttachGroup](ctx, cur, callback, complete)
}
