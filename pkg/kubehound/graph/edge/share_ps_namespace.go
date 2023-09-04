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
)

func init() {
	Register(&SharePSNamespace{}, RegisterDefault)
}

type SharePSNamespace struct {
	BaseEdge
}

type sharedPsNamespaceGroup struct {
	ContainerA []primitive.ObjectID `bson:"_id" json:"containers"`
}

func (e *SharePSNamespace) Label() string {
	return "SHARE_PS_NAMESPACE"
}

func (e *SharePSNamespace) Name() string {
	return "SharePSNamespace"
}

// Processor delegates the processing tasks to to the generic containerEscapeProcessor.
func (e *SharePSNamespace) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return containerEscapeProcessor(ctx, oic, e.Label(), entry)
}

func (e *SharePSNamespace) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {
	// Open an aggregation cursor
	coll := adapter.MongoDB(store).Collection(collections.PodName)
	cur, err := coll.Aggregate(ctx, bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "k8.spec.shareprocessnamespace", Value: true},
		}}},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "containers"},
					{Key: "localField", Value: "_id"},
					{Key: "foreignField", Value: "pod_id"},
					{Key: "as", Value: "containers_with_shared_ns"},
				},
			},
		},
		bson.D{{Key: "$project", Value: bson.D{{Key: "containers_with_shared_ns", Value: bson.D{{Key: "_id", Value: 1}}}}}},
	})

	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[sharedPsNamespaceGroup](ctx, cur, callback, complete)
}
