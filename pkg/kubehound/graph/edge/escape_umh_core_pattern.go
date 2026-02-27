package edge

import (
	"context"
	"database/sql"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
)

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

func (e *EscapeCorePattern) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueEscapeToHost
}

func (e *EscapeCorePattern) AttckTacticID() AttckTacticID {
	return AttckTacticPrivilegeEscalation
}

func (e *EscapeCorePattern) processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return containerEscapeProcessor(ctx, oic, e.Label(), entry, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *EscapeCorePattern) Stream(ctx context.Context, db *sql.DB, w types.EdgeWriter) error {
	oic := converter.NewObjectID(db)
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT c.id, c.node_id
		FROM containers c
		JOIN volumes v ON v.pod_id = c.pod_id
			AND v.run_id = c.run_id
			AND v.cluster_name = c.cluster_name
		WHERE c.run_as_user = 0
		AND v.type = 'HostPath'
		AND v.source IN ('/', '/proc', '/proc/sys', '/proc/sys/kernel')
		AND c.run_id = ? AND c.cluster_name = ?`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var g containerEscapeGroup
		if err := rows.Scan(&g.Container, &g.Node); err != nil {
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
