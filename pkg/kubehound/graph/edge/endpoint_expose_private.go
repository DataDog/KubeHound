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
	Register(&EndpointExposePrivate{}, RegisterDefault)
}

type EndpointExposePrivate struct {
	BaseEdge
}

type privateEndpointGroup struct {
	Endpoint  primitive.ObjectID `bson:"_id" json:"endpoint_id"`
	Container primitive.ObjectID `bson:"container_id" json:"container_id"`
}

func (e *EndpointExposePrivate) Label() string {
	return "ENDPOINT_EXPOSE"
}

func (e *EndpointExposePrivate) Name() string {
	return "EndpointExposePrivate"
}

func (e *EndpointExposePrivate) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*privateEndpointGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Endpoint, typed.Container)
}

func (e *EndpointExposePrivate) Stream(ctx context.Context, store storedb.Provider, c cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	endpoints := adapter.MongoDB(store).Collection(collections.EndpointName)

	// Collect the endpoints with no associated slice. These are directly created from a container port in the
	// pod ingest pipeline and so alrerady have an associated container ID we can use directly.
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

	return adapter.MongoCursorHandler[privateEndpointGroup](ctx, cur, callback, complete)
}
