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
	Register(PodPatch{})
}

// @@DOCLINK: TODO
type PodPatch struct {
}

type podPatchGroup struct {
	Role primitive.ObjectID `bson:"_id" json:"role"`
}

func (e PodPatch) Label() string {
	return "POD_PATCH"
}

func (e PodPatch) Name() string {
	return "PodPatch"
}

func (e PodPatch) BatchSize() int {
	return BatchSizeClusterImpact
}

func (e PodPatch) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*podPatchGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	rid, err := oic.GraphID(ctx, typed.Role.Hex())
	if err != nil {
		return nil, fmt.Errorf("%s edge role id convert: %w", e.Label(), err)
	}

	processed := map[any]any{
		gremlin.T.Label: vertex.RoleLabel,
		gremlin.T.Id:    rid,
	}

	return processed, nil
}

func (e PodPatch) Traversal() types.EdgeTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().
			As("rpp").
			MergeV(__.Select("rpp")).
			Option(gremlin.Merge.OnCreate, __.Fail("missing role vertex on POD_PATCH insert")).
			Has("critical", false).
			As("r").
			V().
			HasLabel("Node").
			AddE(e.Label()).
			From(__.Select("r")).
			Barrier().Limit(0)

		return g
	}
}

// Stream finds all roles that have pod/patch or equivalent wildcard permissions.
func (e PodPatch) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	roles := adapter.MongoDB(store).Collection(collections.RoleName)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"rules": bson.M{
					"$elemMatch": bson.M{
						"$and": bson.A{
							bson.M{"$or": bson.A{
								bson.M{"resources": "pods"},
								bson.M{"resources": "pods/*"},
								bson.M{"resources": "*"},
							}},
							bson.M{"$or": bson.A{
								bson.M{"verbs": "patch"},
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

	return adapter.MongoCursorHandler[podPatchGroup](ctx, cur, callback, complete)
}
