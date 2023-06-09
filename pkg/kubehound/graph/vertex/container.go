package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/utils"
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
	return func(g *gremlingo.GraphTraversalSource, inserts []TraversalInput) *gremlingo.GraphTraversal {
		// g = g.GetGraphTraversal()

		// for _, insert := range inserts {
		// 	i := insert.(map[string]any)
		// 	// command := utils.ToSliceOfAny(i["command"].([]any))
		// 	// args := utils.ToSliceOfAny(i["args"].([]any))
		// 	// capabilities := utils.ToSliceOfAny(i["capabilities"].([]any))

		// 	traversal := g.AddV(v.Label()).
		// 		Property("storeId", i["storeId"]).
		// 		Property("name", i["name"]).
		// 		Property("image", i["image"]).
		// 		Property("privileged", i["privileged"]).
		// 		Property("privesc", i["privesc"]).
		// 		Property("hostPid", i["hostPid"]).
		// 		Property("hostPath", i["hostPath"]).
		// 		Property("hostNetwork", i["hostNetwork"]).
		// 		Property("runAsUser", i["runAsUser"]).
		// 		Property("pod", i["pod"]).
		// 		Property("node", i["node"]).
		// 		Property("compromised", i["compromised"]).
		// 		Property("critical", i["critical"])

		// 	// for _, cmd := range command {
		// 	// 	traversal = traversal.Property(gremlingo.Cardinality.Set, "command", cmd)
		// 	// }
		// 	// for _, arg := range args {
		// 	// 	traversal = traversal.Property(gremlingo.Cardinality.Set, "args", arg)
		// 	// }
		// 	// for _, cap := range capabilities {
		// 	// 	traversal = traversal.Property(gremlingo.Cardinality.Set, "capabilities", cap)
		// 	// }
		// 	// for _, port := range ports {
		// 	// 	traversal = traversal.Property(gremlingo.Cardinality.Set, "ports", port)
		// 	// }
		// }

		// return g.GetGraphTraversal()
		// I have no idea how I can convert the arrays in the inserts to multiple values with cardinality set....
		insertsConverted := utils.ConvertSliceAnyToTyped[graph.Container, TraversalInput](inserts)
		log.I.Infof(" ============== INSERTS Containers ====== %+v", insertsConverted)
		traversal := g.Inject(insertsConverted).Unfold().As("c").
			AddV(v.Label()).
			Property("storeId", gremlingo.T__.Select("c").Select("store_id")).
			Property("name", gremlingo.T__.Select("c").Select("name")).
			Property("image", gremlingo.T__.Select("c").Select("image")).
			// Property(gremlingo.Cardinality.Set, "command", gremlingo.T__.Select("c").Select("command")).
			// Property(gremlingo.Cardinality.Set, "args", gremlingo.T__.Select("c").Select("args")).
			// Property(gremlingo.Cardinality.Set, "capabilities", gremlingo.T__.Select("c").Select("capabilities")).
			Property("privileged", gremlingo.T__.Select("c").Select("privileged")).
			Property("privesc", gremlingo.T__.Select("c").Select("privesc")).
			Property("hostPid", gremlingo.T__.Select("c").Select("hostPid")).
			Property("hostPath", gremlingo.T__.Select("c").Select("hostPath")).
			Property("hostNetwork", gremlingo.T__.Select("c").Select("hostNetwork")).
			Property("runAsUser", gremlingo.T__.Select("c").Select("runAsUser")).
			// Property(gremlingo.Cardinality.Set, "ports", gremlingo.T__.Select("c").Select("ports")).
			Property("pod", gremlingo.T__.Select("c").Select("pod")).
			Property("node", gremlingo.T__.Select("c").Select("node")).
			// Property("compromised", gremlingo.T__.Select("c").Select("compromised")).
			Property("critical", gremlingo.T__.Select("c").Select("critical"))

		return traversal
	}
}
