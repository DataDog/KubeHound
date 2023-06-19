package edge

import (
	"context"
	"encoding/json"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	Register(RoleGrant{})
}

type RoleGrant struct {
}

type roleBindingGroup struct {
	Role       primitive.ObjectID   `bson:"role_id" json:"role"`
	Identities []primitive.ObjectID `bson:"identity_ids" json:"identities"`
}

func (e RoleGrant) Label() string {
	return "ROLE_GRANT"
}

func (e RoleGrant) BatchSize() int {
	return DefaultBatchSize
}

func StructToMap(in interface{}) (map[string]any, error) {
	var res map[string]any

	tmp, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(tmp, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// g.inject([["role":"3", "identities":["6","7"]],["role":"1", "identities":["5","6","7"]]]).unfold().as("rb").select("identities").unfold().as("id").V().hasLabel("Identity").where(eq("id")).by("sid").by().as("i").addE("MEGA_TEST").to(V().hasLabel("Role").where(eq("rb")).by("sid").by("role"))

func (e RoleGrant) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		counter := 0
		insertions := make([]map[string]any, 0)
		for _, i := range inserts {
			rb := i.(*roleBindingGroup)
			counter += len(rb.Identities)

			sm, err := StructToMap(i)
			if err != nil {
				log.I.Error(err)
				return nil
			}
			if len(rb.Identities) > 1 {
				log.I.Infof("Found %d identities for role %#v", len(rb.Identities), sm)
			}
			insertions = append(insertions, sm)
		}

		g := source.GetGraphTraversal().
			Inject(insertions).
			Unfold().As("rb").
			Select("identities").
			Unfold().
			As("id").
			V().HasLabel(vertex.IdentityLabel).
			Where(P.Eq("id")).
			By("storeID").
			By().
			// As("i").
			AddE(e.Label()).
			To(
				__.V().HasLabel(vertex.RoleLabel).
					Where(P.Eq("rb")).
					By("storeID").
					By("role")).
			Barrier().Limit(0)

		log.I.Infof("Should be %d ROLE_GRANT edges", counter)
		// // g = g.Inject(insertions).
		// // 	Unfold().As("rb").
		// // 	V().As("r").HasLabel("Role").
		// // 	Has("StoreID", P.Eq("648cc77f0b6bfc5f3a9ad70b")).
		// // 	V().As("i").HasLabel("Identity").
		// // 	Has("storeID", P.Within("648cc7820b6bfc5f3a9ad794", "648cc7820b6bfc5f3a9ad74a")).
		// // 	AddE("ROLE_GRANT").
		// // 	From("i").
		// // 	To("r").
		// // 	Barrier().Limit(0)

		// for _, i := range inserts {
		// 	rb := i.(*roleBindingGroup)

		// 	g.
		// 		V().As("r").HasLabel("Role").
		// 		Has("storeID", P.Eq(rb.Role.Hex())).
		// 		V().As("i").HasLabel("Identity").
		// 		Has("storeID", P.Eq(rb.Identities[0].Hex())).
		// 		AddE("ROLE_GRANT").
		// 		From("i").
		// 		To("r")
		// }

		// g = g.Inject(insertions).
		// 	Unfold().As("rb").
		// 	V().
		// 	Has("class", vertex.RoleLabel).
		// 	Has("storeID", __.Select("rb").Select("role")).
		// 	As("role").
		// 	V().
		// 	Has("class", vertex.IdentityLabel).
		// 	// Has("storeID", P.Within(__.Select("rb").Select("identities"))).
		// 	As("identity").
		// 	Limit(1).
		// 	AddE(e.Label()).
		// 	From("identity").
		// 	To("role")

		// for _, i := range inserts {
		// 	rb := i.(*roleBindingGroup)

		// 	ids := make([]any, 0, len(rb.Identities))
		// 	for _, id := range rb.Identities {
		// 		ids = append(ids, id.Hex())
		// 	}

		// 	g = g.
		// 		V().
		// 		Has("class", vertex.RoleLabel).
		// 		Has("storeID", rb.Role.Hex()).
		// 		As("role").
		// 		V().
		// 		Has("class", vertex.IdentityLabel).
		// 		Has("storeID", ids[0]).
		// 		//Has("storeID", P.Within(ids...)).
		// 		As("identity").
		// 		AddE(e.Label()).
		// 		From("identity").
		// 		To("role")
		// }

		// Confirmed working!
		// g.V().as("r").hasLabel("Role").has("storeID", eq("648cc77f0b6bfc5f3a9ad70b")).V().as("i").hasLabel("Identity").has("storeID", within("648cc7820b6bfc5f3a9ad794","648cc7820b6bfc5f3a9ad74a")).addE("ROLE_GRANT").from("i").to("r").barrier().limit(0)

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

	return adapter.MongoCursorHandler[roleBindingGroup](ctx, cur, callback, complete)
}
