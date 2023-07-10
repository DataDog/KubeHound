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
	// TODO just mark critical if large cluster switch
	Register(TokenBruteforce{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2887155994/TOKEN+BRUTEFORCE
type TokenBruteforce struct {
}

type tokenBruteforceGroup struct {
	Role primitive.ObjectID `bson:"_id" json:"role"`
}

func (e TokenBruteforce) Label() string {
	return "TOKEN_BRUTEFORCE"
}

func (e TokenBruteforce) Name() string {
	return "TokenBruteforceCluster"
}

func (e TokenBruteforce) BatchSize() int {
	return BatchSizeClusterImpact
}

func (e TokenBruteforce) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*tokenBruteforceGroup)
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

func (e TokenBruteforce) Traversal() types.EdgeTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().
			As("rtb").
			MergeV(__.Select("rtb")).
			Option(gremlin.Merge.OnCreate, __.Fail("missing role vertex on TOKEN_BRUTEFORCE insert")).
			Has("critical", false). // No out edges from critical assets
			As("r").
			V().
			HasLabel("Identity").
			AddE(e.Label()).
			From(__.Select("r")).
			Barrier().Limit(0)

		return g
	}
}

// Stream finds all roles that are NOT namespaced and have secrets/get or equivalent wildcard permissions.
func (e TokenBruteforce) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
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
								bson.M{"apigroups": ""},
								bson.M{"apigroups": "*"},
							}},
							bson.M{"$or": bson.A{
								bson.M{"resources": "secrets"},
								bson.M{"resources": "*"},
							}},
							bson.M{"$or": bson.A{
								bson.M{"verbs": "get"},
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

	return adapter.MongoCursorHandler[tokenBruteforceGroup](ctx, cur, callback, complete)
}
