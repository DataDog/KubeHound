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

// Collector/dumper component spans
const (
	CollectorStream = "kubehound.collector.stream"
	CollectorDump   = "kubehound.collector.dump"

	DumperLaunch = "kubehound.dumper.launch"

	DumperNodes               = "kubehound.dumper.nodes"
	DumperPods                = "kubehound.dumper.pods"
	DumperEndpoints           = "kubehound.dumper.endpoints"
	DumperRoles               = "kubehound.dumper.roles"
	DumperClusterRoles        = "kubehound.dumper.clusterroles"
	DumperRoleBindings        = "kubehound.dumper.rolebindings"
	DumperClusterRoleBindings = "kubehound.dumper.clusterrolebindings"

	DumperReadFile    = "kubehound.dumper.readFile"
	DumperS3Push      = "kubehound.dumper.s3_push"
	DumperS3Download  = "kubehound.dumper.s3_download"
	DumperWriterWrite = "kubehound.dumper.write"
	DumperWriterFlush = "kubehound.dumper.flush"
	DumperWriterClose = "kubehound.dumper.close"
)

// Graph builder spans
const (
	BuildEdge = "kubehound.graph.builder.edge"
)
