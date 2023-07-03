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
	// Register(TokenBruteforceCluster{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2887155994/TOKEN+BRUTEFORCE
type TokenBruteforceCluster struct {
}

type tokenBruteforceClusterGroup struct {
	Role primitive.ObjectID `bson:"_id" json:"role"`
}

func (e TokenBruteforceCluster) Label() string {
	return "TOKEN_BRUTEFORCE"
}

func (e TokenBruteforceCluster) Name() string {
	return "TokenBruteforceCluster"
}

func (e TokenBruteforceCluster) BatchSize() int {
	return BatchSizeClusterImpact
}

func (e TokenBruteforceCluster) Processor(ctx context.Context, oic *converter.ObjectIdConverter, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*tokenBruteforceClusterGroup](ctx, entry)
}

// Traversal expects a list of TokenBruteforceCluster serialized as mapstructure for injection into the graph.
// For each TokenBruteforceCluster, the traversal will: 1) find the role vertex with matching storeID, 2) find ALL
// matching identities in the cluster 3) add a TOKEN_BRUTEFORCE edge between the vertices.
func (e TokenBruteforceCluster) Traversal() types.EdgeTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("tbc").
			V().
			Has("critical", false). // Not out edges from critical assets
			HasLabel(vertex.RoleLabel).
			Has("class", vertex.RoleLabel).
			Has("storeID", __.Where(P.Eq("tbc")).By().By("role")).
			As("r").
			V().
			HasLabel(vertex.IdentityLabel).
			Has("class", vertex.IdentityLabel).
			Has("type", "ServiceAccount").
			Unfold().
			AddE(e.Label()).
			From(__.Select("r")).
			Barrier().Limit(0)

		return g
	}
}

// Stream finds all roles that are NOT namespaced and have secrets/get or equivalent wildcard permissions.
func (e TokenBruteforceCluster) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
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

	return adapter.MongoCursorHandler[tokenBruteforceClusterGroup](ctx, cur, callback, complete)
}
