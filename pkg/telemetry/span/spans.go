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
	CollectorStream          = "kubehound.collector.stream"
	CollectorDump            = "kubehound.collector.dump"
	CollectorReadFile        = "kubehound.collector.readFile"
	CollectorS3Push          = "kubehound.collector.s3_push"
	CollectorS3Download      = "kubehound.collector.s3_download"
	CollectorFileWriterWrite = "kubehound.collector.file_writer.write"
	CollectorFileWriterFlush = "kubehound.collector.file_writer.flush"
	CollectorFileWriterClose = "kubehound.collector.file_writer.close"
	CollectorTarWriterWrite  = "kubehound.collector.tar_writer.write"
	CollectorTarWriterFlush  = "kubehound.collector.tar_writer.flush"
	CollectorTarWriterClose  = "kubehound.collector.tar_writer.close"
)

// Graph builder spans
const (
	BuildEdge = "kubehound.graph.builder.edge"
)
