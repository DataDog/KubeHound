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
	Register(&PodAttach{}, RegisterDefault)
}

type PodAttach struct {
	BaseEdge
}

type podAttachGroup struct {
	Node primitive.ObjectID `bson:"node_id" json:"node"`
	Pod  primitive.ObjectID `bson:"_id" json:"pod"`
}

func (e *PodAttach) Label() string {
	return "POD_ATTACH"
}

func (e *PodAttach) Name() string {
	return "PodAttach"
}

func (e *PodAttach) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueContainerAdministrationCommand
}

func (e *PodAttach) AttckTacticID() AttckTacticID {
	return AttckTacticExecution
}

func (e *PodAttach) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*podAttachGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Node, typed.Pod, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *PodAttach) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	pods := adapter.MongoDB(ctx, store).Collection(collections.PodName)

	// We just need a 1:1 mapping of the node and pod to create this edge
	projection := bson.M{"_id": 1, "node_id": 1}

	filter := bson.M{
		"runtime.runID":   e.runtime.RunID.String(),
		"runtime.cluster": e.runtime.ClusterName,
	}

	cur, err := pods.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[podAttachGroup](ctx, cur, callback, complete)
}
