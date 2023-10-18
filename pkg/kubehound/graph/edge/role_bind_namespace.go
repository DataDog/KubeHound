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
	RoleBindNamespaceLabel = "ROLE_BIND"
	RoleBindspaceName      = "RoleBindNamespace"
)

func init() {
	Register(&RoleBindNamespace{}, RegisterDefault)
}

type RoleBindNamespace struct {
	BaseEdge
}

type roleBindNameSpaceGroup struct {
	FromPerm primitive.ObjectID `bson:"_id" json:"from_permission_set"`
	ToPerm   primitive.ObjectID `bson:"permset" json:"to_permission_set"`
}

func (e *RoleBindNamespace) Label() string {
	return RoleBindNamespaceLabel
}

func (e *RoleBindNamespace) Name() string {
	return RoleBindspaceName
}

func (e *RoleBindNamespace) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*roleBindNameSpaceGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.FromPerm, typed.ToPerm)
}

func (e *RoleBindNamespace) Stream(ctx context.Context, store storedb.Provider, c cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	permissionSets := adapter.MongoDB(store).Collection(collections.PermissionSetName)
	pipeline := bson.A{
		bson.M{
			"$match": bson.M{
				"is_namespaced": true,
			},
		},
		// Gather rolebinding details associated to the PermissionSets
		bson.M{
			"$lookup": bson.M{
				"from":         "rolebindings",
				"localField":   "role_binding_id",
				"foreignField": "_id",
				"as":           "result",
			},
		},
		// Removing all rolebindings without subjects attached
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
		bson.M{
			"$unwind": bson.M{
				"path": "$result",
			},
		},
		// Unfold the subjets associated to the rolebindings
		bson.M{
			"$unwind": bson.M{
				"path": "$result.subjects",
			},
		},
		// Checking that the rolebindings is linked to a ServiceAccount
		// OR the namespace match the namespace of the subject match the one from the rolebinding
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
					},
				},
			},
		},
		// Looking for all permissionSets link to the same namespace
		bson.M{
			"$lookup": bson.M{
				"from":         "permissionsets",
				"localField":   "namespace",
				"foreignField": "namespace",
				"as":           "linkpermset",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$linkpermset",
			},
		},
		// Removing the reference of the current PermissionSet from the pointed PermissionSet
		bson.M{
			"$match": bson.M{
				"$expr": bson.M{
					"$ne": bson.A{
						"$linkpermset._id",
						"$_id",
					},
				},
			},
		},
		// Projecting current PermissionSet with the associated one
		bson.M{
			"$group": bson.M{
				"_id":     "$_id",
				"permset": bson.M{"$first": "$linkpermset._id"},
			},
		},
	}
	cur, err := permissionSets.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[roleBindNameSpaceGroup](ctx, cur, callback, complete)
}
