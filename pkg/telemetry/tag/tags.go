package tag

const (
	CollectorTag = "collector"
	// ResourceTag   = "resource"
	EntityTag     = "entity"
	RunIdTag      = "run_id"
	LabelTag      = "label"
	CollectionTag = "collection"
	BuilderTag    = "builder"
	StorageTag    = "storage"
	CacheKeyTag   = "cache_key"
	EdgeTypeTag   = "edge_type"
)

const (
	StorageJanusGraph = "janusgraph"
	StorageMongoDB    = "mongodb"
	StorageMemCache   = "memcache"
)

const (
	EntityPods                = "pods"
	EntityRoles               = "roles"
	EntityRolebindings        = "rolebindings"
	EntityNodes               = "nodes"
	EntityEndpoints           = "endpoints"
	EntityClusterRoles        = "clusterroles"
	EntityClusterRolebindings = "clusterrolebindings"
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

// func Resource(res string) string {
// 	return makeTag(ResourceTag, res)
// }

func Entity(name string) string {
	return makeTag(EntityTag, name)
}

func Label(label string) string {
	return makeTag(LabelTag, label)
}

func Builder(builder string) string {
	return makeTag(BuilderTag, builder)
}

func Collection(collection string) string {
	return makeTag(CollectionTag, collection)
}

func CacheKey(ck string) string {
	return makeTag(CacheKeyTag, ck)
}

func EdgeType(et string) string {
	return makeTag(EdgeTypeTag, et)
}
