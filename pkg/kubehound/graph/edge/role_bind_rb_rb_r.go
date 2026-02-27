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

const (
	RoleBindspaceName = "RoleBindRoleBindingbRoleBindingRole"
)

func init() {
	Register(&RoleBindRbRbR{}, RegisterDefault)
}

type RoleBindRbRbR struct {
	BaseEdge
}

type roleBindNameSpaceGroup struct {
	FromPerm int64 `json:"from_permission_set"`
	ToPerm   int64 `json:"to_permission_set"`
}

func (e *RoleBindRbRbR) Label() string {
	return RoleBindLabel
}

func (e *RoleBindRbRbR) Name() string {
	return RoleBindspaceName
}

func (e *RoleBindRbRbR) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueValidAccounts
}

func (e *RoleBindRbRbR) AttckTacticID() AttckTacticID {
	return AttckTacticPrivilegeEscalation
}

func (e *RoleBindRbRbR) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*roleBindNameSpaceGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.FromPerm, typed.ToPerm, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *RoleBindRbRbR) Stream(ctx context.Context, _ storedb.Provider, db *sql.DB,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT src.id, tgt.id
		FROM permissionsets src
		JOIN rolebindings rb ON rb.id = src.role_binding_id
			AND rb.run_id = src.run_id
			AND rb.cluster_name = src.cluster_name
		JOIN permissionsets tgt ON (tgt.namespace = src.namespace OR tgt.is_namespaced = 1) AND tgt.run_id = src.run_id AND tgt.cluster_name = src.cluster_name AND tgt.id != src.id
		WHERE src.is_namespaced = 1 AND rb.is_namespaced = 1 AND src.run_id = ? AND src.cluster_name = ?
		AND EXISTS (
			SELECT 1 FROM json_each(src.rules) AS r1
			WHERE EXISTS (SELECT 1 FROM json_each(json_extract(r1.value, '$.apiGroups')) WHERE value IN ('*', 'rbac.authorization.k8s.io'))
		)
		AND EXISTS (
			SELECT 1 FROM json_each(src.rules) AS r2
			WHERE EXISTS (SELECT 1 FROM json_each(json_extract(r2.value, '$.verbs')) WHERE value IN ('create', '*'))
			AND EXISTS (SELECT 1 FROM json_each(json_extract(r2.value, '$.resources')) WHERE value IN ('rolebindings', '*'))
			AND (json_extract(r2.value, '$.resourceNames') IS NULL OR json_extract(r2.value, '$.resourceNames') = '[]')
		)
		AND EXISTS (
			SELECT 1 FROM json_each(src.rules) AS r3
			WHERE EXISTS (SELECT 1 FROM json_each(json_extract(r3.value, '$.verbs')) WHERE value IN ('bind', '*'))
			AND EXISTS (SELECT 1 FROM json_each(json_extract(r3.value, '$.resources')) WHERE value IN ('roles', '*'))
			AND (json_extract(r3.value, '$.resourceNames') IS NULL OR json_extract(r3.value, '$.resourceNames') = '[]')
		)`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}

	return adapter.SQLiteRowHandler[roleBindNameSpaceGroup](ctx, rows, func(row *sql.Rows) (roleBindNameSpaceGroup, error) {
		var g roleBindNameSpaceGroup
		err := row.Scan(&g.FromPerm, &g.ToPerm)
		return g, err
	}, callback, complete)
}
