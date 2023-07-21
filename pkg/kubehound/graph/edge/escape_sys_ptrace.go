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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	Register(&EscapeSysPtrace{}, RegisterDefault)
}

type EscapeSysPtrace struct {
	BaseContainerEscape
}

func (e *EscapeSysPtrace) Label() string {
	return "CE_SYS_PTRACE"
}

func (e *EscapeSysPtrace) Name() string {
	return "ContainerEscapeSysPtrace"
}

// Processor delegates the processing tasks to to the generic containerEscapeProcessor.
func (e *EscapeSysPtrace) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return containerEscapeProcessor(ctx, oic, e.Label(), entry)
}

func (e *EscapeSysPtrace) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	containers := adapter.MongoDB(store).Collection(collections.ContainerName)

	// Escape is possible with shared host pid namespace and SYS_PTRACE/SYS_ADMIN capabilities
	filter := bson.M{
		"$and": bson.A{
			bson.M{"inherited.host_pid": true},
			bson.M{"k8.securitycontext.capabilities.add": "SYS_PTRACE"},
			bson.M{"k8.securitycontext.capabilities.add": "SYS_ADMIN"},
		}}

	// We just need a 1:1 mapping of the node and container to create this edge
	projection := bson.M{"_id": 1, "node_id": 1}

	cur, err := containers.Find(context.Background(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[containerEscapeGroup](ctx, cur, callback, complete)
}
