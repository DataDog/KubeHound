package tag

const (
	CollectorTag = "collector"
	ResourceTag  = "resource"
	RunIdTag     = "run_id"
	LabelTag     = "label"
	StorageTag   = "storage"
)

const (
	StorageJanusGraph = "janusgraph"
	StorageMongoDB    = "mongodb"
	StorageMemCache   = "memcache"
)

const (
	ResourcePods                = "pods"
	ResourceRoles               = "roles"
	ResourceRolebindings        = "rolebindings"
	ResourceNodes               = "nodes"
	ResourceEndpoints           = "endpoints"
	ResourceClusterRoles        = "clusterroles"
	ResourceClusterRolebindings = "clusterrolebindings"
)

var (
	BaseTags = []string{}
)

func makeTag(tag string, value string) string {
	return tag + ":" + value
}

func RunId(uuid string) string {
	return makeTag(RunIdTag, uuid)
}

func Collector(collector string) string {
	return makeTag(CollectorTag, collector)
}

func Storage(storage string) string {
	return makeTag(StorageTag, storage)
}

func Resource(res string) string {
	return makeTag(ResourceTag, res)
}
