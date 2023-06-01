package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	containerLabel = "Container"
)

var _ Vertex = (*Container)(nil)

type Container struct {
}

func (v Container) Label() string {
	return containerLabel
}

func (v Container) Traversal() VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()

		for _, insert := range inserts {
			i := insert.(*graph.Container)
			g = g.AddV(v.Label()).
				Property("id", i.StoreId).
				Property("name", i.Name).
				Property("image", i.Image).
				Property("command", i.Command).
				Property("args", i.Args).
				Property("capabilities", i.Capabilities).
				Property("privileged", i.Privileged).
				Property("privesc", i.PrivEsc).
				Property("host_pid", i.HostPID).
				Property("host_path", i.HostPath).
				Property("host_ipc", i.HostIPC).
				Property("host_network", i.HostNetwork).
				Property("run_as_user", i.RunAsUser).
				Property("ports", i.Ports).
				Property("pod", i.Pod).
				Property("node", i.Node).
				Property("compromised", i.Compromised).
				Property("critical", i.Critical)
		}

		return g
	}
}
