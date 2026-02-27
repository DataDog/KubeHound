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
	Register(&PodAttach{}, RegisterDefault)
}

type PodAttach struct {
	BaseEdge
}

type podAttachGroup struct {
	Pod  int64 `json:"pod"`
	Node int64 `json:"node"`
}

func (e *PodAttach) Label() string {
	return "POD_ATTACH"
}

func (e *PodAttach) Name() string {
	return "PodAttach"
}

func (e *PodAttach) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueContainerAdministrationCommand
}

func (e *PodAttach) AttckTacticID() AttckTacticID {
	return AttckTacticExecution
}

func (e *PodAttach) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*podAttachGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Node, typed.Pod, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *PodAttach) Stream(ctx context.Context, _ storedb.Provider, db *sql.DB,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	rows, err := db.QueryContext(ctx, `SELECT id, node_id FROM pods WHERE run_id = ? AND cluster_name = ?`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}

	return adapter.SQLiteRowHandler[podAttachGroup](ctx, rows, func(row *sql.Rows) (podAttachGroup, error) {
		var g podAttachGroup
		err := row.Scan(&g.Pod, &g.Node)
		return g, err
	}, callback, complete)
}
