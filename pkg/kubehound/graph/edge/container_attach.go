package edge

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
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
	Register(ContainerAttach{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2883354625/CONTAINER+ATTACH
type ContainerAttach struct {
	cfg *config.EdgeBuilderConfig
}

type containerAttachGroup struct {
	Pod       primitive.ObjectID `bson:"pod_id" json:"pod"`
	Container primitive.ObjectID `bson:"_id" json:"container"`
}

func (e ContainerAttach) Initialize(cfg *config.EdgeBuilderConfig) error {
	e.cfg = cfg
	return nil
}

func (e ContainerAttach) Label() string {
	return "CONTAINER_ATTACH"
}

func (e ContainerAttach) Name() string {
	return "ContainerAttach"
}

func (e ContainerAttach) BatchSize() int {
	return BatchSizeDefault
	//return e.cfg.BatchSizeDefault
}

func (e ContainerAttach) Processor(ctx context.Context, oic *converter.ObjectIdConverter, entry any) (any, error) {
	typed, ok := entry.(*containerAttachGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Pod, typed.Container)
}

func (e ContainerAttach) Traversal() types.EdgeTraversal {
	return adapter.DefaultEdgeTraversal()
}

func (e ContainerAttach) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	containers := adapter.MongoDB(store).Collection(collections.ContainerName)

	// We just need a 1:1 mapping of the container and pod to create this edge
	projection := bson.M{"_id": 1, "pod_id": 1}

	cur, err := containers.Find(context.Background(), bson.M{}, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[containerAttachGroup](ctx, cur, callback, complete)
}
