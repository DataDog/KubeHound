package edge

import (
	"context"
	"database/sql"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
)

func init() {
	Register(&EscapeModuleLoad{}, RegisterDefault)
}

type EscapeModuleLoad struct {
	BaseContainerEscape
}

func (e *EscapeModuleLoad) Label() string {
	return "CE_MODULE_LOAD"
}

func (e *EscapeModuleLoad) Name() string {
	return "ContainerEscapeModuleLoad"
}

func (e *EscapeModuleLoad) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueEscapeToHost
}

func (e *EscapeModuleLoad) AttckTacticID() AttckTacticID {
	return AttckTacticPrivilegeEscalation
}

// processor delegates the processing tasks to the generic containerEscapeProcessor.
func (e *EscapeModuleLoad) processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return containerEscapeProcessor(ctx, oic, e.Label(), entry, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *EscapeModuleLoad) Stream(ctx context.Context, db *sql.DB, w types.EdgeWriter) error {
	oic := converter.NewObjectID(db)
	rows, err := db.QueryContext(ctx,
		`SELECT id, node_id FROM containers WHERE (privileged = 1 OR EXISTS (SELECT 1 FROM json_each(capabilities_add) WHERE value = 'SYS_MODULE')) AND run_id = ? AND cluster_name = ?`,
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
