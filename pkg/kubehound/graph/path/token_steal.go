package path

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2891284481/TOKEN+STEAL
	tokenStealLabel = "TOKEN_STEAL"
)

var (
	tokenStealPathLabel = fmt.Sprintf("(%s)-[%s]->(%s)-[%s]->(%s)",
		vertex.VolumeLabel, tokenStealLabel, vertex.TokenLabel, edge.IdentityAssumeLabel, vertex.IdentityLabel)
)

func init() {
	Register(TokenSteal{})
}

type volumeQueryResult struct {
	Volume            store.Volume `bson:"volume" json:"volume"`
	PodNamespace      string       `bson:"namespace" json:"namespace"`
	PodServiceAccount string       `bson:"serviceaccount" json:"serviceaccount"`
}

type tokenStealPath struct {
	Vertex     *graph.Token `bson:"properties" json:"properties"`
	VolumeId   string       `bson:"volume" json:"volume"`
	IdentityId string       `bson:"identity" json:"identity"`
}

type TokenSteal struct {
}

func (p TokenSteal) Label() string {
	return tokenStealPathLabel
}

func (p TokenSteal) BatchSize() int {
	return DefaultBatchSize
}

func (p TokenSteal) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*tokenStealPath](ctx, entry)
}

func (p TokenSteal) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		// Create the path from the provided inputs
		// (Volume)-[TOKEN_STEAL]->(Token)-[IDENTITY_ASSUME]->(Identity)
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("ts").
			// Create a new token vertex
			AddV(vertex.TokenLabel).As("tv").
			Property("class", vertex.TokenLabel). // labels are not indexed - use a mirror property
			SideEffect(
				__.Select("ts").
					Select("properties").
					Unfold().As("kv").
					Select("tv").
					Property(
						__.Select("kv").By(Column.Keys),
						__.Select("kv").By(Column.Values))).
			// Create the IDENTITY_ASSUME edge between the new token and an existing identity
			AddE(edge.IdentityAssumeLabel).
			To(
				__.V().
					HasLabel(vertex.IdentityLabel).
					Has("storeID", __.Where(P.Eq("ts")).By().By("identity"))).
			// Create the TOKEN_STEAL edge between an existing volume and the new token
			AddE(tokenStealLabel).
			To(__.Select("tv")).
			From(
				__.V().
					HasLabel(vertex.VolumeLabel).
					Has("storeID", __.Where(P.Eq("ts")).By().By("volume"))).
			Barrier().Limit(0)

		return g
	}
}

func (p TokenSteal) Stream(ctx context.Context, sdb storedb.Provider, cache cache.CacheReader,
	process types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	volumes := adapter.MongoDB(sdb).Collection(collections.VolumeName)

	// Find all volumes with projected service account tokens. The mounts and source fields we need to match on a projected
	// service account token are all deeply nested arrays so matching on the naming convention is the simplest/fastest match
	filter := bson.M{
		"source.volumesource.projected": bson.M{"$exists": true, "$ne": "null"},
		"source.name":                   bson.M{"$regex": primitive.Regex{Pattern: "^kube-api-access"}},
	}

	// Find the volume and associated pod namespace and service account.
	pipeline := []bson.M{
		{
			"$match": filter,
		},
		{
			"$lookup": bson.M{
				"from":         "pods",
				"localField":   "pod_id",
				"foreignField": "_id",
				"as":           "pod",
			},
		},
		{
			"$project": bson.M{
				"namespace": bson.M{
					"$first": "$pod.k8.objectmeta.namespace",
				},
				"serviceaccount": bson.M{
					"$first": "$pod.k8.spec.serviceaccountname",
				},
				"volume": "$$ROOT",
			},
		},
		{
			"$project": bson.M{
				"volume.pod": 0,
				"_id":        0,
			},
		},
	}

	cur, err := volumes.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	convert := converter.NewGraph()
	var res volumeQueryResult
	for cur.Next(ctx) {
		err := cur.Decode(&res)
		if err != nil {
			return err
		}

		// Retrieve the associated identity store ID from the cache
		said, err := cache.Get(ctx, cachekey.Identity(res.PodServiceAccount, res.PodNamespace))
		if err != nil {
			// This is completely fine. Most pods will run under a default account with no permissions which we treat
			// as having no identity. As such we do not want to create a token vertex here!
			continue
		}

		// Convert to our graph vertex representation
		v, err := convert.Token(res.PodServiceAccount, res.PodNamespace, &res.Volume)
		if err != nil {
			return err
		}

		// Create the container that holds all the data required by the traversal function
		err = process(ctx, &tokenStealPath{
			Vertex:     v,
			VolumeId:   res.Volume.Id.Hex(),
			IdentityId: said,
		})
		if err != nil {
			return err
		}
	}

	return complete(ctx)
}
