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
	Register(&RoleGrant{}, RegisterDefault)
}

type RoleGrant struct {
	BaseEdge
}

type roleGrantGroup struct {
	Role     primitive.ObjectID `bson:"_id" json:"role"`
	Identity primitive.ObjectID `bson:"identity_id" json:"identity"`
}

func (e *RoleGrant) Label() string {
	return "ROLE_GRANT"
}

func (e *RoleGrant) Name() string {
	return "RoleGrant"
}

func (e *RoleGrant) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*roleGrantGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Identity, typed.Role)
}

func (e *RoleGrant) Stream(ctx context.Context, store storedb.Provider, c cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	roleBindings := adapter.MongoDB(store).Collection(collections.PermissionSetName)

	pipeline := bson.A{
		bson.M{
			"$lookup": bson.M{
				"from":         "rolebindings",
				"localField":   "role_binding_id",
				"foreignField": "_id",
				"as":           "result",
			},
		},
		bson.M{
			"$match": bson.M{
				"result": bson.M{
					"$elemMatch": bson.M{
						"subjects": bson.M{
							"$exists": true,
							"$ne":     bson.A{},
						},
					},
				},
			},
		},
		bson.M{"$unwind": bson.M{"path": "$result"}},
		bson.M{"$unwind": bson.M{"path": "$result.subjects"}},
		bson.M{
			"$match": bson.M{
				"$expr": bson.M{
					"$or": bson.A{
						bson.M{
							"$and": bson.A{
								bson.M{
									"$eq": bson.A{
										"$namespace",
										"$result.subjects.subject.namespace",
									},
								},
								bson.M{
									"$eq": bson.A{
										"$is_namespaced",
										true,
									},
								},
							},
						},
						bson.M{
							"$eq": bson.A{
								"$result.subjects.subject.kind",
								"ServiceAccount",
							},
						},
						bson.M{
							"$eq": bson.A{
								"$is_namespaced",
								false,
							},
						},
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"_id":         1,
				"identity_id": "$result.subjects.identity_id",
			},
		},
	}

	cur, err := roleBindings.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[roleGrantGroup](ctx, cur, callback, complete)
}
