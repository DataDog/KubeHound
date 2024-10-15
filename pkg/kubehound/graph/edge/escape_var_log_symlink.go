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
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	Register(&EscapeVarLogSymlink{}, RegisterGraphDependency)
}

type EscapeVarLogSymlink struct {
	BaseContainerEscape
}

// The mongodb query returns a list of permissionSet
type permissionSetIDEscapeGroup struct {
	PermissionSetID primitive.ObjectID `bson:"_id" json:"permission_set"`
}

func (e *EscapeVarLogSymlink) Label() string {
	return "CE_VAR_LOG_SYMLINK"
}

// List of needed edges to run the traversal query
func (e *EscapeVarLogSymlink) Dependencies() []string {
	return []string{"PERMISSION_DISCOVER", "IDENTITY_ASSUME", "VOLUME_DISCOVER", "VOLUME_ACCESS"}
}

func (e *EscapeVarLogSymlink) Name() string {
	return "ContainerEscapeVarLogSymlink"
}

func (e *EscapeVarLogSymlink) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*permissionSetIDEscapeGroup)
	if !ok {
		return nil, fmt.Errorf("Invalid type passed to processor: %T", entry)
	}

	permissionSetVertexID, err := oic.GraphID(ctx, typed.PermissionSetID.Hex())
	if err != nil {
		return nil, fmt.Errorf("%s edge IN id convert: %w", e.Label(), err)
	}

	return permissionSetVertexID, nil
}

func (e *EscapeVarLogSymlink) Traversal() types.EdgeTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []any) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()
		// reduce the graph to only these permission sets
		g.V(inserts...).HasLabel("PermissionSet").
			// get identity vertices
			InE("PERMISSION_DISCOVER").OutV().
			// get container vertices
			InE("IDENTITY_ASSUME").OutV().
			// save container vertices as "c" so we can link to it to the node via CE_VAR_LOG_SYMLINK
			HasLabel("Container").As("c").
			// Get all the volumes
			OutE("VOLUME_DISCOVER").InV().
			Has("type", shared.VolumeTypeHost).
			// filter only the volumes that are "affected" by this attacks ("/", "/var", "/var/log").
			Has("sourcePath", P.Within("/", "/var", "/var/log")).
			// get the node related to that volume mount
			InE("VOLUME_ACCESS").OutV().
			HasLabel("Node").As("n").
			AddE("CE_VAR_LOG_SYMLINK").From("c").To("n").
			Barrier().Limit(0)

		return g
	}
}

func (e *EscapeVarLogSymlink) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	permissionSets := adapter.MongoDB(ctx, store).Collection(collections.PermissionSetName)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"rules": bson.M{
					"$elemMatch": bson.M{
						"$and": bson.A{
							bson.M{"$or": bson.A{
								bson.M{"apigroups": ""},
								bson.M{"apigroups": "*"},
							}},
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
				"runtime.runID":   e.runtime.RunID.String(),
				"runtime.cluster": e.runtime.ClusterName,
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

	return adapter.MongoCursorHandler[permissionSetIDEscapeGroup](ctx, cur, callback, complete)
}
