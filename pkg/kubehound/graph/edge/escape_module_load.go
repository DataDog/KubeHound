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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	Register(EscapeModuleLoad{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2890006884/CE+MODULE+LOAD
type EscapeModuleLoad struct {
}

type moduleLoadGroup struct {
	Node      primitive.ObjectID `bson:"node_id" json:"node"`
	Container primitive.ObjectID `bson:"_id" json:"container"`
}

func (e EscapeModuleLoad) Label() string {
	return "CE_MODULE_LOAD"
}

func (e EscapeModuleLoad) BatchSize() int {
	return DefaultBatchSize
}

func (e EscapeModuleLoad) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()

		for _, i := range inserts {
			ml := i.(*moduleLoadGroup)

			g = g.V().
				Has("class", vertex.ContainerLabel).
				Has("storeID", ml.Container.Hex()).
				As("container").
				V().
				Has("class", vertex.NodeLabel).
				Has("storeID", ml.Node.Hex()).
				As("node").
				AddE(e.Label()).
				From("container").
				To("node")
		}

		return g
	}
}

func (e EscapeModuleLoad) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	containers := adapter.MongoDB(store).Collection(collections.ContainerName)

	// Escape is possible with privileged containers or CAP_SYS_MODULE loaded explicitly
	filter := bson.M{
		"$or": bson.A{
			bson.M{"k8.securitycontext.privileged": true},
			bson.M{"k8.securitycontext.capabilities.add": "SYS_MODULE"},
		}}

	// We just need a 1:1 mapping of the node and container to create this edge
	projection := bson.M{"_id": 1, "node_id": 1}

	cur, err := containers.Find(context.Background(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[moduleLoadGroup](ctx, cur, callback, complete)
}
