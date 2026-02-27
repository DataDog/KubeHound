package edge

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
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

func (e *VolumeDiscover) processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*volumeMountGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Container, typed.Volume, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *VolumeDiscover) Stream(ctx context.Context, db *sql.DB, w types.EdgeWriter) error {
	oic := converter.NewObjectID(db)
	rows, err := db.QueryContext(ctx, `SELECT id, container_id FROM volumes WHERE run_id = ? AND cluster_name = ?`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var g volumeMountGroup
		if err := rows.Scan(&g.Volume, &g.Container); err != nil {
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
