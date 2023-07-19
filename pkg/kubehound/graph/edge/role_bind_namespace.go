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
	return "ROLE_BIND"
}

func (e *RoleBindNamespace) Name() string {
	return "RoleBindNamespace"
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
				"$and": []bson.M{
					{
						"rules": bson.M{
							"$elemMatch": bson.M{
								"verbs":     "create",
								"resources": "rolebindings",
							},
						},
					},
					{
						"rules": bson.M{
							"$elemMatch": bson.M{
								"verbs":         "bind",
								"resourcenames": "*",
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
				"path":                       "$roleLinkedInNamespace",
				"preserveNullAndEmptyArrays": true,
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
