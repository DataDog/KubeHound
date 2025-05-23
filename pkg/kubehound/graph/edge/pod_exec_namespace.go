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
	Register(&PodExecNamespace{}, RegisterDefault)
}

type PodExecNamespace struct {
	BaseEdge
}

type podExecNSGroup struct {
	Role primitive.ObjectID `bson:"_id" json:"role"`
	Pod  primitive.ObjectID `bson:"pod" json:"pod"`
}

func (e *PodExecNamespace) Label() string {
	return "POD_EXEC"
}

func (e *PodExecNamespace) Name() string {
	return "PodExecNamespace"
}

func (e *PodExecNamespace) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueContainerAdministrationCommand
}

func (e *PodExecNamespace) AttckTacticID() AttckTacticID {
	return AttckTacticExecution
}

func (e *PodExecNamespace) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*podExecNSGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Role, typed.Pod, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

// Stream finds all roles that are namespaced and have pod/exec or equivalent wildcard permissions and matching pods.
// Matching pods are defined as all pods that share the role namespace or non-namespaced pods.
func (e *PodExecNamespace) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	permissionSets := adapter.MongoDB(ctx, store).Collection(collections.PermissionSetName)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"is_namespaced":   true,
				"runtime.runID":   e.runtime.RunID.String(),
				"runtime.cluster": e.runtime.ClusterName,
				"rules": bson.M{
					"$elemMatch": bson.M{
						"$and": bson.A{
							bson.M{"$or": bson.A{
								bson.M{"apigroups": ""},
								bson.M{"apigroups": "*"},
							}},
							bson.M{"$or": bson.A{
								bson.M{"resources": "pods/exec"},
								bson.M{"resources": "pods/*"},
								bson.M{"resources": "*"},
							}},
							bson.M{"$or": bson.A{
								bson.M{"verbs": "create"},
								bson.M{"verbs": "*"},
							}},
							bson.M{"resourcenames": nil}, // TODO: handle resource scope
						},
					},
				},
			},
		},
		{
			"$lookup": bson.M{
				"as":   "podsInNamespace",
				"from": "pods",
				"let": bson.M{
					"roleNamespace": "$namespace",
				},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$or": bson.A{
								bson.M{"$expr": bson.M{
									"$eq": bson.A{
										"$k8.objectmeta.namespace", "$$roleNamespace",
									},
								}},
								bson.M{"is_namespaced": false},
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
				},
			},
		},
		{
			"$unwind": "$podsInNamespace",
		},
		{
			"$project": bson.M{
				"_id": 1,
				"pod": "$podsInNamespace._id",
			},
		},
	}

	cur, err := permissionSets.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[podExecNSGroup](ctx, cur, callback, complete)
}
