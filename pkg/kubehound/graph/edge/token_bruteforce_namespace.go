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
	Register(&TokenBruteforceNamespace{}, RegisterDefault)
}

type TokenBruteforceNamespace struct {
	BaseEdge
}

type tokenBruteforceNSGroup struct {
	Role     int64 `json:"role"`
	Identity int64 `json:"identity"`
}

func (e *TokenBruteforceNamespace) Label() string {
	return "TOKEN_BRUTEFORCE"
}

func (e *TokenBruteforceNamespace) Name() string {
	return "TokenBruteforceNamespace"
}

func (e *TokenBruteforceNamespace) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueStealApplicationAccessTokens
}

func (e *TokenBruteforceNamespace) AttckTacticID() AttckTacticID {
	return AttckTacticCredentialAccess
}

func (e *TokenBruteforceNamespace) processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*tokenBruteforceNSGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Role, typed.Identity, map[string]any{
		"attckTechniqueID": string(e.AttckTechniqueID()),
		"attckTacticID":    string(e.AttckTacticID()),
	})
}

// Stream finds all roles that are namespaced and have secrets/get or equivalent wildcard permissions and matching identities.
// Matching identities are defined as namespaced identities that share the role namespace or non-namespaced identities.
func (e *TokenBruteforceNamespace) Stream(ctx context.Context, db *sql.DB, w types.EdgeWriter) error {
	oic := converter.NewObjectID(db)
	var query string
	if e.cfg.LargeClusterOptimizations {
		// For large clusters do not create a redundant edge already covered by the TOKEN_LIST attack as this technique is much more complex
		query = `
		SELECT DISTINCT ps.id, i.id FROM permissionsets ps, json_each(ps.rules) AS r
		JOIN identities i ON (i.namespace = ps.namespace OR i.is_namespaced = 0) AND i.type = 'ServiceAccount' AND i.run_id = ps.run_id AND i.cluster_name = ps.cluster_name
		WHERE ps.is_namespaced = 1 AND ps.run_id = ? AND ps.cluster_name = ?
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.apiGroups')) WHERE value IN ('', '*'))
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.resources')) WHERE value IN ('secrets', '*'))
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.verbs')) WHERE value IN ('get'))
		AND (json_extract(r.value, '$.resourceNames') IS NULL OR json_extract(r.value, '$.resourceNames') = '[]')`
	} else {
		query = `
		SELECT DISTINCT ps.id, i.id FROM permissionsets ps, json_each(ps.rules) AS r
		JOIN identities i ON (i.namespace = ps.namespace OR i.is_namespaced = 0) AND i.type = 'ServiceAccount' AND i.run_id = ps.run_id AND i.cluster_name = ps.cluster_name
		WHERE ps.is_namespaced = 1 AND ps.run_id = ? AND ps.cluster_name = ?
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.apiGroups')) WHERE value IN ('', '*'))
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.resources')) WHERE value IN ('secrets', '*'))
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.verbs')) WHERE value IN ('get', '*'))
		AND (json_extract(r.value, '$.resourceNames') IS NULL OR json_extract(r.value, '$.resourceNames') = '[]')`
	}

	rows, err := db.QueryContext(ctx, query, e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var g tokenBruteforceNSGroup
		if err := rows.Scan(&g.Role, &g.Identity); err != nil {
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
