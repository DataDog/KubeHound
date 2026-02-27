package vertex

import (
	"database/sql"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
)

const (
	IdentityLabel = "Identity"
)

var _ Builder = (*Identity)(nil)

type Identity struct {
	BaseVertex
}

func (v *Identity) Label() string {
	return IdentityLabel
}

func (v *Identity) Query(runID, clusterName string) string {
	return "SELECT id, name, is_namespaced, namespace, type, app, team, service, run_id, cluster_name FROM identities WHERE run_id = '" + runID + "' AND cluster_name = '" + clusterName + "'"
}

func (v *Identity) Scanner(rows *sql.Rows) (map[string]any, error) {
	var id int64
	var name, namespace, idType, app, team, service, runID, cluster string
	var isNamespaced int
	if err := rows.Scan(&id, &name, &isNamespaced, &namespace, &idType, &app, &team, &service, &runID, &cluster); err != nil {
		return nil, err
	}
	return map[string]any{
		"storeID":      store.Hex(id),
		"name":         name,
		"isNamespaced": isNamespaced == 1,
		"namespace":    namespace,
		"type":         idType,
		"critical":     false,
		"app":          app,
		"team":         team,
		"service":      service,
		"cluster":      cluster,
		"runID":        runID,
	}, nil
}

func (v *Identity) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}
