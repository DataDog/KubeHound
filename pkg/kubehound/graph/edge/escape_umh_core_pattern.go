package edge

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"go.mongodb.org/mongo-driver/bson"
)

var ProcMountList = bson.A{
	"/",
	"/proc",
	"/proc/sys",
	"/proc/sys/kernel",
}

func init() {
	Register(&EscapeCorePattern{}, RegisterDefault)
}

type EscapeCorePattern struct {
	BaseContainerEscape
}

func (e *EscapeCorePattern) Label() string {
	return "CE_UMH_CORE_PATTERN"
}

func (e *EscapeCorePattern) Name() string {
	return "ContainerEscapeCorePattern"
}

func (e *EscapeCorePattern) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return containerEscapeProcessor(ctx, oic, e.Label(), entry)
}

func (e *EscapeCorePattern) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {
	containers := adapter.MongoDB(store).Collection(collections.ContainerName)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"k8.securitycontext.runasuser": 0,
				"runtime.runID":                e.runtime.RunID.String(),
				"runtime.cluster":              e.runtime.ClusterName,
			},
		},
		{
			"$lookup": bson.M{
				"as":   "procMountContainers",
				"from": "volumes",
				"let": bson.M{
					"rootContainerId": "$container_id",
				},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$and": bson.A{
								bson.M{"$expr": bson.M{
									"$eq": bson.A{
										"$container_id", "$$rootContainerId",
									},
								}},
							},
							"type": shared.VolumeTypeHost,
							"source": bson.M{
								"$in": ProcMountList,
							},
							"runtime.runID":   e.runtime.RunID.String(),
							"runtime.cluster": e.runtime.ClusterName,
						},
					},
				},
			},
		},
		{
			"$project": bson.M{
				"_id":     1,
				"node_id": 1,
			},
		},
	}

	cur, err := containers.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)
	return adapter.MongoCursorHandler[containerEscapeGroup](ctx, cur, callback, complete)
}
