package edge

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
)

func init() {
	Register(&PodExecNamespace{}, RegisterDefault)
}

type PodExecNamespace struct {
	BaseEdge
}

type podExecNSGroup struct {
	Role int64 `json:"role"`
	Pod  int64 `json:"pod"`
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

func (e *PodExecNamespace) processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
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
func (e *PodExecNamespace) Stream(ctx context.Context, db *sql.DB, w types.EdgeWriter) error {
	oic := converter.NewObjectID(db)
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT ps.id, p.id FROM permissionsets ps, json_each(ps.rules) AS r
		JOIN pods p ON (p.namespace = ps.namespace OR p.is_namespaced = 0) AND p.run_id = ps.run_id AND p.cluster_name = ps.cluster_name
		WHERE ps.is_namespaced = 1 AND ps.run_id = ? AND ps.cluster_name = ?
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.apiGroups')) WHERE value IN ('', '*'))
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.resources')) WHERE value IN ('pods/exec', 'pods/*', '*'))
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.verbs')) WHERE value IN ('create', '*'))
		AND (json_extract(r.value, '$.resourceNames') IS NULL OR json_extract(r.value, '$.resourceNames') = '[]')`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var g podExecNSGroup
		if err := rows.Scan(&g.Role, &g.Pod); err != nil {
			return err
		}
		insert, err := e.processor(ctx, oic, &g)
		if err != nil {
			return err
		}
		if err := w.Queue(ctx, insert); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return w.Flush(ctx)
}
