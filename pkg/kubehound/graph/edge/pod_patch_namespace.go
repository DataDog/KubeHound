package edge

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
)

func init() {
	Register(&PodPatchNamespace{}, RegisterDefault)
}

type PodPatchNamespace struct {
	BaseEdge
}

type podPatchNSGroup struct {
	Role int64 `json:"role"`
	Pod  int64 `json:"pod"`
}

func (e *PodPatchNamespace) Label() string {
	return "POD_PATCH"
}

func (e *PodPatchNamespace) Name() string {
	return "PodPatchNamespace"
}

func (e *PodPatchNamespace) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueContainerAdministrationCommand
}

func (e *PodPatchNamespace) AttckTacticID() AttckTacticID {
	return AttckTacticExecution
}

func (e *PodPatchNamespace) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*podPatchNSGroup)
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
func (e *PodPatchNamespace) Stream(ctx context.Context, _ storedb.Provider, db *sql.DB,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT ps.id, p.id FROM permissionsets ps, json_each(ps.rules) AS r
		JOIN pods p ON (p.namespace = ps.namespace OR p.is_namespaced = 0) AND p.run_id = ps.run_id AND p.cluster_name = ps.cluster_name
		WHERE ps.is_namespaced = 1 AND ps.run_id = ? AND ps.cluster_name = ?
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.apiGroups')) WHERE value IN ('', '*'))
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.resources')) WHERE value IN ('pods', 'cronjobs', 'daemonsets', 'deployments', 'jobs', 'replicasets', 'replicationcontrollers', 'statefulsets', '*'))
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.verbs')) WHERE value IN ('patch', 'update', '*'))
		AND (json_extract(r.value, '$.resourceNames') IS NULL OR json_extract(r.value, '$.resourceNames') = '[]')`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}

	return adapter.SQLiteRowHandler[podPatchNSGroup](ctx, rows, func(row *sql.Rows) (podPatchNSGroup, error) {
		var g podPatchNSGroup
		err := row.Scan(&g.Role, &g.Pod)
		return g, err
	}, callback, complete)
}
