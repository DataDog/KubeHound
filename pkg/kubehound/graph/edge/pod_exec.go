package edge

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	Register(PodExec{})
}

// @@DOCLINK: TODO
type PodExec struct {
}

type podExecGroup struct {
	Role primitive.ObjectID `bson:"_id" json:"role"`
}

func (e PodExec) Label() string {
	return "POD_EXEC"
}

func (e PodExec) Name() string {
	return "PodExec"
}

func (e PodExec) BatchSize() int {
	return BatchSizeClusterImpact
}

func (e PodExec) Processor(ctx context.Context, oic *converter.ObjectIdConverter, entry any) (any, error) {
	typed, ok := entry.(*podExecGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	rid, err := oic.GraphId(ctx, typed.Role.Hex())
	if err != nil {
		return nil, fmt.Errorf("%s edge role id convert: %w", e.Label(), err)
	}

	processed := map[any]any{
		gremlin.T.Label: vertex.RoleLabel,
		gremlin.T.Id:    rid,
	}

	return processed, nil
}

func (e PodExec) Traversal() types.EdgeTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().
			As("rpe").
			MergeV(__.Select("rpe")).
			Option(gremlin.Merge.OnCreate, __.Fail("missing role vertex on POD_EXEC insert")).
			Has("critical", false). // No out edges from critical assets
			As("r").
			V().
			HasLabel("Pod").
			AddE(e.Label()).
			From(__.Select("r")).
			Barrier().Limit(0)

		return g
	}
}

// Stream finds all roles that are NOT namespaced and have pod/exec or equivalent wildcard permissions.
func (e PodExec) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	roles := adapter.MongoDB(store).Collection(collections.RoleName)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"is_namespaced": false,
				"rules": bson.M{
					"$elemMatch": bson.M{
						"$and": bson.A{
							bson.M{"$or": bson.A{
								bson.M{"resources": "pods"},
								bson.M{"resources": "pods/*"},
								bson.M{"resources": "*"},
							}},
							bson.M{"$or": bson.A{
								bson.M{"verbs": "exec"},
								bson.M{"verbs": "*"},
							}},
						},
					},
				},
			},
		},
		{
			"$project": bson.M{
				"_id": 1,
			},
		},
	}

	cur, err := roles.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[podExecGroup](ctx, cur, callback, complete)
}
