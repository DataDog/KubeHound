package globals

const (
	DDServiceName = "kubehound"
	DDEnv         = "prod"
)

func ProdEnv() bool {
	return false
}

const (
	FileCollectorComponent = "file-collector"
	IngestorComponent      = "pipeline-ingestor"
	BuilderComponent       = "graph-builder"
)
