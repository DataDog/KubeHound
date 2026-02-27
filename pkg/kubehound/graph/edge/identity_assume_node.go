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
	Register(&IdentityAssumeNode{}, RegisterDefault)
}

type IdentityAssumeNode struct {
	BaseEdge
}

type nodeIdentityGroup struct {
	Node     int64 `json:"node"`
	Identity int64 `json:"user_id"`
}

func (e *IdentityAssumeNode) Label() string {
	return "IDENTITY_ASSUME"
}

func (e *IdentityAssumeNode) Name() string {
	return "IdentityAssumeNode"
}

func (e *IdentityAssumeNode) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueValidAccounts
}

func (e *IdentityAssumeNode) AttckTacticID() AttckTacticID {
	return AttckTacticPrivilegeEscalation
}

func (e *IdentityAssumeNode) processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*nodeIdentityGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Node, typed.Identity, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *IdentityAssumeNode) Stream(ctx context.Context, db *sql.DB, w types.EdgeWriter) error {
	oic := converter.NewObjectID(db)
	rows, err := db.QueryContext(ctx, `SELECT id, user_id FROM nodes WHERE user_id != 0 AND run_id = ? AND cluster_name = ?`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var g nodeIdentityGroup
		if err := rows.Scan(&g.Node, &g.Identity); err != nil {
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
