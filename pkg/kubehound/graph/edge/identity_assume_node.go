package edge

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kube"
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

type nodeQueryResult struct {
	NodeId   primitive.ObjectID `bson:"_id" json:"node_id"`
	NodeName string             `bson:"name" json:"node_name"`
}

type nodeIdentityGroup struct {
	Node primitive.ObjectID `bson:"node" json:"node"`
	Role primitive.ObjectID `bson:"role" json:"role"`
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

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Node, typed.Role)
}

func (e *IdentityAssumeNode) Stream(ctx context.Context, store storedb.Provider, c cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	nodes := adapter.MongoDB(store).Collection(collections.NodeName)

	// Nodes will either have a dedicated user based on node name or use the default system:nodes group
	// See reference for details: https://kubernetes.io/docs/reference/access-authn-authz/node/
	projection := bson.M{"_id": 1, "name": "$k8.objectmeta.name"}

	cur, err := nodes.Find(context.Background(), bson.M{}, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var entry nodeQueryResult
		err := cur.Decode(&entry)
		if err != nil {
			return err
		}

		TODO WE SHOULD DO THIOS ON INGEST AND ADD AN IDENTITYID FIELD TO NODE
		// Resolve the node identity id
		nodeUserId, err := kube.NodeIdentity(ctx, c, entry.NodeName)
		if err != nil {
			// This is ok - there won't always be an identity
			continue
		}

		err = callback(ctx, &nodeIdentityGroup{
			Node: entry.NodeId,
			Role: nodeUserId,
		})
		if err != nil {
			return err
		}
	}

	return complete(ctx)
}
