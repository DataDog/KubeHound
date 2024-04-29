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
	Register(&IdentityAssumeNode{}, RegisterDefault)
}

type IdentityAssumeNode struct {
	BaseEdge
}

type nodeIdentityGroup struct {
	Node     primitive.ObjectID `bson:"_id" json:"node"`
	Identity primitive.ObjectID `bson:"user_id" json:"user_id"`
}

func (e *IdentityAssumeNode) Label() string {
	return "IDENTITY_ASSUME"
}

func (e *IdentityAssumeNode) Name() string {
	return "IdentityAssumeNode"
}

func (e *IdentityAssumeNode) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*nodeIdentityGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Node, typed.Identity)
}

func (e *IdentityAssumeNode) Stream(ctx context.Context, store storedb.Provider, c cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	nodes := adapter.MongoDB(store).Collection(collections.NodeName)

	// Nodes will either have a dedicated user based on node name or use the default system:nodes group
	// See reference for details: https://kubernetes.io/docs/reference/access-authn-authz/node/
	projection := bson.M{"_id": 1, "user_id": 1}

	// If the default node group has no permissions, we do not set a user id
	filter := bson.M{
		"user_id":         bson.M{"$ne": primitive.NilObjectID},
		"runtime.runID":   e.runtime.RunID.String(),
		"runtime.cluster": e.runtime.ClusterName,
	}

	cur, err := nodes.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[nodeIdentityGroup](ctx, cur, callback, complete)
}
