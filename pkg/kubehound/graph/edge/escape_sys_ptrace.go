package edge

import (
	"context"
	"database/sql"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
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

// processor delegates the processing tasks to the generic containerEscapeProcessor.
func (e *EscapeSysPtrace) processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return containerEscapeProcessor(ctx, oic, e.Label(), entry, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *EscapeSysPtrace) Stream(ctx context.Context, db *sql.DB, w types.EdgeWriter) error {
	oic := converter.NewObjectID(db)
	rows, err := db.QueryContext(ctx, `
		SELECT id, node_id FROM containers
		WHERE host_pid = 1
		AND EXISTS (SELECT 1 FROM json_each(capabilities_add) WHERE value = 'SYS_PTRACE')
		AND EXISTS (SELECT 1 FROM json_each(capabilities_add) WHERE value = 'SYS_ADMIN')
		AND run_id = ? AND cluster_name = ?`,
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
