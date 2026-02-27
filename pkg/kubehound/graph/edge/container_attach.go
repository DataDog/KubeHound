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
	Register(&ContainerAttach{}, RegisterDefault)
}

type ContainerAttach struct {
	BaseEdge
}

type containerAttachGroup struct {
	Container int64 `json:"container"`
	Pod       int64 `json:"pod"`
}

func (e *ContainerAttach) Label() string {
	return "CONTAINER_ATTACH"
}

func (e *ContainerAttach) Name() string {
	return "ContainerAttach"
}

func (e *ContainerAttach) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueContainerAdministrationCommand
}

func (e *ContainerAttach) AttckTacticID() AttckTacticID {
	return AttckTacticExecution
}

func (e *ContainerAttach) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*containerAttachGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Pod, typed.Container, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *ContainerAttach) Traversal() types.EdgeTraversal {
	return adapter.DefaultEdgeTraversal()
}

func (e *ContainerAttach) Stream(ctx context.Context, _ storedb.Provider, db *sql.DB,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	rows, err := db.QueryContext(ctx, `SELECT id, pod_id FROM containers WHERE run_id = ? AND cluster_name = ?`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}

	return adapter.SQLiteRowHandler[containerAttachGroup](ctx, rows, func(row *sql.Rows) (containerAttachGroup, error) {
		var g containerAttachGroup
		err := row.Scan(&g.Container, &g.Pod)
		return g, err
	}, callback, complete)
}
