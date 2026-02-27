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
	Register(&PermissionDiscover{}, RegisterDefault)
}

type PermissionDiscover struct {
	BaseEdge
}

type permissionDiscoverGroup struct {
	PermissionSet int64 `json:"permission_set"`
	Identity      int64 `json:"identity"`
}

func (e *PermissionDiscover) Label() string {
	return "PERMISSION_DISCOVER"
}

func (e *PermissionDiscover) Name() string {
	return "PermissionDiscover"
}

func (e *PermissionDiscover) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniquePermissionGroupsDiscovery
}

func (e *PermissionDiscover) AttckTacticID() AttckTacticID {
	return AttckTacticDiscovery
}

func (e *PermissionDiscover) processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*permissionDiscoverGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Identity, typed.PermissionSet, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *PermissionDiscover) Stream(ctx context.Context, db *sql.DB, w types.EdgeWriter) error {
	oic := converter.NewObjectID(db)
	rows, err := db.QueryContext(ctx, `
		SELECT ps.id, CAST(json_extract(sub.value, '$.identity_id') AS INTEGER)
		FROM permissionsets ps
		JOIN rolebindings rb ON rb.id = ps.role_binding_id
			AND rb.run_id = ps.run_id
			AND rb.cluster_name = ps.cluster_name,
		json_each(rb.subjects) AS sub
		WHERE ps.run_id = ? AND ps.cluster_name = ?
		AND (
			(ps.is_namespaced = 1 AND ps.namespace = json_extract(sub.value, '$.subject.namespace'))
			OR json_extract(sub.value, '$.subject.namespace') = ''
			OR ps.is_namespaced = 0
		)`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var g permissionDiscoverGroup
		if err := rows.Scan(&g.PermissionSet, &g.Identity); err != nil {
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
