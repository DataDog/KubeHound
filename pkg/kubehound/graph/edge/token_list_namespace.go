package edge

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
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
	Register(&TokenListNamespace{}, RegisterDefault)
}

type TokenListNamespace struct {
	BaseEdge
}

type tokenListNSGroup struct {
	Role     primitive.ObjectID `bson:"_id" json:"role"`
	Identity primitive.ObjectID `bson:"identity" json:"identity"`
}

func (e *TokenListNamespace) Label() string {
	return "TOKEN_LIST"
}

func (e *TokenListNamespace) Name() string {
	return "TokenListNamespace"
}

func (e *TokenListNamespace) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*tokenListNSGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Role, typed.Identity)
}

// Stream finds all roles that are namespaced and have secrets/list or equivalent wildcard permissions and matching identities.
// Matching identities are defined as namespaced identities that share the role namespace or non-namespaced identities.
func (e *TokenListNamespace) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader, runtime *config.DynamicConfig,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	permissionSets := adapter.MongoDB(store).Collection(collections.PermissionSetName)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"is_namespaced":   true,
				"runtime.runID":   runtime.RunID.String(),
				"runtime.cluster": runtime.ClusterName,
				"rules": bson.M{
					"$elemMatch": bson.M{
						"$and": bson.A{
							bson.M{"$or": bson.A{
								bson.M{"apigroups": ""},
								bson.M{"apigroups": "*"},
							}},
							bson.M{"$or": bson.A{
								bson.M{"resources": "secrets"},
								bson.M{"resources": "*"},
							}},
							bson.M{"$or": bson.A{
								bson.M{"verbs": "list"},
								bson.M{"verbs": "*"},
							}},
							bson.M{"resourcenames": nil}, // TODO: handle resource scope
						},
					},
				},
			},
		},
		{
			"$lookup": bson.M{
				"as":   "idsInNamespace",
				"from": "identities",
				"let": bson.M{
					"roleNamespace": "$namespace",
				},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$and": bson.A{
								bson.M{"$or": bson.A{
									bson.M{"$expr": bson.M{
										"$eq": bson.A{
											"$namespace", "$$roleNamespace",
										},
									}},
									bson.M{"is_namespaced": false},
								}},
								bson.M{"type": "ServiceAccount"},
							},
							"runtime.runID":   runtime.RunID.String(),
							"runtime.cluster": runtime.ClusterName,
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
		{
			"$unwind": "$idsInNamespace",
		},
		{
			"$project": bson.M{
				"_id":      1,
				"identity": "$idsInNamespace._id",
			},
		},
	}

	cur, err := permissionSets.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[tokenListNSGroup](ctx, cur, callback, complete)
}
