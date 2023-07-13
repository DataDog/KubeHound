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
	Register(&VolumeMountNode{}, RegisterDefault)
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2891251713/VOLUME+MOUNT
type VolumeMountNode struct {
	BaseEdge
}

type nodeMountGroup struct {
	Volume primitive.ObjectID `bson:"_id" json:"volume"`
	Node   primitive.ObjectID `bson:"node_id" json:"node"`
}

func (e *VolumeMountNode) Label() string {
	return "VOLUME_MOUNT"
}

func (e *VolumeMountNode) Name() string {
	return "VolumeMountNode"
}

func (e *VolumeMountNode) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*nodeMountGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Node, typed.Volume)
}

func (e *VolumeMountNode) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	volumes := adapter.MongoDB(store).Collection(collections.VolumeName)

	// Only match volumes that have at least one mount
	filter := bson.M{
		"mounts": bson.M{
			"$exists": true,
			"$ne":     bson.A{},
		},
	}

	// We just need a 1:1 mapping of the node and volume to create this edge
	projection := bson.M{"_id": 1, "node_id": 1}

	cur, err := volumes.Find(context.Background(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[nodeMountGroup](ctx, cur, callback, complete)
}
