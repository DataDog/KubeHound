package vertex

import (
	"database/sql"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
)

const (
	VolumeLabel = "Volume"
)

var _ Builder = (*Volume)(nil)

type Volume struct {
	BaseVertex
}

func (v *Volume) Label() string {
	return VolumeLabel
}

func (v *Volume) Query(runID, clusterName string) string {
	return "SELECT v.id, v.name, v.type, v.source, v.mount, v.readonly, p.namespace, v.app, v.team, v.service, v.run_id, v.cluster_name FROM volumes v JOIN pods p ON v.pod_id = p.id AND v.run_id = p.run_id AND v.cluster_name = p.cluster_name WHERE v.run_id = '" + runID + "' AND v.cluster_name = '" + clusterName + "'"
}

func (v *Volume) Scanner(rows *sql.Rows) (map[string]any, error) {
	var id int64
	var name, volType, sourcePath, mountPath, namespace, app, team, service, runID, cluster string
	var readonly int
	if err := rows.Scan(&id, &name, &volType, &sourcePath, &mountPath, &readonly, &namespace, &app, &team, &service, &runID, &cluster); err != nil {
		return nil, err
	}
	return map[string]any{
		"storeID":      store.Hex(id),
		"name":         name,
		"type":         volType,
		"sourcePath":   sourcePath,
		"mountPath":    mountPath,
		"readonly":     readonly == 1,
		"isNamespaced": namespace != "",
		"namespace":    namespace,
		"app":          app,
		"team":         team,
		"service":      service,
		"cluster":      cluster,
		"runID":        runID,
	}, nil
}

func (v *Volume) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}
