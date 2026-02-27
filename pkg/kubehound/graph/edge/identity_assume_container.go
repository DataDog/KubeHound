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
	Register(&IdentityAssumeContainer{}, RegisterDefault)
}

type IdentityAssumeContainer struct {
	BaseEdge
}

type containerIdentityGroup struct {
	Container int64 `json:"container"`
	Identity  int64 `json:"identity"`
}

func (e *IdentityAssumeContainer) Label() string {
	return "IDENTITY_ASSUME"
}

func (e *IdentityAssumeContainer) Name() string {
	return "IdentityAssumeContainer"
}

func (e *IdentityAssumeContainer) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueValidAccounts
}

func (e *IdentityAssumeContainer) AttckTacticID() AttckTacticID {
	return AttckTacticPrivilegeEscalation
}

func (e *IdentityAssumeContainer) processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*containerIdentityGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Container, typed.Identity, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *IdentityAssumeContainer) Stream(ctx context.Context, db *sql.DB, w types.EdgeWriter) error {
	oic := converter.NewObjectID(db)
	rows, err := db.QueryContext(ctx, `
		SELECT c.id, i.id
		FROM containers c
		JOIN identities i ON i.name = c.service_account
			AND i.namespace = c.namespace
			AND i.type = 'ServiceAccount'
			AND i.run_id = c.run_id
			AND i.cluster_name = c.cluster_name
		WHERE c.run_id = ? AND c.cluster_name = ?`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var g containerIdentityGroup
		if err := rows.Scan(&g.Container, &g.Identity); err != nil {
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
