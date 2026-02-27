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

func (e *EscapeCorePattern) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	return containerEscapeProcessor(ctx, oic, e.Label(), entry, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *EscapeCorePattern) Stream(ctx context.Context, _ storedb.Provider, db *sql.DB,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

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

	return adapter.SQLiteRowHandler[containerEscapeGroup](ctx, rows, func(row *sql.Rows) (containerEscapeGroup, error) {
		var g containerEscapeGroup
		err := row.Scan(&g.Container, &g.Node)
		return g, err
	}, callback, complete)
}
