package collector

var (
	MetricCollectorNodesCount               = "kubehound.collector.nodes.count"
	MetricCollectorPodsCount                = "kubehound.collector.pods.count"
	MetricCollectorRolesCount               = "kubehound.collector.roles.count"
	MetricCollectorRoleBindingsCount        = "kubehound.collector.rolebindings.count"
	MetricCollectorClusterRolesCount        = "kubehound.collector.clusterroles.count"
	MetricCollectorClusterRoleBindingsCount = "kubehound.collector.clusterrolebindings.count"
)

var (
	SpanOperationStream   = "kubehound.collector.stream"
	SpanOperationReadFile = "kubehound.collector.readFile"
)
