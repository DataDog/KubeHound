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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	Register(&EscapeTokenVarLogSymlink{}, RegisterGraphDependency)
}

type EscapeTokenVarLogSymlink struct {
	BaseContainerEscape
}

// this is the same as containerEscapeGroup but with the container tag set to container_id
type containerIDToNodeIDEscapeGroup struct {
	Node      primitive.ObjectID `bson:"node_id" json:"node"`
	Container primitive.ObjectID `bson:"container_id" json:"container"`
}

func (e *EscapeTokenVarLogSymlink) Label() string {
	return "CE_VAR_LOG_SYMLINK"
}

// List of needed edges to run the traversal query
func (e *EscapeTokenVarLogSymlink) Dependencies() []string {
	return []string{"PERMISSION_DISCOVER", "IDENTITY_ASSUME", "VOLUME_DISCOVER"}
}

func (e *EscapeTokenVarLogSymlink) Name() string {
	return "ContainerEscapeVarLogSymlink"
}

func podToNodeProcessor(ctx context.Context, oic *converter.ObjectIDConverter, edgeLabel string, entry any) (any, error) {
	typed, ok := entry.(*containerIDToNodeIDEscapeGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, edgeLabel, typed.Node, typed.Container)
}

// Processor delegates the processing tasks to to the generic containerEscapeProcessor.
func (e *EscapeTokenVarLogSymlink) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return podToNodeProcessor(ctx, oic, e.Label(), entry)
}

func (e *EscapeTokenVarLogSymlink) Traversal() types.EdgeTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []any) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()
		g.V().HasLabel("PermissionSet").
			InE("PERMISSION_DISCOVER").OutV().
			InE("IDENTITY_ASSUME").OutV().
			HasLabel("Container").As("c").OutE("VOLUME_DISCOVER").
			Has("sourcePath", P.Within("/", "/var", "/var/log")).
			AddE(e.Label()).
			From("c").To(inserts)
		return g
	}
}

func (e *EscapeTokenVarLogSymlink) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	permissionSets := adapter.MongoDB(store).Collection(collections.PermissionSetName)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"rules": bson.M{
					"$elemMatch": bson.M{
						"$and": bson.A{
							bson.M{"$or": bson.A{
								bson.M{"resources": "pods/log"},
								bson.M{"resources": "pods/*"},
								bson.M{"resources": "*"},
							}},
							bson.M{"$or": bson.A{
								bson.M{"verbs": "get"},
								bson.M{"verbs": "*"},
							}},
							bson.M{"resourcenames": nil}, // TODO: handle resource scope
						},
					},
				},
			},
		},
		{
			"$project": bson.M{
				"_id": 1,
			},
		},
	}

	cur, err := permissionSets.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[containerIDToNodeIDEscapeGroup](ctx, cur, callback, complete)
}
