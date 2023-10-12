package edge

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	Register(&VolumeAccess{}, RegisterDefault)
}

type VolumeAccess struct {
	BaseEdge
}

type volumeAccessGroup struct {
	Volume primitive.ObjectID `bson:"_id" json:"volume"`
	Node   primitive.ObjectID `bson:"node_id" json:"node"`
}

func (e *VolumeAccess) Label() string {
	return "VOLUME_ACCESS"
}

func (e *VolumeAccess) Name() string {
	return "VolumeAccess"
}

func (e *VolumeAccess) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*volumeAccessGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Node, typed.Volume)
}

func (e *VolumeAccess) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	volumes := adapter.MongoDB(store).Collection(collections.VolumeName)

	// We just need a 1:1 mapping of the node and volume to create this edge
	projection := bson.M{"_id": 1, "node_id": 1}

	cur, err := volumes.Find(ctx, bson.M{}, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[volumeAccessGroup](ctx, cur, callback, complete)
}
