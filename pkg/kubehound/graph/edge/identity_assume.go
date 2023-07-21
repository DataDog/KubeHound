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

const (
	IdentityAssumeLabel = "IDENTITY_ASSUME"
)

func init() {
	Register(&IdentityAssume{}, RegisterDefault)
}

type IdentityAssume struct {
	BaseEdge
}

type identityGroup struct {
	Container primitive.ObjectID `bson:"container_id" json:"container"`
	Identity  primitive.ObjectID `bson:"identity_id" json:"identity"`
}

func (e *IdentityAssume) Label() string {
	return IdentityAssumeLabel
}

func (e *IdentityAssume) Name() string {
	return "IdentityAssume"
}

func (e *IdentityAssume) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*identityGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Container, typed.Identity)
}

func (e *IdentityAssume) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
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
