package metric

// Collector metrics
var (
	CollectorCount = "kubehound.collector.count"
	CollectorWait  = "kubehound.collector.wait"
	CollectorSkip  = "kubehound.collector.skip"
	CollectorSize  = "kubehound.s3.size"

	CollectorRunDuration   = "kubehound.collector.run.duration"
	CollectorRunWait       = "kubehound.collector.run.wait"
	CollectorRunThrottling = "kubehound.collector.run.throttling"
)

// Dumper metrics
var (
	DumperCount = "kubehound.collector.count"
	DumperrWait = "kubehound.collector.wait"
	DumperSkip  = "kubehound.collector.skip"
	DumperSize  = "kubehound.s3.size"
)

// Pipeline storage metrics
var (
	ObjectWrite          = "kubehound.storage.batchwrite.object"
	VertexWrite          = "kubehound.storage.batchwrite.vertex"
	EdgeWrite            = "kubehound.storage.batchwrite.edge"
	QueueSize            = "kubehound.storage.queue.size"
	BackgroundWriterCall = "kubehound.storage.writer.background"
	FlushWriterCall      = "kubehound.storage.writer.flush"
	RetryWriterCall      = "kubehound.storage.writer.retry"
)

// Cache metrics
var (
	CacheMiss      = "kubehound.cache.miss"
	CacheHit       = "kubehound.cache.hit"
	CacheWrite     = "kubehound.cache.write"
	CacheDuplicate = "kubehound.cache.duplicate"
)

// Ingestion metrics
const (
	IngestionRunDuration    = "kubehound.ingestion.run.duration"
	IngestionBuildDuration  = "kubehound.ingestion.build.duration"
	IngestionIngestDuration = "kubehound.ingestion.ingest.duration"
)
