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
	Register(&EscapeModuleLoad{}, RegisterDefault)
}

type EscapeModuleLoad struct {
	BaseContainerEscape
}

func (e *EscapeModuleLoad) Label() string {
	return "CE_MODULE_LOAD"
}

func (e *EscapeModuleLoad) Name() string {
	return "ContainerEscapeModuleLoad"
}

// Processor delegates the processing tasks to the generic containerEscapeProcessor.
func (e *EscapeModuleLoad) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return containerEscapeProcessor(ctx, oic, e.Label(), entry)
}

func (e *EscapeModuleLoad) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	containers := adapter.MongoDB(store).Collection(collections.ContainerName)

	// Escape is possible with privileged containers or CAP_SYS_MODULE loaded explicitly
	filter := bson.M{
		"$or": bson.A{
			bson.M{"k8.securitycontext.privileged": true},
			bson.M{"k8.securitycontext.capabilities.add": "SYS_MODULE"},
		}}

	// We just need a 1:1 mapping of the node and container to create this edge
	projection := bson.M{"_id": 1, "node_id": 1}

	cur, err := containers.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[containerEscapeGroup](ctx, cur, callback, complete)
}
