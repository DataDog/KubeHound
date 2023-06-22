package edge

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	Register(VolumeMount{})
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2891251713/VOLUME+MOUNT
type VolumeMount struct {
}

type mountGroup struct {
	Volume     primitive.ObjectID   `bson:"_id" json:"volume"`
	Node       primitive.ObjectID   `bson:"node_id" json:"node"`
	Containers []primitive.ObjectID `bson:"container" json:"containers"`
}

func (e VolumeMount) Label() string {
	return "VOLUME_MOUNT"
}

func (e VolumeMount) Name() string {
	return "VolumeMount"
}

func (e VolumeMount) BatchSize() int {
	return DefaultBatchSize
}

func (e VolumeMount) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*mountGroup](ctx, entry)
}

// Traversal expects a list of mountGroup objects serialized as map structures for injection into the graph.
// For each mountGroup, the traversal will: 1) find the node vertex with the same storeID as the mountGroup's node
// field, 2) find the volume vertex with the same storeID as the mountGroup's volume field, 3) create an edge between
// the node and volume vertices with the label "VOLUME_MOUNT", and 4) create an edge between each container vertex
// with the label "CONTAINER_MOUNT" and the volume vertex.
func (e VolumeMount) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("mg").
			SideEffect(
				__.V().HasLabel(vertex.NodeLabel).
					Where(P.Eq("mg")).
					By("storeID").
					By("node").
					AddE(e.Label()).
					To(
						__.V().HasLabel(vertex.VolumeLabel).
							Where(P.Eq("mg")).
							By("storeID").
							By("volume"))).
			Select("containers").
			Unfold().
			As("c").
			V().HasLabel(vertex.ContainerLabel).
			Where(P.Eq("c")).
			By("storeID").
			By().
			AddE(e.Label()).
			To(
				__.V().HasLabel(vertex.VolumeLabel).
					Where(P.Eq("mg")).
					By("storeID").
					By("volume")).
			Barrier().Limit(0)

		return g
	}
}

func (e VolumeMount) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	volumes := adapter.MongoDB(store).Collection(collections.VolumeName)
	pipeline := []bson.M{
		// Match volumes that have at least one mount
		{
			"$match": bson.M{
				"mounts": bson.M{
					"$exists": true,
					"$ne":     bson.A{},
				},
			},
		},
		// Group by volume ID and node ID
		{
			"$group": bson.M{
				"_id": bson.M{
					"volume": "$_id",
					"node":   "$node_id",
				},
				"containers": bson.M{
					"$addToSet": "$mounts.container_id",
				},
			},
		},
		// Flatten the containers set
		{
			"$unwind": "$containers",
		},
		// Project fields for MountGroup object
		{
			"$project": bson.M{
				"_id":       "$_id.volume",
				"node_id":   "$_id.node",
				"container": "$containers",
			},
		},
	}

	cur, err := volumes.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[mountGroup](ctx, cur, callback, complete)
}
