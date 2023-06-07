package vertex

import (
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
		// 	i := insert.(graph.Container)
		// 	g.AddV(v.Label()).
		// 		Property("id", i.StoreId).
		// 		Property("name", i.Name)
		// }

		// return g.GetGraphTraversal()
		return g.Inject(inserts).Unfold().As("c").
			AddV("Container").
			Property("storeId", gremlingo.T__.Select("c").Select("store_id")).
			Property("name", gremlingo.T__.Select("c").Select("name")).
			Property("image", gremlingo.T__.Select("c").Select("image")).
			Property("capabilities", gremlingo.T__.Select("c").Select("capabilities")).
			Property("command", gremlingo.T__.Select("c").Select("command")).
			Property("capabilities", gremlingo.T__.Select("c").Select("capabilities")).
			Property("privileged", gremlingo.T__.Select("c").Select("privileged")).
			Property("privesc", gremlingo.T__.Select("c").Select("privesc")).
			Property("hostPid", gremlingo.T__.Select("c").Select("hostPid")).
			Property("hostPath", gremlingo.T__.Select("c").Select("hostPath")).
			Property("hostNetwork", gremlingo.T__.Select("c").Select("hostNetwork")).
			Property("runAsUser", gremlingo.T__.Select("c").Select("runAsUser")).
			Property("ports", gremlingo.T__.Select("c").Select("ports")).
			Property("pod", gremlingo.T__.Select("c").Select("pod")).
			Property("node", gremlingo.T__.Select("c").Select("node")).
			Property("compromised", gremlingo.T__.Select("c").Select("compromised")).
			Property("critical", gremlingo.T__.Select("c").Select("critical"))
	}
}
