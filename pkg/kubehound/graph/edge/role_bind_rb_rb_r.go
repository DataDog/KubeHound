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
	RoleBindspaceName = "RoleBindRoleBindingbRoleBindingRole"
)

func init() {
	Register(&RoleBindRbRbR{}, RegisterDefault)
}

type RoleBindRbRbR struct {
	BaseEdge
}

type roleBindNameSpaceGroup struct {
	FromPerm primitive.ObjectID `bson:"_id" json:"from_permission_set"`
	ToPerm   primitive.ObjectID `bson:"permset" json:"to_permission_set"`
}

func (e *RoleBindRbRbR) Label() string {
	return RoleBindLabel
}

func (e *RoleBindRbRbR) Name() string {
	return RoleBindspaceName
}

func (e *RoleBindRbRbR) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueValidAccounts
}

func (e *RoleBindRbRbR) AttckTacticID() AttckTacticID {
	return AttckTacticPrivilegeEscalation
}

func (e *RoleBindRbRbR) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*roleBindNameSpaceGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.FromPerm, typed.ToPerm, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *RoleBindRbRbR) Stream(ctx context.Context, store storedb.Provider, c cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	permissionSets := adapter.MongoDB(ctx, store).Collection(collections.PermissionSetName)
	pipeline := bson.A{
		bson.M{
			"$match": bson.M{
				// looking for RB CR/R role only
				"is_namespaced":   true,
				"runtime.runID":   e.runtime.RunID.String(),
				"runtime.cluster": e.runtime.ClusterName,
				"$and": []bson.M{
					// Looking for RBAC objects
					{
						"rules": bson.M{
							"$elemMatch": bson.M{
								"$or": []bson.M{
									{"apigroups": "*"},
									{"apigroups": "rbac.authorization.k8s.io"},
								},
							},
						},
					},
					// Looking for creation of rolebindings
					{
						"rules": bson.M{
							"$elemMatch": bson.M{
								"$and": []bson.M{
									{
										"$or": []bson.M{
											{"verbs": "create"},
											{"verbs": "*"},
										},
									},
									{
										"$or": []bson.M{
											{"resources": "rolebindings"},
											{"resources": "*"},
										},
									},
									{
										"$or": []bson.M{
											{"resourcenames": nil},
										},
									},
								},
							},
						},
					},
					// Looking for binding roles
					{
						"rules": bson.M{
							"$elemMatch": bson.M{
								"$and": []bson.M{
									{
										"$or": []bson.M{
											{"verbs": "bind"},
											{"verbs": "*"},
										},
									},
									{
										"$or": []bson.M{
											{"resources": "roles"},
											{"resources": "*"},
										},
									},
									{
										"$or": []bson.M{
											{"resourcenames": nil},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		// Looking for all permissionSets link to the same namespace
		bson.M{
			"$lookup": bson.M{
				"as":   "linkpermset",
				"from": "permissionsets",
				"let": bson.M{
					"roleNamespace": "$namespace",
				},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$or": bson.A{
								bson.M{"$expr": bson.M{
									"$eq": bson.A{
										"$k8.objectmeta.namespace", "$$roleNamespace",
									},
								}},
								bson.M{"is_namespaced": true},
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
			"$project": bson.M{
				"_id":     1,
				"permset": "$linkpermset._id",
			},
		},
	}
	cur, err := permissionSets.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[roleBindNameSpaceGroup](ctx, cur, callback, complete)
}
