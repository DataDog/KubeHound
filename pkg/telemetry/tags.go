package telemetry

const (
	TagTypeJanusGraph      = "type:janusgraph"
	TagTypeMongodb         = "type:mongodb"
	TagCollectorTypeK8sApi = "collector:k8s-api"
	TagCollectorTypeFile   = "collector:file"

	TagKeyResource = "resource"
	TagKeyLabel    = "label"
	TagKeyRunId    = "run_id"

	TagResourcePods                = "pods"
	TagResourceRoles               = "roles"
	TagResourceRolebindings        = "rolebindings"
	TagResourceNodes               = "nodes"
	TagResourceEndpoints           = "endpoints"
	TagResourceClusterRoles        = "clusterroles"
	TagResourceClusterRolebindings = "clusterrolebindings"
	// BaseTags represents the minimal tags sent by the application
	// Each sub-component of the app will add to their local usage their own tags depending on their needs.
)

var (
	BaseTags = []string{}
)

func MakeTag(tag string, value string) string {
	return tag + ":" + value
}

func UserIdTag(userId string) string {
	return MakeTag(definitions.TagUserId, userId)
}
