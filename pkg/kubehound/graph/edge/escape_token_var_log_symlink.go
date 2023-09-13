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
	Register(&EscapeTokenVarLogSymlink{}, RegisterDefault)
}

type EscapeTokenVarLogSymlink struct {
	BaseContainerEscape
}

// this is the same as containerEscapeGroup but with the container tag set to container_id
type podToNodeEscapeGroup struct {
	Node primitive.ObjectID `bson:"node_id" json:"node"`
	Pod  primitive.ObjectID `bson:"pod_id" json:"pod"`
}

func (e *EscapeTokenVarLogSymlink) Label() string {
	return "CE_VAR_LOG_SYMLINK"
}

func (e *EscapeTokenVarLogSymlink) Name() string {
	return "ContainerEscapeVarLogSymlink"
}

func podToNodeProcessor(ctx context.Context, oic *converter.ObjectIDConverter, edgeLabel string, entry any) (any, error) {
	typed, ok := entry.(*podToNodeEscapeGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, edgeLabel, typed.Node, typed.Pod)
}

// Processor delegates the processing tasks to to the generic containerEscapeProcessor.
func (e *EscapeTokenVarLogSymlink) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return podToNodeProcessor(ctx, oic, e.Label(), entry)
}

func (e *EscapeTokenVarLogSymlink) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	volumes := adapter.MongoDB(store).Collection(collections.VolumeName)

	// Container.volumeMounts[*].hostPath.path contains /var/log
	// Container.volumeMounts[*].hostPath.readOnly is false
	// Container.securityContext.runAsUser is 0
	pipeline := bson.A{
		bson.D{
			{Key: "$match",
				Value: bson.D{
					{Key: "$and",
						Value: bson.A{
							bson.D{
								{Key: "$or",
									Value: bson.A{
										bson.D{{Key: "source", Value: "/var/log"}},
										bson.D{{Key: "source", Value: "/var"}},
										bson.D{{Key: "source", Value: "/"}},
									},
								},
							},
							bson.D{{Key: "readonly", Value: false}},
						},
					},
				},
			},
		},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "containers"},
					{Key: "localField", Value: "container_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "containers"},
				},
			},
		},
		bson.D{
			{Key: "$unwind",
				Value: bson.D{
					{Key: "path", Value: "$containers"},
					{Key: "preserveNullAndEmptyArrays", Value: false},
				},
			},
		},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "identities"},
					{Key: "localField", Value: "containers.inherited.service_account"},
					{Key: "foreignField", Value: "name"},
					{Key: "as", Value: "service_account"},
				},
			},
		},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "rolebindings"},
					{Key: "localField", Value: "service_account.name"},
					{Key: "foreignField", Value: "subjects.subject.name"},
					{Key: "as", Value: "rolebindings"},
				},
			},
		},
		bson.D{
			{Key: "$unwind",
				Value: bson.D{
					{Key: "path", Value: "$rolebindings"},
					{Key: "preserveNullAndEmptyArrays", Value: false},
				},
			},
		},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "permissionsets"},
					{Key: "localField", Value: "rolebindings.role_id"},
					{Key: "foreignField", Value: "role_id"},
					{Key: "as", Value: "perms"},
				},
			},
		},
		bson.D{
			{Key: "$match",
				Value: bson.D{
					{Key: "$and",
						Value: bson.A{
							bson.D{
								{Key: "perms.rules.verbs",
									Value: bson.D{
										{Key: "$in",
											Value: bson.A{
												"get",
											},
										},
									},
								},
							},
							bson.D{
								{Key: "perms.rules.resources",
									Value: bson.D{
										{Key: "$in",
											Value: bson.A{
												"pods/log",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		bson.D{
			{Key: "$project",
				Value: bson.D{
					{Key: "node_id", Value: 1},
					{Key: "pod_id", Value: 1},
				},
			},
		},
	}

	cur, err := volumes.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	// We just need a 1:1 mapping of the node and container to create this edge
	// projection := bson.M{"container_id": 1, "node_id": 1}

	// cur, err := containers.Find(context.Background(), filter, options.Find().SetProjection(projection))
	// if err != nil {
	// 	return err
	// }
	// defer cur.Close(ctx)

	return adapter.MongoCursorHandler[podToNodeEscapeGroup](ctx, cur, callback, complete)
}
