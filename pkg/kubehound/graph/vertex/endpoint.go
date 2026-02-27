package vertex

import (
	"database/sql"
	"encoding/json"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
)

const (
	EndpointLabel = "Endpoint"
)

var _ Builder = (*Endpoint)(nil)

type Endpoint struct {
	BaseVertex
}

func (v *Endpoint) Label() string {
	return EndpointLabel
}

func (v *Endpoint) Query(runID, clusterName string) string {
	return "SELECT id, is_namespaced, namespace, name, service_name, service_dns, address_type, addresses, port, port_name, protocol, exposure, app, team, service, run_id, cluster_name FROM endpoints WHERE run_id = '" + runID + "' AND cluster_name = '" + clusterName + "'"
}

func (v *Endpoint) Scanner(rows *sql.Rows) (map[string]any, error) {
	var id int64
	var name, namespace, serviceName, serviceDns, addressType, addressesJSON, portName, protocol, app, team, service, runID, cluster string
	var port, exposure, isNamespaced int
	if err := rows.Scan(&id, &isNamespaced, &namespace, &name, &serviceName, &serviceDns, &addressType, &addressesJSON, &port, &portName, &protocol, &exposure, &app, &team, &service, &runID, &cluster); err != nil {
		return nil, err
	}

	var addresses []string
	_ = json.Unmarshal([]byte(addressesJSON), &addresses)
	if addresses == nil {
		addresses = []string{}
	}

	return map[string]any{
		"storeID":         store.Hex(id),
		"isNamespaced":    isNamespaced == 1,
		"namespace":       namespace,
		"name":            name,
		"serviceEndpoint": serviceName,
		"serviceDns":      serviceDns,
		"addressType":     addressType,
		"addresses":       addresses,
		"port":            port,
		"portName":        portName,
		"protocol":        protocol,
		"exposure":        shared.EndpointExposureType(exposure),
		"compromised":     float64(shared.CompromiseNone),
		"app":             app,
		"team":            team,
		"service":         service,
		"cluster":         cluster,
		"runID":           runID,
	}, nil
}

func (v *Endpoint) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}
