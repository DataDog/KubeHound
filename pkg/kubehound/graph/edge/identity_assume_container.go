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
	Register(&IdentityAssumeContainer{}, RegisterDefault)
}

type IdentityAssumeContainer struct {
	BaseEdge
}

type containerIdentityGroup struct {
	Container primitive.ObjectID `bson:"container_id" json:"container"`
	Identity  primitive.ObjectID `bson:"identity_id" json:"identity"`
}

func (e *IdentityAssumeContainer) Label() string {
	return "IDENTITY_ASSUME"
}

func (e *IdentityAssumeContainer) Name() string {
	return "IdentityAssumeContainer"
}

func (e *IdentityAssumeContainer) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*containerIdentityGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Container, typed.Identity)
}

func (e *IdentityAssumeContainer) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	// containers := adapter.MongoDB(store).Collection(collections.ContainerName)
	// pipeline := bson.A{
	// 	bson.M{
	// 		"$lookup": bson.M{
	// 			"from":         collections.IdentityName,
	// 			"localField":   "inherited.service_account",
	// 			"foreignField": "name",
	// 			"as":           "identity",
	// 		},
	// 	},
	// 	bson.M{
	// 		"$match": bson.M{
	// 			"identity": bson.M{
	// 				"$ne": bson.A{},
	// 			},
	// 		},
	// 	},
	// 	bson.M{
	// 		"$project": bson.M{
	// 			"container_id": "$_id",
	// 			"identity_id":  bson.M{"$arrayElemAt": []interface{}{"$identity._id", 0}},
	// 			"_id":          0,
	// 		},
	// 	},
	// }
	// cur, err := containers.Aggregate(ctx, pipeline)
	// TODO also need to match namespace
	identities := adapter.MongoDB(store).Collection(collections.IdentityName)
	pipeline := bson.A{
		bson.M{
			"$lookup": bson.M{
				"from":         collections.ContainerName,
				"localField":   "name",
				"foreignField": "inherited.service_account",
				"as":           "idc",
			},
		},
		bson.M{
			"$unwind": "$idc",
		},
		bson.M{
			"$project": bson.M{
				"container_id": "$idc._id",
				"identity_id":  "$_id",
				"_id":          0,
			},
		},
	}
	cur, err := identities.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[containerIdentityGroup](ctx, cur, callback, complete)
}
