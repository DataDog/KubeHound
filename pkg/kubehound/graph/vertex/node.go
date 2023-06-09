package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/utils"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	nodeLabel = "Node"
)

var _ Builder = (*Node)(nil)

type Node struct {
}

func (v Node) Label() string {
	return nodeLabel
}

func (v Node) BatchSize() int {
	return DefaultBatchSize
}

func (v Node) Traversal() VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		insertsConverted := utils.ConvertSliceAnyToTyped[graph.Node, TraversalInput](inserts)
		toStore := utils.ConvertToSliceMapAny(insertsConverted)
		log.I.Infof(" ============== INSERTS Nodes ====== %+v", insertsConverted)
		log.I.Infof(" ============== toStore Nodes ====== %+v", toStore)
		g := source.GetGraphTraversal()
		for _, i := range toStore {
			// i := insert.(map[string]any)
			// command := utils.ToSliceOfAny(i["command"].([]any))
			// args := utils.ToSliceOfAny(i["args"].([]any))
			// capabilities := utils.ToSliceOfAny(i["capabilities"].([]any))

			g = g.AddV(v.Label()).
				Property("storeId", i["storeId"]).
				Property("name", i["name"]).
				Property("image", i["image"]).
				Property("privileged", i["privileged"]).
				Property("privesc", i["privesc"]).
				Property("hostPid", i["hostPid"]).
				Property("hostPath", i["hostPath"]).
				Property("hostNetwork", i["hostNetwork"]).
				Property("runAsUser", i["runAsUser"]).
				Property("pod", i["pod"]).
				Property("node", i["node"]).
				// Property("compromised", i["compromised"]).
				Property("critical", i["critical"])

			// for _, cmd := range command {
			// 	traversal = traversal.Property(gremlingo.Cardinality.Set, "command", cmd)
			// }
			// for _, arg := range args {
			// 	traversal = traversal.Property(gremlingo.Cardinality.Set, "args", arg)
			// }
			// for _, cap := range capabilities {
			// 	traversal = traversal.Property(gremlingo.Cardinality.Set, "capabilities", cap)
			// }
			// for _, port := range ports {
			// 	traversal = traversal.Property(gremlingo.Cardinality.Set, "ports", port)
			// }
		}
		return g
		// return g.Inject(toStore).Unfold().As("c").
		// 	AddV(v.Label()).
		// 	Property("store_id", gremlingo.T__.Select("c").Select("store_id")).
		// 	Property("name", gremlingo.T__.Select("c").Select("name")).
		// 	Property("is_namespaced", gremlingo.T__.Select("c").Select("is_namespaced")).
		// 	Property("namespace", gremlingo.T__.Select("c").Select("namespace")).
		// 	Property("compromised", gremlingo.T__.Select("c").Select("compromised")).
		// 	Property("critical", gremlingo.T__.Select("c").Select("critical"))

	}
}
