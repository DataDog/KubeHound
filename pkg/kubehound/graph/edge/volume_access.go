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
	Register(&VolumeAccess{}, RegisterDefault)
}

type VolumeAccess struct {
	BaseEdge
}

type volumeAccessGroup struct {
	Volume int64 `json:"volume"`
	Node   int64 `json:"node"`
}

func (e *VolumeAccess) Label() string {
	return "VOLUME_ACCESS"
}

func (e *VolumeAccess) Name() string {
	return "VolumeAccess"
}

func (e *VolumeAccess) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueContainerAndResourceDiscovery
}

func (e *VolumeAccess) AttckTacticID() AttckTacticID {
	return AttckTacticDiscovery
}

func (e *VolumeAccess) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*volumeAccessGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Node, typed.Volume, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *VolumeAccess) Stream(ctx context.Context, _ storedb.Provider, db *sql.DB,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	rows, err := db.QueryContext(ctx, `SELECT id, node_id FROM volumes WHERE run_id = ? AND cluster_name = ?`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}

	return adapter.SQLiteRowHandler[volumeAccessGroup](ctx, rows, func(row *sql.Rows) (volumeAccessGroup, error) {
		var g volumeAccessGroup
		err := row.Scan(&g.Volume, &g.Node)
		return g, err
	}, callback, complete)
}
