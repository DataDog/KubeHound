package vertex

import (
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	containerLabel = "Container"
)

var _ Builder = (*Container)(nil)

type Container struct {
}

func (v Container) Label() string {
	return containerLabel
}

func (v Container) BatchSize() int {
	return DefaultBatchSize
}

func (v Container) Traversal() VertexTraversal {
	return func(source *gremlingo.GraphTraversalSource, inserts []TraversalInput) *gremlingo.GraphTraversal {
		g := source.GetGraphTraversal()

		for _, insert := range inserts {
			i := insert.(*graph.Container)
			g = g.AddV(v.Label()).
				Property("storeId", i.StoreId).
				Property("name", i.Name).
				Property("image", i.Image).
				Property("privileged", i.Privileged).
				Property("privesc", i.PrivEsc).
				Property("hostPid", i.HostPID).
				Property("hostPath", i.HostPath).
				Property("hostNetwork", i.HostNetwork).
				Property("runAsUser", i.RunAsUser).
				Property("pod", i.Pod).
				Property("node", i.Node).
				Property("compromised", int(i.Compromised)).
				Property("critical", i.Critical)

			fmt.Printf("@@@@@@@@ Command for container: %+v\n", i.Command)
			for _, cmd := range i.Command {
				g = g.Property(gremlingo.Cardinality.Set, "command", cmd)
			}
			fmt.Printf("@@@@@@@@ Args for container: %+v\n", i.Args)
			for _, arg := range i.Args {
				g = g.Property(gremlingo.Cardinality.Set, "args", arg)
			}
			fmt.Printf("@@@@@@@@ Capabilities for container: %+v\n", i.Capabilities)
			for _, cap := range i.Capabilities {
				g = g.Property(gremlingo.Cardinality.Set, "capabilities", cap)
			}
			fmt.Printf("@@@@@@@@ Ports for container: %+v\n", i.Ports)
			for _, port := range i.Ports {
				g = g.Property(gremlingo.Cardinality.Set, "ports", port)
			}
		}
		return g
	}
}
