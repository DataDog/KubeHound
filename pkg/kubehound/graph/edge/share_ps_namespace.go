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
)

func init() {
	Register(&SharePSNamespace{}, RegisterDefault)
}

type SharePSNamespace struct {
	BaseEdge
}

type sharedPsNamespaceGroup struct {
	Containers []primitive.ObjectID `bson:"container_ids" json:"container_ids"`
}
type sharedPsNamespaceGroupPair struct {
	ContainerA primitive.ObjectID `bson:"container_a_id" json:"container_a"`
	ContainerB primitive.ObjectID `bson:"container_b_id" json:"container_b"`
}

func (e *SharePSNamespace) Label() string {
	return "SHARE_PS_NAMESPACE"
}

func (e *SharePSNamespace) Name() string {
	return "SharePSNamespace"
}

// Processor delegates the processing tasks to to the generic containerEscapeProcessor.
func (e *SharePSNamespace) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*sharedPsNamespaceGroupPair)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.ContainerA, typed.ContainerB)
}

func (e *SharePSNamespace) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	coll := adapter.MongoDB(store).Collection(collections.PodName)
	pipeline := bson.A{
		bson.D{{"$match", bson.D{{"k8.spec.shareprocessnamespace", true}}}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "containers"},
					{"localField", "_id"},
					{"foreignField", "pod_id"},
					{"as", "containers_with_shared_ns"},
				},
			},
		},
		bson.D{
			{"$project",
				bson.D{
					{"_id", 1},
					{"containers_with_shared_ns", bson.D{{"_id", 1}}},
				},
			},
		},
		bson.D{
			{"$project",
				bson.D{
					{"_id", 0},
					{"container_ids", "$containers_with_shared_ns._id"},
				},
			},
		},
	}
	cur, err := coll.Aggregate(ctx, pipeline)

	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	pairs := []sharedPsNamespaceGroupPair{}
	for cur.Next(ctx) {
		var entry sharedPsNamespaceGroup
		err := cur.Decode(&entry)
		if err != nil {
			return err
		}

		for _, containerSrc := range entry.Containers {
			for _, containerDst := range entry.Containers {
				// No need to create a link with itself
				if containerSrc == containerDst {
					continue
				}
				pairs = append(pairs,
					sharedPsNamespaceGroupPair{
						ContainerA: containerSrc,
						ContainerB: containerDst,
					},
				)
			}
		}
	}

	for _, pair := range pairs {
		err = callback(ctx, &pair)
		if err != nil {
			return err
		}
	}

	return complete(ctx)
}
