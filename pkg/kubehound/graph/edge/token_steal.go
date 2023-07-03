package edge

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	Register(TokenSteal{})
}

type volumeQueryResult struct {
	Volume            store.Volume `bson:"volume" json:"volume"`
	PodNamespace      string       `bson:"namespace" json:"namespace"`
	PodServiceAccount string       `bson:"serviceaccount" json:"serviceaccount"`
}

type tokenStealGroup struct {
	VolumeId   string `bson:"volume" json:"volume"`
	IdentityId string `bson:"identity" json:"identity"`
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2891284481/TOKEN+STEAL
type TokenSteal struct {
}

func (e TokenSteal) Label() string {
	return "TOKEN_STEAL"
}

func (e TokenSteal) Name() string {
	return "TokenSteal"
}

func (e TokenSteal) BatchSize() int {
	return BatchSizeDefault
}

func (e TokenSteal) Processor(ctx context.Context, oic *converter.ObjectIdConverter, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*tokenStealGroup](ctx, entry)
}

func (e TokenSteal) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("ts").
			V().
			HasLabel(vertex.IdentityLabel).
			Has("class", vertex.IdentityLabel).
			Has("storeID", __.Where(P.Eq("ts")).By().By("identity")).
			AddE(e.Label()).
			From(
				__.V().
					HasLabel(vertex.VolumeLabel).
					Has("class", vertex.VolumeLabel).
					Has("storeID", __.Where(P.Eq("ts")).By().By("volume"))).
			Barrier().Limit(0)

		return g
	}
}

func (e TokenSteal) Stream(ctx context.Context, sdb storedb.Provider, cache cache.CacheReader,
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

		// Create the container that holds all the data required by the traversal function
		err = process(ctx, &tokenStealGroup{
			VolumeId:   res.Volume.Id.Hex(),
			IdentityId: said,
		})
		if err != nil {
			return err
		}
	}

	return complete(ctx)
}
