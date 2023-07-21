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
	Register(&VolumeMount{}, RegisterDefault)
}

type VolumeMount struct {
	BaseEdge
}

type volumeMountGroup struct {
	Volume    primitive.ObjectID `bson:"_id" json:"volume"`
	Container primitive.ObjectID `bson:"container_id" json:"container"`
}

func (e *VolumeMount) Label() string {
	return "VOLUME_MOUNT"
}

func (e *VolumeMount) Name() string {
	return "VolumeMount"
}

func (e *VolumeMount) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*volumeMountGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Container, typed.Volume)
}

func (e *VolumeMount) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	volumes := adapter.MongoDB(store).Collection(collections.VolumeName)

	// We just need a 1:1 mapping of the container and volume to create this edge
	projection := bson.M{"_id": 1, "container_id": 1}

	cur, err := volumes.Find(context.Background(), bson.M{}, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}

	return adapter.MongoCursorHandler[volumeMountGroup](ctx, cur, callback, complete)
}
