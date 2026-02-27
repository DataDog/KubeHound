package vertex

import (
	"database/sql"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
)

const (
	PodLabel = "Pod"
)

var _ Builder = (*Pod)(nil)

type Pod struct {
	BaseVertex
}

func (v *Pod) Label() string {
	return PodLabel
}

func (v *Pod) Query(runID, clusterName string) string {
	return "SELECT id, name, is_namespaced, namespace, share_process_namespace, service_account, node_name, app, team, service, run_id, cluster_name FROM pods WHERE run_id = '" + runID + "' AND cluster_name = '" + clusterName + "'"
}

func (v *Pod) Scanner(rows *sql.Rows) (map[string]any, error) {
	var id int64
	var name, namespace, serviceAccount, nodeName, app, team, service, runID, cluster string
	var isNamespaced, shareProcessNamespace int
	if err := rows.Scan(&id, &name, &isNamespaced, &namespace, &shareProcessNamespace, &serviceAccount, &nodeName, &app, &team, &service, &runID, &cluster); err != nil {
		return nil, err
	}
	return map[string]any{
		"storeID":               store.Hex(id),
		"name":                  name,
		"isNamespaced":          isNamespaced == 1,
		"namespace":             namespace,
		"shareProcessNamespace": shareProcessNamespace == 1,
		"serviceAccount":        serviceAccount,
		"node":                  nodeName,
		"compromised":           float64(shared.CompromiseNone),
		"critical":              false,
		"app":                   app,
		"team":                  team,
		"service":               service,
		"cluster":               cluster,
		"runID":                 runID,
	}, nil
}

func (v *Pod) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}
