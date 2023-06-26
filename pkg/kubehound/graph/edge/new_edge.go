package edge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	Register(TestEdge{})
}

type testEdgeGroup struct {
	Node      primitive.ObjectID `bson:"node_id" json:"node"`
	Container primitive.ObjectID `bson:"_id" json:"container"`
}

// @@DOCLINK: TODO
type TestEdge struct {
}

func (e TestEdge) Label() string {
	return "TEST_EDGE"
}

func (e TestEdge) Name() string {
	return "TestEdge"
}

func (e TestEdge) BatchSize() int {
	return DefaultBatchSize
}

// Traversal delegates the traversal creation to the generic containerEscapeTraversal.
func (e TestEdge) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {

		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("te").
			MergeE(__.Select("te")).
			Barrier().Limit(0)

		return g
	}
}

func (e TestEdge) Processor(ctx context.Context, entry any) (any, error) {

	typed, ok := entry.(*testEdgeGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	// [(T.id):3753,(T.label):'route',(OUT):1,(IN):7,'dist':546]]

	processed := map[any]any{
		gremlin.T.Label:       e.Label(),
		gremlin.Direction.In:  cache.IdMap[typed.Node.Hex()],
		gremlin.Direction.Out: cache.IdMap[typed.Container.Hex()],
	}

	return processed, nil
}

func (e TestEdge) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	containers := adapter.MongoDB(store).Collection(collections.ContainerName)

	// Escape is possible with privileged containers via mounting the root directory on the host
	// and editing sensitive files e.g SSH keys, cronjobs, etc
	filter := bson.M{}

	// We just need a 1:1 mapping of the node and container to create this edge
	projection := bson.M{"_id": 1, "node_id": 1}

	cur, err := containers.Find(context.Background(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[testEdgeGroup](ctx, cur, callback, complete)
}

// structToMap creates a map from a simple input struct.
func structToMap(in any) (map[string]any, error) {
	var res map[string]any

	tmp, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(tmp, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
