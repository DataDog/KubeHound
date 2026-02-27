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
	Register(&TokenSteal{}, RegisterDefault)
}

type tokenStealGroup struct {
	Volume   int64 `json:"volume"`
	Identity int64 `json:"identity"`
}

type TokenSteal struct {
	BaseEdge
}

func (e *TokenSteal) Label() string {
	return "TOKEN_STEAL"
}

func (e *TokenSteal) Name() string {
	return "TokenSteal"
}

func (e *TokenSteal) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueUnsecuredCredentials
}

func (e *TokenSteal) AttckTacticID() AttckTacticID {
	return AttckTacticCredentialAccess
}

func (e *TokenSteal) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*tokenStealGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Volume, typed.Identity, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

func (e *TokenSteal) Stream(ctx context.Context, _ storedb.Provider, db *sql.DB,
	process types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	rows, err := db.QueryContext(ctx, `SELECT id, projected_id FROM volumes WHERE type = 'Projected' AND projected_id IS NOT NULL AND projected_id != 0 AND run_id = ? AND cluster_name = ?`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}

	return adapter.SQLiteRowHandler[tokenStealGroup](ctx, rows, func(row *sql.Rows) (tokenStealGroup, error) {
		var g tokenStealGroup
		err := row.Scan(&g.Volume, &g.Identity)
		return g, err
	}, process, complete)
}
