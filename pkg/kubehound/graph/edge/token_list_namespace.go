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
	Register(&TokenListNamespace{}, RegisterDefault)
}

type TokenListNamespace struct {
	BaseEdge
}

type tokenListNSGroup struct {
	Role     int64 `json:"role"`
	Identity int64 `json:"identity"`
}

func (e *TokenListNamespace) Label() string {
	return "TOKEN_LIST"
}

func (e *TokenListNamespace) Name() string {
	return "TokenListNamespace"
}

func (e *TokenListNamespace) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueStealApplicationAccessTokens
}

func (e *TokenListNamespace) AttckTacticID() AttckTacticID {
	return AttckTacticCredentialAccess
}

func (e *TokenListNamespace) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*tokenListNSGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Role, typed.Identity, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

// Stream finds all roles that are namespaced and have secrets/list or equivalent wildcard permissions and matching identities.
// Matching identities are defined as namespaced identities that share the role namespace or non-namespaced identities.
func (e *TokenListNamespace) Stream(ctx context.Context, _ storedb.Provider, db *sql.DB,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT ps.id, i.id FROM permissionsets ps, json_each(ps.rules) AS r
		JOIN identities i ON (i.namespace = ps.namespace OR i.is_namespaced = 0) AND i.type = 'ServiceAccount' AND i.run_id = ps.run_id AND i.cluster_name = ps.cluster_name
		WHERE ps.is_namespaced = 1 AND ps.run_id = ? AND ps.cluster_name = ?
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.apiGroups')) WHERE value IN ('', '*'))
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.resources')) WHERE value IN ('secrets', '*'))
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.verbs')) WHERE value IN ('list', '*'))
		AND (json_extract(r.value, '$.resourceNames') IS NULL OR json_extract(r.value, '$.resourceNames') = '[]')`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}

	return adapter.SQLiteRowHandler[tokenListNSGroup](ctx, rows, func(row *sql.Rows) (tokenListNSGroup, error) {
		var g tokenListNSGroup
		err := row.Scan(&g.Role, &g.Identity)
		return g, err
	}, callback, complete)
}
