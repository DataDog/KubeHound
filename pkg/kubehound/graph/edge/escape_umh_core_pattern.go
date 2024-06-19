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

var ProcMountList = bson.A{
	"/",
	"/proc",
	"/proc/sys",
	"/proc/sys/kernel",
}

func init() {
	Register(&EscapeCorePattern{}, RegisterDefault)
}

type EscapeCorePattern struct {
	BaseContainerEscape
}

func (e *EscapeCorePattern) Label() string {
	return "CE_UMH_CORE_PATTERN"
}

func (e *EscapeCorePattern) Name() string {
	return "ContainerEscapeCorePattern"
}

type escapeCorePatternGroup struct {
	Node      primitive.ObjectID `bson:"node_id" json:"node"`
	Container primitive.ObjectID `bson:"container_id" json:"container"`
}

func (e *EscapeCorePattern) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*escapeCorePatternGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Container, typed.Node)
}

func (e *EscapeCorePattern) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {
	volumes := adapter.MongoDB(store).Collection(collections.VolumeName)

	filter := bson.M{
		"type": shared.VolumeTypeHost,
		"source": bson.M{
			"$in": ProcMountList,
		},
		"runtime.runID":   e.runtime.RunID.String(),
		"runtime.cluster": e.runtime.ClusterName,
	}

	// We just need a 1:1 mapping of the node and container to create this edge
	projection := bson.M{"container_id": 1, "node_id": 1}

	cur, err := volumes.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[escapeCorePatternGroup](ctx, cur, callback, complete)
}
