package edge

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	Register(SharedProcessNamespace{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2880275294/SHARED+PS+NAMESPACE
type SharedProcessNamespace struct {
}

type SharedProcessNamespaceContainers struct {
	Pod        primitive.ObjectID   `bson:"_id" json:"pod"`
	Containers []primitive.ObjectID `bson:"containers" json:"containers"`
}

func (e SharedProcessNamespace) Label() string {
	return "SHARED_PS_NAMESPACE"
}

func (e SharedProcessNamespace) Traversal() EdgeTraversal {
	return func(g *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		return g.Inject(inserts).Unfold().As("ca").
			V().HasLabel("Pod").Has("SharedProcessNamespace", true).Has("storeId", gremlin.T__.Select("ca").Select("pod")).As("pod").
			V().HasLabel("Container").Has("storeId", gremlin.T__.Select("ca").Select("container")).As("container").
			MergeE(e.Label()).From("pod").To("container")
	}
}

func (e SharedProcessNamespace) Processor(ctx context.Context, entry DataContainer) (TraversalInput, error) {
	return MongoProcessor[*SharedProcessNamespaceContainers](ctx, entry)
}

func (e SharedProcessNamespace) Stream(ctx context.Context, store storedb.Provider,
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

	return MongoCursorHandler[SharedProcessNamespaceContainers](ctx, cur, callback, complete)
}
