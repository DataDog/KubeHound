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
	RoleBindNamespaceLabel = RoleBindLabel
	RoleBindNamespaceName  = "RoleBindNamespace"
)

func init() {
	Register(&RoleBindNamespace{}, RegisterDefault)
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2889355675/ROLE+BIND
type RoleBindNamespace struct {
	BaseEdge
}

type roleBindGroupNamespace struct {
	Role       primitive.ObjectID `bson:"_id" json:"role"`
	LinkedRole primitive.ObjectID `bson:"linkedRoleID" json:"linkedRole"`
}

func (e *RoleBindNamespace) Label() string {
	return RoleBindNamespaceLabel
}

func (e *RoleBindNamespace) Name() string {
	return RoleBindNamespaceName
}

func (e *RoleBindNamespace) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*roleBindGroupNamespace)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Role, typed.LinkedRole)
}

func (e *RoleBindNamespace) Stream(ctx context.Context, store storedb.Provider, c cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	roleBindings := adapter.MongoDB(store).Collection(collections.RoleName)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"is_namespaced": true,
				"rules": bson.M{
					"$elemMatch": bson.M{
						"$or": bson.A{
							bson.M{"apigroups": "*"},
							bson.M{"apigroups": "rbac.authorization.k8s.io"},
						},
					},
				},
				"$and": bson.A{
					bson.M{
						"rules": bson.M{
							"$elemMatch": bson.M{
								"$or": bson.A{
									bson.M{"verbs": "create"},
									bson.M{"verbs": "*"},
								},
							},
						},
					},
					bson.M{
						"rules": bson.M{
							"$elemMatch": bson.M{
								"$or": bson.A{
									bson.M{"verbs": "bind"},
									bson.M{"verbs": "*"},
								},
							},
						},
					},
					bson.M{
						"rules": bson.M{
							"$elemMatch": bson.M{
								"$or": bson.A{
									bson.M{"resources": "clusterrolebindings"},
									bson.M{"resources": "rolebindings"},
									bson.M{"verbs": "*"},
								},
							},
						},
					},
				},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "roles",
				"as":           "roleLinkedInNamespace",
				"localField":   "namespace",
				"foreignField": "namespace",
			},
		},
		{
			"$unwind": bson.M{
				"path": "$roleLinkedInNamespace",
			},
		},
		{
			"$project": bson.M{
				"_id":          1,
				"linkedRoleID": "$roleLinkedInNamespace._id",
			},
		},
	}
	cur, err := roleBindings.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[roleBindGroupNamespace](ctx, cur, callback, complete)
}
