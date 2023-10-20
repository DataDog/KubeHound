package metric

import (
	"fmt"
)

const (
	Prefix = "kubehound"
)

func metricFQN(name string) string {
	return fmt.Sprintf("%s.%s", Prefix, name)
}

// Collector metrics
var (
	CollectorCount = metricFQN("collector.count")
)

// Pipeline storage metrics
var (
	ObjectWrite          = metricFQN("storage.batchwrite.object")
	VertexWrite          = metricFQN("storage.batchwrite.vertex")
	EdgeWrite            = metricFQN("storage.batchwrite.edge")
	QueueSize            = metricFQN("storage.queue.size")
	BackgroundWriterCall = metricFQN("storage.writer.background")
	FlushWriterCall      = metricFQN("storage.writer.flush")
)

// Cache metrics
var (
	CacheMiss      = metricFQN("cache.miss")
	CacheHit       = metricFQN("cache.hit")
	CacheWrite     = metricFQN("cache.write")
	CacheDuplicate = metricFQN("cache.duplicate")
)
