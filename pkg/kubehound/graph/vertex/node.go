package vertex

import (
	"database/sql"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
)

const (
	NodeLabel = "Node"
)

var _ Builder = (*Node)(nil)

type Node struct {
	BaseVertex
}

func (v *Node) Label() string {
	return NodeLabel
}

func (v *Node) Query(runID, clusterName string) string {
	return "SELECT id, name, is_namespaced, namespace, app, team, service, run_id, cluster_name FROM nodes WHERE run_id = '" + runID + "' AND cluster_name = '" + clusterName + "'"
}

func (v *Node) Scanner(rows *sql.Rows) (map[string]any, error) {
	var id int64
	var name, namespace, app, team, service, runID, cluster string
	var isNamespaced int
	if err := rows.Scan(&id, &name, &isNamespaced, &namespace, &app, &team, &service, &runID, &cluster); err != nil {
		return nil, err
	}
	return map[string]any{
		"storeID":      store.Hex(id),
		"name":         name,
		"isNamespaced": isNamespaced == 1,
		"namespace":    namespace,
		"compromised":  float64(shared.CompromiseNone),
		"critical":     false,
		"app":          app,
		"team":         team,
		"service":      service,
		"cluster":      cluster,
		"runID":        runID,
	}, nil
}

func (v *Node) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}
