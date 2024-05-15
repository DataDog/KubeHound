package tag

import "sync"

const (
	ActionTypeTag         = "action"
	CollectorTag          = "collector"
	CollectorClusterTag   = "cluster"
	DumperS3BucketTag     = "s3_bucket"
	DumperS3keyTag        = "s3_key"
	DumperFilePathTag     = "file_path"
	DumperWorkerNumberTag = "worker_number"
	DumperWriterTypeTag   = "writer_type"
	EntityTag             = "entity"
	WaitTag               = "wait"
	RunIdTag              = "run_id"
	IngestionRunIdTag     = "ingestion_run_id"
	LabelTag              = "label"
	CollectionTag         = "collection"
	BuilderTag            = "builder"
	StorageTag            = "storage"
	CacheKeyTag           = "cache_key"
	EdgeTypeTag           = "edge_type"
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

type BasesTags struct {
	mu   sync.Mutex
	tags []string
}

var currentBaseTag = BasesTags{}

func SetupBaseTags() {
	currentBaseTag = BasesTags{
		mu:   sync.Mutex{},
		tags: []string{},
	}
}

func AppendBaseTags(tags ...string) {
	currentBaseTag.mu.Lock()
	defer currentBaseTag.mu.Unlock()
	currentBaseTag.tags = append(currentBaseTag.tags, tags...)
}

func GetBaseTagsWith(optTags ...string) []string {
	currentBaseTag.mu.Lock()
	defer currentBaseTag.mu.Unlock()
	var tags []string
	copy(currentBaseTag.tags, tags)

	return append(tags, optTags...)
}

func GetBaseTags() []string {
	currentBaseTag.mu.Lock()
	defer currentBaseTag.mu.Unlock()

	return currentBaseTag.tags
}

func MakeTag(tag string, value string) string {
	return tag + ":" + value
}

func RunID(uuid string) string {
	return MakeTag(RunIdTag, uuid)
}

func IngestionRunID(uuid string) string {
	return MakeTag(IngestionRunIdTag, uuid)
}

func Collector(collector string) string {
	return MakeTag(CollectorTag, collector)
}

func Storage(storage string) string {
	return MakeTag(StorageTag, storage)
}

func Entity(name string) string {
	return MakeTag(EntityTag, name)
}

func Label(label string) string {
	return MakeTag(LabelTag, label)
}

func Builder(builder string) string {
	return MakeTag(BuilderTag, builder)
}

func Collection(collection string) string {
	return MakeTag(CollectionTag, collection)
}

func CacheKey(ck string) string {
	return MakeTag(CacheKeyTag, ck)
}

func EdgeType(et string) string {
	return MakeTag(EdgeTypeTag, et)
}

func ClusterName(cluster string) string {
	return MakeTag(CollectorClusterTag, cluster)
}

func ActionType(action string) string {
	return MakeTag(ActionTypeTag, action)
}

func S3Bucket(bucket string) string {
	return MakeTag(DumperS3BucketTag, bucket)
}

func S3Key(key string) string {
	return MakeTag(DumperS3keyTag, key)
}
