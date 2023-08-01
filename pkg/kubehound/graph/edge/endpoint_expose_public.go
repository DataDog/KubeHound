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
)

func init() {
	Register(&EndpointExposePublic{}, RegisterDefault)
}

type EndpointExposePublic struct {
	BaseEdge
}

type endpointExposeGroup struct {
	Endpoint  primitive.ObjectID `bson:"_id" json:"endpoint_id"`
	Container primitive.ObjectID `bson:"container_id" json:"container_id"`
}

func (e *EndpointExposePublic) Label() string {
	return "ENDPOINT_EXPOSE"
}

func (e *EndpointExposePublic) Name() string {
	return "EndpointExposePublic"
}

func (e *EndpointExposePublic) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*endpointExposeGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Endpoint, typed.Container)
}

func (e *EndpointExposePublic) Stream(ctx context.Context, store storedb.Provider, c cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	endpoints := adapter.MongoDB(store).Collection(collections.EndpointName)

	pipeline := []bson.M{
		// Match bindings that have at least one subject
		{
			"$match": bson.M{
				"subjects": bson.M{
					"$exists": true,
					"$ne":     bson.A{},
				},
			},
		},
		// Flatten the subjects set
		{
			"$unwind": "$subjects",
		},
		// Project a role id / identity id pair
		{
			"$project": bson.M{
				"_id":         0,
				"role_id":     1,
				"identity_id": "$subjects.identity_id",
			},
		},
	}

	cur, err := endpoints.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[endpointExposeGroup](ctx, cur, callback, complete)
}
