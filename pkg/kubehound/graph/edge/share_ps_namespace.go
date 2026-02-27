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
	Register(&SharePSNamespace{}, RegisterDefault)
}

type SharePSNamespace struct {
	BaseEdge
}

type sharedPsNamespaceGroupPair struct {
	ContainerA int64 `json:"container_a"`
	ContainerB int64 `json:"container_b"`
}

func (e *SharePSNamespace) Label() string {
	return "SHARE_PS_NAMESPACE"
}

func (e *SharePSNamespace) Name() string {
	return "SharePSNamespace"
}

func (e *SharePSNamespace) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueTaintedSharedContent
}

func (e *SharePSNamespace) AttckTacticID() AttckTacticID {
	return AttckTacticLateralMovement
}

// Processor delegates the processing tasks to the generic containerEscapeProcessor.
func (e *SharePSNamespace) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*sharedPsNamespaceGroupPair)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.ContainerA, typed.ContainerB, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *SharePSNamespace) Stream(ctx context.Context, _ storedb.Provider, db *sql.DB,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	rows, err := db.QueryContext(ctx, `
		SELECT ca.id, cb.id
		FROM pods p
		JOIN containers ca ON ca.pod_id = p.id
			AND ca.run_id = p.run_id
			AND ca.cluster_name = p.cluster_name
		JOIN containers cb ON cb.pod_id = p.id
			AND cb.run_id = p.run_id
			AND cb.cluster_name = p.cluster_name
			AND cb.id != ca.id
		WHERE p.share_process_namespace = 1
		AND p.run_id = ? AND p.cluster_name = ?`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}

	return adapter.SQLiteRowHandler[sharedPsNamespaceGroupPair](ctx, rows, func(row *sql.Rows) (sharedPsNamespaceGroupPair, error) {
		var g sharedPsNamespaceGroupPair
		err := row.Scan(&g.ContainerA, &g.ContainerB)
		return g, err
	}, callback, complete)
}
