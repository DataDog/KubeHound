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
	Register(&VolumeDiscover{}, RegisterDefault)
}

type VolumeDiscover struct {
	BaseEdge
}

type volumeMountGroup struct {
	Volume    int64 `json:"volume"`
	Container int64 `json:"container"`
}

func (e *VolumeDiscover) Label() string {
	return "VOLUME_DISCOVER"
}

func (e *VolumeDiscover) Name() string {
	return "VolumeDiscover"
}

func (e *VolumeDiscover) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueContainerAndResourceDiscovery
}

func (e *VolumeDiscover) AttckTacticID() AttckTacticID {
	return AttckTacticDiscovery
}

func (e *VolumeDiscover) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*volumeMountGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Container, typed.Volume, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *VolumeDiscover) Stream(ctx context.Context, _ storedb.Provider, db *sql.DB,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	rows, err := db.QueryContext(ctx, `SELECT id, container_id FROM volumes WHERE run_id = ? AND cluster_name = ?`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}

	return adapter.SQLiteRowHandler[volumeMountGroup](ctx, rows, func(row *sql.Rows) (volumeMountGroup, error) {
		var g volumeMountGroup
		err := row.Scan(&g.Volume, &g.Container)
		return g, err
	}, callback, complete)
}
