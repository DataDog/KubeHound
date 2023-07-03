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
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	Register(EscapeNsenter{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2897872074/CE+NSENTER
type EscapeNsenter struct {
}

func (e EscapeNsenter) Label() string {
	return "CE_NSENTER"
}

func (e EscapeNsenter) Name() string {
	return "ContainerEscapeNsenter"
}

func (e EscapeNsenter) BatchSize() int {
	return BatchSizeDefault
}

// Traversal delegates the traversal creation to the generic containerEscapeTraversal.
func (e EscapeNsenter) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("ce").
			MergeE(__.Select("ce")).
			Barrier().Limit(0)

		return g
	}
}

func (e EscapeNsenter) Processor(ctx context.Context, oic *converter.ObjectIdConverter, entry any) (any, error) {
	typed, ok := entry.(*containerEscapeGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	nid, err := oic.GraphId(ctx, typed.Node.Hex())
	if err != nil {
		return nil, err
	}

	cid, err := oic.GraphId(ctx, typed.Container.Hex())
	if err != nil {
		return nil, err
	}

	processed := map[any]any{
		gremlin.T.Label:       e.Label(),
		gremlin.Direction.In:  nid,
		gremlin.Direction.Out: cid,
	}

	return processed, nil
}

func (e EscapeNsenter) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	containers := adapter.MongoDB(store).Collection(collections.ContainerName)

	// Escape is possible with privileged containers that share the PID namespace
	filter := bson.M{
		"k8.securitycontext.privileged": true,
		"inherited.host_pid":            true,
	}

	// We just need a 1:1 mapping of the node and container to create this edge
	projection := bson.M{"_id": 1, "node_id": 1}

	cur, err := containers.Find(context.Background(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[containerEscapeGroup](ctx, cur, callback, complete)
}
