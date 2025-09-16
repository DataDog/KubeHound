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
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	AppArmorEnabledMajorVersion = 1
	AppArmorEnabledMinorVersion = 31
)

func init() {
	Register(&EscapeSysPtrace{}, RegisterDefault)
}

type EscapeSysPtrace struct {
	BaseContainerEscape
}

func (e *EscapeSysPtrace) Label() string {
	return "CE_SYS_PTRACE"
}

func (e *EscapeSysPtrace) Name() string {
	return "ContainerEscapeSysPtrace"
}

func (e *EscapeSysPtrace) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueEscapeToHost
}

func (e *EscapeSysPtrace) AttckTacticID() AttckTacticID {
	return AttckTacticPrivilegeEscalation
}

// Processor delegates the processing tasks to the generic containerEscapeProcessor.
func (e *EscapeSysPtrace) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return containerEscapeProcessor(ctx, oic, e.Label(), entry, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *EscapeSysPtrace) Stream(ctx context.Context, store storedb.Provider, _ cache.CacheReader,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	containers := adapter.MongoDB(ctx, store).Collection(collections.ContainerName)

	// Escape is possible with shared host pid namespace and SYS_PTRACE/SYS_ADMIN capabilities
	filter := bson.M{
		"$and": bson.A{
			bson.M{"inherited.host_pid": true},
			bson.M{"k8.securitycontext.capabilities.add": "SYS_PTRACE"},
			bson.M{"k8.securitycontext.capabilities.add": "SYS_ADMIN"},
			bson.M{
				"$or": bson.A{
					// Before Kubernetes 1.31, AppArmor is disabled by default so we don't need to check for it.
					// The AppArmor profile does not appear in the security context unless it is modified, so
					// we just check if it was disabled. See the CE_SYS_PTRACE attack doc for more details.
					bson.M{"k8.securitycontext.apparmorprofile.type": "Unconfined"},

					// AppArmor is enabled by default since Kubernetes 1.31
					bson.M{
						"$and": bson.A{
							// Ensure version fields exist and are not empty
							bson.M{"runtime.cluster.version_major": bson.M{"$exists": true, "$ne": ""}},
							bson.M{"runtime.cluster.version_minor": bson.M{"$exists": true, "$ne": ""}},
							// Numerical comparison using $expr
							bson.M{
								"$expr": bson.M{
									"$and": bson.A{
										bson.M{"$lte": bson.A{bson.M{"$toInt": "$runtime.cluster.version_major"}, AppArmorEnabledMajorVersion}},
										bson.M{"$lt": bson.A{bson.M{"$toInt": "$runtime.cluster.version_minor"}, AppArmorEnabledMinorVersion}},
									},
								},
							},
						},
					},

					// Technically true, but in this case the CE_NSENTER attack is also possible and easier to execute
					// bson.M{"k8.securitycontext.privileged": true},
				},
			},
		},
		"runtime.runID":        e.runtime.RunID.String(),
		"runtime.cluster.name": e.runtime.Cluster.Name,
	}

	// We just need a 1:1 mapping of the node and container to create this edge
	projection := bson.M{"_id": 1, "node_id": 1}

	cur, err := containers.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return adapter.MongoCursorHandler[containerEscapeGroup](ctx, cur, callback, complete)
}
