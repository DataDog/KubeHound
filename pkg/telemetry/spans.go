package telemetry

const (
	SpanJanusGraphOperationFlush      = "kubehound.janusgraph.flush"
	SpanJanusGraphOperationBatchWrite = "kubehound.janusgraph.batchwrite"

	SpanMongodbOperationFlush      = "kubehound.mongo.flush"
	SpanMongodbOperationBatchWrite = "kubehound.mongo.batchwrite"

	SpanOperationIngestData = "kubehound.ingestData"
	SpanOperationBuildGraph = "kubehound.buildGraph"
	SpanOperationLaunch     = "kubehound.launch"

	SpanOperationStream   = "kubehound.collector.stream"
	SpanOperationReadFile = "kubehound.collector.readFile"

	SpanOperationRun = "kubehound.graph.builder.run"
)
