package edge

import (
	"context"

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
	Register(TokenBruteforceNamespace{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2887155994/TOKEN+BRUTEFORCE
type TokenBruteforceNamespace struct {
}

type tokenBruteforceGroup struct {
	Role       primitive.ObjectID   `bson:"_id" json:"role"`
	Identities []primitive.ObjectID `bson:"idsInNamespace" json:"identities"`
}

func (e TokenBruteforceNamespace) Label() string {
	return "TOKEN_BRUTEFORCE"
}

func (e TokenBruteforceNamespace) Name() string {
	return "TokenBruteforceNamespace"
}

func (e TokenBruteforceNamespace) BatchSize() int {
	return BatchSizeClusterImpact
}

func (e TokenBruteforceNamespace) Processor(ctx context.Context, oic *converter.ObjectIdConverter, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*tokenBruteforceGroup](ctx, entry)
}

// Traversal expects a list of tokenBruteforceGroup serialized as mapstructure for injection into the graph.
// For each tokenBruteforceGroup, the traversal will: 1) find the role vertex with matching storeID, 2) find the
// identity vertices for each matching storeID, and 3) add a TOKEN_BRUTEFORCE edge between the vertices.
func (e TokenBruteforceNamespace) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("tbg").
			Select("identities").
			Unfold().
			As("id").
			V().
			HasLabel(vertex.IdentityLabel).
			Has("class", vertex.IdentityLabel).
			Has("storeID", __.Where(P.Eq("id"))).
			As("i").
			V().
			HasLabel(vertex.RoleLabel).
			Has("critical", false). // No out edges from critical assets
			Has("storeID", __.Where(P.Eq("tbg")).By().By("role")).
			As("r").
			AddE(e.Label()).
			From(__.Select("r")).
			To(__.Select("i")).
			Barrier().Limit(0)

		return g
	}
}

// Stream finds all roles that are namespaced and have secrets/get or equivalent wildcard permissions and matching identities.
// Matching identities are defined as namespaced identities that share the role namespace or non-namespaced identities.
func (e TokenBruteforceNamespace) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	roles := adapter.MongoDB(store).Collection(collections.RoleName)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"is_namespaced": true,
				"rules": bson.M{
					"$elemMatch": bson.M{
						"$and": bson.A{
							bson.M{"$or": bson.A{
								bson.M{"resources": "secrets"},
								bson.M{"resources": "secrets/*"},
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
			"$lookup": bson.M{
				"as":   "idsInNamespace",
				"from": "identities",
				"let": bson.M{
					"roleNamespace": "$namespace",
				},
				"pipeline": []bson.M{
					{
						"$match": bson.M{"$and": bson.A{
							bson.M{"$or": bson.A{
								bson.M{"$expr": bson.M{
									"$eq": bson.A{
										"$namespace", "$$roleNamespace",
									},
								}},
								bson.M{"is_namespaced": false},
							}},
							bson.M{"$or": bson.A{
								bson.M{"type": "ServiceAccount"},
								bson.M{"type": "User"},
							}},
						}},
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
			"$project": bson.M{
				"_id":            1,
				"idsInNamespace": "$idsInNamespace._id",
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
