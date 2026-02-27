package edge

import (
	"context"
	"database/sql"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
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

// Processor delegates the processing tasks to the generic containerEscapeProcessor.
func (e *EscapeModuleLoad) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return containerEscapeProcessor(ctx, oic, e.Label(), entry, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *EscapeModuleLoad) Stream(ctx context.Context, _ storedb.Provider, db *sql.DB,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	rows, err := db.QueryContext(ctx,
		`SELECT id, node_id FROM containers WHERE (privileged = 1 OR EXISTS (SELECT 1 FROM json_each(capabilities_add) WHERE value = 'SYS_MODULE')) AND run_id = ? AND cluster_name = ?`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}

	return adapter.SQLiteRowHandler[containerEscapeGroup](ctx, rows, func(row *sql.Rows) (containerEscapeGroup, error) {
		var g containerEscapeGroup
		err := row.Scan(&g.Container, &g.Node)
		return g, err
	}, callback, complete)
}
