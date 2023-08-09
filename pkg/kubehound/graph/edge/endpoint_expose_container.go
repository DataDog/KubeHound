package edge

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	Register(&EndpointExposeContainer{}, RegisterDefault)
}

type EndpointExposeContainer struct {
	BaseEdge
}

type containerEndpointGroup struct {
	Endpoint  primitive.ObjectID `bson:"_id" json:"endpoint_id"`
	Container primitive.ObjectID `bson:"container_id" json:"container_id"`
}

func (e *EndpointExposeContainer) Label() string {
	return "ENDPOINT_EXPOSE"
}

func (e *EndpointExposeContainer) Name() string {
	return "EndpointExposeContainer"
}

func (e *EndpointExposeContainer) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*containerEndpointGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Endpoint, typed.Container)
}

func (e *EndpointExposeContainer) Stream(ctx context.Context, store storedb.Provider, c cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	endpoints := adapter.MongoDB(store).Collection(collections.EndpointName)

	// Collect the endpoints with no associated slice. These are directly created from a container port in the
	// pod ingest pipeline and so already have an associated container ID we can use directly.
	filter := bson.M{
		"has_slice": false,
	}

	// We just need a 1:1 mapping of the (private) endpoint and container to create this edge
	projection := bson.M{"_id": 1, "container_id": 1}

	cur, err := endpoints.Find(context.Background(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[containerEndpointGroup](ctx, cur, callback, complete)
}
