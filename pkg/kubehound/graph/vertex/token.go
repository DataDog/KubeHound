package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	tokenLabel = "Token"
)

func init() {
	Register(Token{})
}

type Token struct {
}

func (v Token) Label() string {
	return tokenLabel
}

func (v Token) BatchSize() int {
	return DefaultBatchSize
}

func (v Token) Traversal() VertexTraversal {
	return nil
}

func (v Token) Stream(ctx context.Context, sdb storedb.Provider, cache cache.CacheReader,
	process types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {
	// All volumes
	// wehere source.porojected != nil
	// AND source.projected.sources.seriveAccountToken  != nil
	// AND return
	// pod id
	// _id

	volumes := adapter.MongoDB(sdb).Collection(collections.VolumeName)
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id": "$pod_id",
			"containers": bson.M{
				"$push": "$_id",
			},
		},
		},
	}

	cur, err := volumes.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	convert := converter.NewGraph()
	var entry store.Volume
	for cur.Next(ctx) {
		err := cur.Decode(&entry)
		if err != nil {
			return err
		}

		// Retrieve the service account name for the pod from the cache
		sa, err := cache.Get(ctx, cachekey.PodIdentity(entry.PodId.Hex()))
		if err != nil {
			log.Trace(ctx).Errorf("cache miss pod identity: %w", err)
			continue
		}

		// Retrieve the associated identity store ID from the cache
		said, err := cache.Get(ctx, cachekey.Identity(sa))
		if err != nil {
			// This is completely fine. Most pods will run under a default account with no permissions which we treat
			// as having no identity. As such we do not want to create a token vertex here!
			continue
		}

		// Convert to our graph representation
		v, err := convert.Token(sa, said, &entry)
		if err != nil {
			return err
		}

		err = process(ctx, v)
		if err != nil {
			return err
		}
	}

	return complete(ctx)
}
