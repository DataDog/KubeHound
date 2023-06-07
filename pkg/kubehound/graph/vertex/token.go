package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	tokenLabel = "Token"
)

func init() {
	Register(Token{})
}

type TokenQueryGroup struct {
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

func (v Token) Stream(ctx context.Context, store storedb.Provider, cache cache.CacheReader,
	process types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	// TODO add identity - identity id cache!!
	// TODO add podid - identity cache
	// TODO check pod id has a real service aqccount (not the default with no permissions!!)
	// #
	All volumes
	wehere source.porojected != nil 
	AND source.projected.sources.seriveAccountToken  != nil
	AND lookup pod with pod_id == volume.pod id
	


	volumes := adapter.MongoDB(store).Collection(collections.VolumeName)
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

	var entry TokenQueryGroup
	for cur.Next(ctx) {
		err := cur.Decode(&entry)
		if err != nil {
			return err
		}

		// 	THen lookup pod_id -> identity_id cache for an entry
	    // if something then create a token!

		err = callback(ctx, &entry)
		if err != nil {
			return err
		}
	}

	return complete(ctx)
}

func (v Token) Processor(ctx context.Context, entry types.DataContainer) (types.TraversalInput, error) {
	return adapter.MongoProcessor[*TokenQueryGroup](ctx, entry)
}
