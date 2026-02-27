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

func (e *PermissionDiscover) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*permissionDiscoverGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Identity, typed.PermissionSet, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *PermissionDiscover) Stream(ctx context.Context, _ storedb.Provider, db *sql.DB,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

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

	return adapter.SQLiteRowHandler[permissionDiscoverGroup](ctx, rows, func(row *sql.Rows) (permissionDiscoverGroup, error) {
		var g permissionDiscoverGroup
		err := row.Scan(&g.PermissionSet, &g.Identity)
		return g, err
	}, callback, complete)
}
