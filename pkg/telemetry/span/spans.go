package span

// Top level spans
const (
	IngestData = "kubehound.ingestData"
	BuildGraph = "kubehound.buildGraph"
	Launch     = "kubehound.launch"
)

// JanusGraph provider spans
const (
	JanusGraphFlush      = "kubehound.janusgraph.flush"
	JanusGraphBatchWrite = "kubehound.janusgraph.batchwrite"
)

// MongoDB provider spans
const (
	MongoDBFlush      = "kubehound.mongo.flush"
	MongoDBBatchWrite = "kubehound.mongo.batchwrite"
)

// Collector component spans
const (
	CollectorStream   = "kubehound.collector.stream"
	CollectorReadFile = "kubehound.collector.readFile"
)

// Graph builder spans
const (
	BuildEdgeMutating  = "kubehound.graph.builder.mutating"
	BuildEdgeSimple    = "kubehound.graph.builder.simple"
	BuildEdgeDependent = "kubehound.graph.builder.dependent"
)
