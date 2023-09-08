package edge

import (
	"context"

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
	Register(&EscapeTokenVarLogSymlink{}, RegisterDefault)
}

type EscapeTokenVarLogSymlink struct {
	BaseContainerEscape
}

// this is the same as containerEscapeGroup but with the container tag set to container_id
type containerFieldEscapeGroup struct {
	Node      primitive.ObjectID `bson:"node_id" json:"node"`
	Container primitive.ObjectID `bson:"container_id" json:"container"`
}

func (e *EscapeTokenVarLogSymlink) Label() string {
	return "CE_VAR_LOG_SYMLINK"
}

func (e *EscapeTokenVarLogSymlink) Name() string {
	return "ContainerEscapeVarLogSymlink"
}

// Processor delegates the processing tasks to to the generic containerEscapeProcessor.
func (e *EscapeTokenVarLogSymlink) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return containerEscapeProcessor(ctx, oic, e.Label(), entry)
}

func (e *EscapeTokenVarLogSymlink) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	containers := adapter.MongoDB(store).Collection(collections.ContainerName)

	// Container.volumeMounts[*].hostPath.path contains /var/log
	// Container.volumeMounts[*].hostPath.readOnly is false
	// Container.securityContext.runAsUser is 0
	filter := bson.M{
		"$and": bson.A{
			bson.M{"source": "/var/log"},
			bson.M{"readonly": false},
		},
	}

	// We just need a 1:1 mapping of the node and container to create this edge
	projection := bson.M{"container_id": 1, "node_id": 1}

	cur, err := containers.Find(context.Background(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[containerEscapeGroup](ctx, cur, callback, complete)
}
