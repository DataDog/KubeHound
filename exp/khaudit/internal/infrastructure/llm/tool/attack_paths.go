package tool

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/container"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/kubehound"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func ceAttackPaths(containers container.Reader) server.ServerTool {
	type args struct {
		Cluster   string `json:"cluster"`
		RunID     string `json:"runID"`
		Namespace string `json:"namespace"`
		Image     string `json:"image"`
	}

	return server.ServerTool{
		Tool: mcp.NewTool(
			"kh_container_escape_paths",
			mcp.WithDescription(`List container escape paths that could be used to reach a kubernetes Node.
	
	The following is a list of nodes and edges that can be used to build a container escape profile.
	List of nodes:
	- Container: A container image running on a Kubernetes pod. Containers in a Pod are co-located and co-scheduled to run on the same node.
	- Endpoint: A network endpoint exposed by a container accessible via a Kubernetes service, external node port or cluster IP/port tuple.
	- Identity: Identity represents a Kubernetes user or service account.
	- Node: A Kubernetes node. Kubernetes runs workloads by placing containers into Pods to run on Nodes. A node may be a virtual or physical machine, depending on the cluster.
	- PermissionSet: A permission set represents a Kubernetes RBAC Role or ClusterRole, which contain rules that represent a set of permissions that has been bound to an identity via a RoleBinding or ClusterRoleBinding. Permissions are purely additive (there are no "deny" rules).
	- Pod: A Kubernetes pod - the smallest deployable units of computing that you can create and manage in Kubernetes.
	- Volume: Volume represents a volume mounted in a container and exposed by a node.
	
	List of edges:
	- CE_MODULE_LOAD: A container can load a kernel module on the node.
	- CE_NSENTER: Container escape via the nsenter built-in linux program that allows executing a binary into another namespace.
	- CE_PRIV_MOUNT: Mount the host disk and gain access to the host via arbitrary filesystem write.
	- CE_SYS_TRACE: Abuse the legitimate OS debugging mechanisms to escape the container via attaching to a node process.
	- CE_UMH_CORE_PATTERN: Abuse the User Mode Helper (UMH) mechanism to execute arbitrary code in the host.
	- CE_VAR_LOG_SYMLINK: Abuse the /var/log symlink to gain access to the host filesystem.
	- CONTAINER_ATTACH: Attach to a running container to execute commands or inspect the container.
	- ENDPOINT_EXPLOIT: Represents a network endpoint exposed by a container that could be exploited by an attacker (via means known or unknown). This can correspond to a Kubernetes service, node service, node port, or container port.
	- EXPLOIT_CONTAINERD_SOCK: Exploit the containerd socket to gain access to the host.
	- EXPLOIT_HOST_READ: Read sensitive files on the host.
	- EXPLOIT_HOST_WRITE: Write sensitive files on the host.
	- IDENTITY_ASSUME: Represents the capacity to act as an Identity via ownership of a service account token, user PKI certificate, etc.
	- IDENTITY_IMPERSONATE: Impersonate an identity.
	- PERMISSION_DISCOVER: Discover permissions granted to an identity.
	- POD_ATTACH: Attach to a running pod to execute commands or inspect the pod.
	- POD_CREATE: Create a pod on a node.
	- POD_EXEC: Execute a command in a pod.
	- POD_PATCH: Patch a pod on a node.
	- ROLE_BIND: Bind a role to an identity.
	- SHARE_PS_NAMESPACE: All containers in a pod share the same process namespace.
	- TOKEN_BRUTEFORCE: Bruteforce a token.
	- TOKEN_LIST: List tokens.
	- TOKEN_STEAL: Steal a token.
	- VOLUME_ACCESS: Access a volume mounted in a container.
	- VOLUME_DISCOVER: Discover volumes mounted in a container.
	`),
			mcp.WithString(
				"cluster",
				mcp.Required(),
				mcp.Description("Kubernetes cluster name"),
			),
			mcp.WithString(
				"runID",
				mcp.Required(),
				mcp.Description("KubeHound import identifier"),
			),
			mcp.WithString(
				"namespace",
				mcp.Required(),
				mcp.Description("Kubernetes namespace to start from"),
			),
			mcp.WithString(
				"image",
				mcp.Required(),
				mcp.Description("Pod image to start from"),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Parse the arguments.
			var args args
			if err := request.BindArguments(&args); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			// Run the query.
			resultChan := make(chan kubehound.AttackPath, 100)
			if err := containers.GetAttackPaths(ctx, args.Cluster, args.RunID, container.AttackPathFilter{
				Namespace: &args.Namespace,
				Image:     &args.Image,
			}, resultChan); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Close the channel.
			close(resultChan)

			// Prepare a response.
			var output bytes.Buffer
			encoder := json.NewEncoder(&output)
			for r := range resultChan {
				if err := encoder.Encode(r); err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
			}

			return mcp.NewToolResultText(output.String()), nil
		},
	}
}
