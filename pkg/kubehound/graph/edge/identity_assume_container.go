package edge

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
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

	containers := adapter.MongoDB(ctx, store).Collection(collections.ContainerName)
	pipeline := bson.A{
		bson.M{
			"$match": bson.M{
				"runtime.runID":   e.runtime.RunID.String(),
				"runtime.cluster": e.runtime.ClusterName,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"as":   "idc",
				"from": collections.IdentityName,
				"let": bson.M{
					"idName":      "$inherited.service_account",
					"idNamespace": "$inherited.namespace",
				},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$and": bson.A{
								bson.M{"$expr": bson.M{
									"$eq": bson.A{
										"$name", "$$idName",
									},
								}},
								bson.M{"$expr": bson.M{
									"$eq": bson.A{
										"$namespace", "$$idNamespace",
									},
								}},
								bson.M{"type": shared.IdentityTypeSA},
							},
							"runtime.runID":   e.runtime.RunID.String(),
							"runtime.cluster": e.runtime.ClusterName,
						},
					},
					{
						"$project": bson.M{
							"_id": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": "$idc",
		},
		bson.M{
			"$project": bson.M{
				"container_id": "$_id",
				"identity_id":  "$idc._id",
				"_id":          0,
			},
		},
	}

	cur, err := containers.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[containerIdentityGroup](ctx, cur, callback, complete)
}
