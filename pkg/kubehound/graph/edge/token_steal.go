package edge

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	Register(&TokenSteal{}, RegisterDefault)
}

type tokenStealGroup struct {
	Volume   primitive.ObjectID `bson:"_id" json:"volume"`
	Identity primitive.ObjectID `bson:"projected_id" json:"identity"`
}

type TokenSteal struct {
	BaseEdge
}

func (e *TokenSteal) Label() string {
	return "TOKEN_STEAL"
}

func (e *TokenSteal) Name() string {
	return "TokenSteal"
}

func (e *TokenSteal) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*tokenStealGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Volume, typed.Identity)
}

func (e *TokenSteal) Stream(ctx context.Context, sdb storedb.Provider, c cache.CacheReader,
	process types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	volumes := adapter.MongoDB(sdb).Collection(collections.VolumeName)

	filter := bson.M{
		"type":            shared.VolumeTypeProjected,
		"projected_id":    bson.M{"$ne": nil},
		"runtime.runID":   e.runtime.RunID.String(),
		"runtime.cluster": e.runtime.ClusterName,
	}

	// We just need a 1:1 mapping of the volume and projected service account to create this edge
	projection := bson.M{"_id": 1, "projected_id": 1}

	cur, err := volumes.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[tokenStealGroup](ctx, cur, process, complete)
}
