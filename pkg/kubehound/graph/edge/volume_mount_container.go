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
)

func init() {
	Register(&VolumeMountContainer{}, RegisterDefault)
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2891251713/VOLUME+MOUNT
type VolumeMountContainer struct {
	BaseEdge
}

type containerMountGroup struct {
	Volume    primitive.ObjectID `bson:"_id" json:"volume"`
	Container primitive.ObjectID `bson:"container_id" json:"container"`
}

func (e *VolumeMountContainer) Label() string {
	return "VOLUME_MOUNT"
}

func (e *VolumeMountContainer) Name() string {
	return "VolumeMountContainer"
}

func (e *VolumeMountContainer) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*containerMountGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Container, typed.Volume)
}

func (e *VolumeMountContainer) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	volumes := adapter.MongoDB(store).Collection(collections.VolumeName)
	pipeline := []bson.M{
		// Match volumes that have at least one mount
		{
			"$match": bson.M{
				"mounts": bson.M{
					"$exists": true,
					"$ne":     bson.A{},
				},
			},
		},
		// Flatten the mounts set
		{
			"$unwind": "$mounts",
		},
		// Project a volume id / container id pair
		{
			"$project": bson.M{
				"_id":          1,
				"container_id": "$mounts.container_id",
			},
		},
	}

	cur, err := volumes.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[containerMountGroup](ctx, cur, callback, complete)
}
