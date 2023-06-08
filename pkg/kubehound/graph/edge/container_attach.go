package edge

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	Register(ContainerAttach{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2883354625/CONTAINER+ATTACH
type ContainerAttach struct {
}

type ContainerAttachGroup struct {
	Pod        primitive.ObjectID   `bson:"_id" json:"pod"`
	Containers []primitive.ObjectID `bson:"containers" json:"containers"`
}

func (e ContainerAttach) Label() string {
	return "CONTAINER_ATTACH"
}

func (e ContainerAttach) BatchSize() int {
	return DefaultBatchSize
}

func (e ContainerAttach) Traversal() EdgeTraversal {
	return func(g *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		log.I.Errorf("CONVERT ME TO SOMETHING TYPED OTHERWISE THIS WILL BROKE")
		return g.Inject(inserts).Unfold().As("ca").
			V().HasLabel("Pod").Has("storeId", gremlin.T__.Select("ca").Select("pod")).As("pod").
			V().HasLabel("Container").Has("storeId", gremlin.T__.Select("ca").Select("container")).As("container").
			MergeE(e.Label()).From("pod").To("container")
	}
}

func (e ContainerAttach) Processor(ctx context.Context, entry DataContainer) (TraversalInput, error) {
	return MongoProcessor[*ContainerAttachGroup](ctx, entry)
}

func (e ContainerAttach) Stream(ctx context.Context, store storedb.Provider,
	callback ProcessEntryCallback, complete CompleteQueryCallback) error {

	containers := MongoDB(store).Collection(collections.ContainerName)
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

	return MongoCursorHandler[ContainerAttachGroup](ctx, cur, callback, complete)
}
