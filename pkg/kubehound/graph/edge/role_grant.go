package edge

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	Register(RoleGrant{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2880471602/ROLE+GRANT
type RoleGrant struct {
}

type roleBindingGroup struct {
	Role       primitive.ObjectID   `bson:"role_id" json:"role"`
	Identities []primitive.ObjectID `bson:"identity_ids" json:"identities"`
}

func (e RoleGrant) Label() string {
	return "ROLE_GRANT"
}

func (e RoleGrant) Name() string {
	return "RoleGrant"
}

func (e RoleGrant) BatchSize() int {
	return DefaultBatchSize
}

func (e RoleGrant) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*roleBindingGroup](ctx, entry)
}

// Traversal expects a list of roleBindings objects serialized as map structures for injection into the graph.
// For each roleBindings, the traversal will: 1) find the container vertex with matching storeID, 2) find the
// identity vertex with matching storeID, and 3) add a ROLE_GRANT edge between the two vertices.
func (e RoleGrant) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("rb").
			Select("identities").
			Unfold().
			As("id").
			V().HasLabel(vertex.IdentityLabel).
			Where(P.Eq("id")).
			By("storeID").
			By().
			AddE(e.Label()).
			To(
				__.V().HasLabel(vertex.RoleLabel).
					Where(P.Eq("rb")).
					By("storeID").
					By("role")).
			Barrier().Limit(0)

		return g
	}
}

func (e RoleGrant) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	roleBindings := adapter.MongoDB(store).Collection(collections.RoleBindingName)
	pipeline := bson.A{
		bson.M{
			"$unwind": "$subjects",
		},
		bson.M{
			"$group": bson.M{
				"_id": "$role_id",
				"identity_ids": bson.M{
					"$addToSet": "$subjects.identity_id",
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"role_id":      "$_id",
				"identity_ids": 1,
				"_id":          0,
			},
		},
	}

	cur, err := roleBindings.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)
	// TODO not all identities exist! check cache first
	return adapter.MongoCursorHandler[roleBindingGroup](ctx, cur, callback, complete)
}
