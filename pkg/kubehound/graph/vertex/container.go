package vertex

import (
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
)

const (
	ContainerLabel = "Container"
)

var _ Builder = (*Container)(nil)

type Container struct {
	BaseVertex
}

func (v *Container) Label() string {
	return ContainerLabel
}

func (v *Container) Query(runID, clusterName string) string {
	return "SELECT id, name, image, command, args, capabilities_add, privileged, priv_esc, host_pid, host_ipc, host_network, run_as_user, ports, pod_name, node_name, namespace, app, team, service, run_id, cluster_name FROM containers WHERE run_id = '" + runID + "' AND cluster_name = '" + clusterName + "'"
}

func (v *Container) Scanner(rows *sql.Rows) (map[string]any, error) {
	var id, runAsUser int64
	var name, image, commandJSON, argsJSON, capsJSON, portsJSON string
	var podName, nodeName, namespace, app, team, service, runID, cluster string
	var privileged, privEsc, hostPID, hostIPC, hostNetwork int
	if err := rows.Scan(&id, &name, &image, &commandJSON, &argsJSON, &capsJSON,
		&privileged, &privEsc, &hostPID, &hostIPC, &hostNetwork, &runAsUser,
		&portsJSON, &podName, &nodeName, &namespace, &app, &team, &service, &runID, &cluster); err != nil {
		return nil, err
	}

	var command, args, capabilities []string
	var ports []int32
	_ = json.Unmarshal([]byte(commandJSON), &command)
	_ = json.Unmarshal([]byte(argsJSON), &args)
	_ = json.Unmarshal([]byte(capsJSON), &capabilities)
	_ = json.Unmarshal([]byte(portsJSON), &ports)

	if command == nil {
		command = []string{}
	}
	if args == nil {
		args = []string{}
	}
	if capabilities == nil {
		capabilities = []string{}
	}

	// Convert ports to string array
	portStrs := make([]string, 0, len(ports))
	for _, p := range ports {
		portStrs = append(portStrs, strconv.Itoa(int(p)))
	}

	return map[string]any{
		"storeID":      store.Hex(id),
		"name":         name,
		"image":        image,
		"command":       command,
		"args":          args,
		"capabilities":  capabilities,
		"privileged":    privileged == 1,
		"privesc":       privEsc == 1,
		"hostPid":       hostPID == 1,
		"hostIpc":       hostIPC == 1,
		"hostNetwork":   hostNetwork == 1,
		"runAsUser":     runAsUser,
		"ports":         portStrs,
		"pod":           podName,
		"node":          nodeName,
		"isNamespaced":  namespace != "",
		"namespace":     namespace,
		"compromised":   float64(shared.CompromiseNone),
		"app":           app,
		"team":          team,
		"service":       service,
		"cluster":       cluster,
		"runID":         runID,
	}, nil
}

func (v *Container) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}
