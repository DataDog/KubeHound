package edge

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
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
	Register(&EscapeNsenter{}, RegisterDefault)
}

type EscapeNsenter struct {
	BaseContainerEscape
}

func (e *EscapeNsenter) Label() string {
	return "CE_NSENTER"
}

func (e *EscapeNsenter) Name() string {
	return "ContainerEscapeNsenter"
}

// Processor delegates the processing tasks to the generic containerEscapeProcessor.
func (e *EscapeNsenter) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return containerEscapeProcessor(ctx, oic, e.Label(), entry)
}

func (e *EscapeNsenter) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader, runtime *config.DynamicConfig,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	containers := adapter.MongoDB(store).Collection(collections.ContainerName)

	// Escape is possible with privileged containers that share the PID namespace
	filter := bson.M{
		"k8.securitycontext.privileged": true,
		"inherited.host_pid":            true,
		"runtime.runID":                 runtime.RunID.String(),
		"runtime.cluster":               runtime.ClusterName,
	}

	// We just need a 1:1 mapping of the node and container to create this edge
	projection := bson.M{"_id": 1, "node_id": 1}

	cur, err := containers.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[containerEscapeGroup](ctx, cur, callback, complete)
}
