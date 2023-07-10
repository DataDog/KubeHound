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
	Register(&RoleGrant{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2880471602/ROLE+GRANT
type RoleGrant struct {
	cfg *config.EdgeBuilderConfig
}

type roleGrantGroup struct {
	Role     primitive.ObjectID `bson:"role_id" json:"role"`
	Identity primitive.ObjectID `bson:"identity_id" json:"identity"`
}

func (e *RoleGrant) Initialize(cfg *config.EdgeBuilderConfig) error {
	e.cfg = cfg
	return nil
}

func (e *RoleGrant) Label() string {
	return "ROLE_GRANT"
}

func (e *RoleGrant) Name() string {
	return "RoleGrant"
}

func (e *RoleGrant) BatchSize() int {
	return e.cfg.BatchSize
}

func (e *RoleGrant) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*roleGrantGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Identity, typed.Role)
}

func (e *RoleGrant) Traversal() types.EdgeTraversal {
	return adapter.DefaultEdgeTraversal()
}

func (e *RoleGrant) Stream(ctx context.Context, store storedb.Provider, c cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	roleBindings := adapter.MongoDB(store).Collection(collections.RoleBindingName)

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

	cur, err := roleBindings.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[roleGrantGroup](ctx, cur, callback, complete)
}
