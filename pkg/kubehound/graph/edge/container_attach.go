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
	return func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		log.I.Errorf("CONVERT ME TO SOMETHING TYPED OTHERWISE THIS WILL BREAK")

		g := source.GetGraphTraversal()
		g.AddV("dummy_entry")
		// for _, i := range inserts {
		// 	data := i.(*ContainerAttachGroup)
		// 	podId, err := data.Pod.MarshalJSON()
		// 	if err != nil {
		// 		log.I.Errorf("failed to get pod id: %v", err)
		// 	}

		// 	for _, container := range data.Containers {
		// 		containerID, err := container.MarshalJSON()
		// 		if err != nil {
		// 			log.I.Errorf("failed to get pod id: %v", err)
		// 		}
		// 		pod, err := g.V("Pod").Has("store_id", podId).Next()
		// 		containers, err := g.V("Container").Has("store_id", containerID).Next()
		// 		clist, err := containers.GetSlice()

		// 		for _, container := range *clist {
		// 			fmt.Printf("containers edge blabla %+v\n", container)
		// 			g = g.V(pod).AddE(e.Label()).To(container)
		// 		}
		// 	}
		// }
		return g
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
