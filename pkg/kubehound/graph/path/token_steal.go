package path

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
	tokenStealLabel = "TOKEN_STEAL"
)

func init() {
	Register(TokenSteal{})
}

type TokenStealPath struct {
	token
}

type TokenSteal struct {
}

func (v TokenSteal) Label() string {
	return tokenStealLabel
}

func (v TokenSteal) BatchSize() int {
	return DefaultBatchSize
}

func (v TokenSteal) Traversal() PathTraversal {
	return nil
}

func (v TokenSteal) Stream(ctx context.Context, sdb storedb.Provider, cache cache.CacheReader,
	process types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	// Find all volumes with projected service account tokens. The mounts and source fields we need to match on a projected
	// service account token are all deeply nested arrays so matching on the naming convention is the simplest/fastest match
	volumes := adapter.MongoDB(sdb).Collection(collections.VolumeName)
	filter := bson.M{
		"source.volumesource.projected": bson.M{"$exists": true, "$ne": "null"},
		"source.name":                   bson.M{"$regex": "/^kube-api-access/"},
	}

	cur, err := volumes.Find(ctx, filter)
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
