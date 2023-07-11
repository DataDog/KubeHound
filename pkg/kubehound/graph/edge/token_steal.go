package edge

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	Register(&TokenSteal{}, RegisterDefault)
}

type volumeQueryResult struct {
	Volume            store.Volume `bson:"volume" json:"volume"`
	PodNamespace      string       `bson:"namespace" json:"namespace"`
	PodServiceAccount string       `bson:"serviceaccount" json:"serviceaccount"`
}

type tokenStealGroup struct {
	Volume   primitive.ObjectID `bson:"volume" json:"volume"`
	Identity primitive.ObjectID `bson:"identity" json:"identity"`
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2891284481/TOKEN+STEAL
type TokenSteal struct {
	cfg *config.EdgeBuilderConfig
}

func (e *TokenSteal) Initialize(cfg *config.EdgeBuilderConfig) error {
	e.cfg = cfg
	return nil
}

func (e *TokenSteal) Label() string {
	return "TOKEN_STEAL"
}

func (e *TokenSteal) Name() string {
	return "TokenSteal"
}

func (e *TokenSteal) BatchSize() int {
	return e.cfg.BatchSize
}

func (e *TokenSteal) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*tokenStealGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Volume, typed.Identity)
}

func (e *TokenSteal) Traversal() types.EdgeTraversal {
	return adapter.DefaultEdgeTraversal()
}

func (e *TokenSteal) Stream(ctx context.Context, sdb storedb.Provider, c cache.CacheReader,
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
		said, err := c.Get(ctx, cachekey.Identity(res.PodServiceAccount, res.PodNamespace)).ObjectID()
		switch err {
		case nil:
			// We have a matching identity object in the store, create an edge.
		case cache.ErrNoEntry:
			// This is completely fine. Most pods will run under a default account with no permissions which we treat
			// as having no identity. As such we do not want to create a token vertex here!
			continue
		default:
			return err
		}

		// Create the container that holds all the data required by the traversal function
		err = process(ctx, &tokenStealGroup{
			Volume:   res.Volume.Id,
			Identity: said,
		})
		if err != nil {
			return err
		}
	}

	return complete(ctx)
}
