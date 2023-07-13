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
	Register(&shareProcessNamespace{}, RegisterDefault)
}

// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2880275294/SHARED+PS+NAMESPACE
type shareProcessNamespace struct {
	BaseEdge
}

// We want to map all the containers that shares the same process namespace
type shareProcessNamespaceContainers struct {
	ContainerA primitive.ObjectID `bson:"containerA" json:"containerA"`
	ContainerB primitive.ObjectID `bson:"containerB" json:"containerB"`
}

func (e shareProcessNamespace) Label() string {
	return "SHARE_PS_NAMESPACE"
}

func (e shareProcessNamespace) Name() string {
	return "shareProcessNamespace"
}

func (e *shareProcessNamespace) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*shareProcessNamespaceContainers)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.ContainerA, typed.ContainerB)
}

// Stream finds all roles that are namespaced and have pod/exec or equivalent wildcard permissions and matching pods.
// Matching pods are defined as all pods that share the role namespace or non-namespaced pods.
func (e *shareProcessNamespace) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	roles := adapter.MongoDB(store).Collection(collections.RoleName)
	pipeline := []bson.M{
		// find pods that have shareProcessNamespace set
		{
			"$match": bson.M{"shareProcessNamespace": true},
		},
		// Gather pods ID and their related containers
		{
			"$group": bson.M{
				"_id": "$pod_id",
				"containers": bson.M{
					"$push": "$_id",
				},
			},
		},
	}

	cur, err := roles.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[shareProcessNamespaceContainers](ctx, cur, callback, complete)
}
