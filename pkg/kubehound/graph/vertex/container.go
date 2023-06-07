package vertex

import (
	"fmt"

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

		for _, insert := range inserts {
			i := insert.(map[string]any)
			command, ok := i["command"].([]string)
			if !ok {
				command = []string{"failed_to_parse_command"}
			}

			args, ok := i["args"].([]string)
			if !ok {
				args = []string{"failed_to_parse_args"}
			}

			capabilities, ok := i["capabilities"].([]string)
			if !ok {
				capabilities = []string{"failed_to_parse_capabilities"}
			}
			ports, ok := i["ports"].([]int)
			if !ok {
				ports = []int{0}
			}

			promise := g.AddV(v.Label()).
				Property("storeId", i["storeId"]).
				Property("name", i["name"]).
				Property("image", i["image"]).
				Property([]interface{}{gremlingo.Cardinality.Set, "command", command}...).
				Property([]interface{}{gremlingo.Cardinality.Set, "args", args}...).
				Property([]interface{}{gremlingo.Cardinality.Set, "capabilities", capabilities}...).
				Property("privileged", i["privileged"]).
				Property("privesc", i["privesc"]).
				Property("hostPid", i["hostPid"]).
				Property("hostPath", i["hostPath"]).
				Property("hostNetwork", i["hostNetwork"]).
				Property("runAsUser", i["runAsUser"]).
				Property([]interface{}{gremlingo.Cardinality.Set, "ports", ports}...).
				Property("pod", i["pod"]).
				Property("node", i["node"]).
				Property("compromised", i["compromised"]).
				Property("critical", i["critical"]).
				Iterate()
			err := <-promise
			if err != nil {
				fmt.Printf("promise err: %v\n", err)
			}
		}
		return g.GetGraphTraversal()

		// I have no idea how I can convert the arrays in the inserts to multiple values with cardinality set....
		// promise := g.Inject(inserts).Unfold().As("c").
		// 	AddV(v.Label()).
		// 	Property("storeId", gremlingo.T__.Select("c").Select("store_id")).
		// 	Property("name", gremlingo.T__.Select("c").Select("name")).
		// 	Property("image", gremlingo.T__.Select("c").Select("image")).
		// 	Property(gremlingo.Cardinality.Set, "command", gremlingo.T__.Select("c").Select("command")).
		// 	Property(gremlingo.Cardinality.Set, "args", gremlingo.T__.Select("c").Select("args")).
		// 	Property(gremlingo.Cardinality.Set, "capabilities", gremlingo.T__.Select("c").Select("capabilities")).
		// 	Property("privileged", gremlingo.T__.Select("c").Select("privileged")).
		// 	Property("privesc", gremlingo.T__.Select("c").Select("privesc")).
		// 	Property("hostPid", gremlingo.T__.Select("c").Select("hostPid")).
		// 	Property("hostPath", gremlingo.T__.Select("c").Select("hostPath")).
		// 	Property("hostNetwork", gremlingo.T__.Select("c").Select("hostNetwork")).
		// 	Property("runAsUser", gremlingo.T__.Select("c").Select("runAsUser")).
		// 	Property(gremlingo.Cardinality.Set, "ports", gremlingo.T__.Select("c").Select("ports")).
		// 	Property("pod", gremlingo.T__.Select("c").Select("pod")).
		// 	Property("node", gremlingo.T__.Select("c").Select("node")).
		// 	Property("compromised", gremlingo.T__.Select("c").Select("compromised")).
		// 	Property("critical", gremlingo.T__.Select("c").Select("critical")).
		// 	Iterate()

		// 	// The returned promised is a go channel to wait for all submitted steps to finish execution and return error.
		// err := <-promise
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// return g.GetGraphTraversal()
	}
}
