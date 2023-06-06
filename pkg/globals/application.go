package globals

const (
	DDServiceName = "kubehound"
	DDTeamName    = "ase"
	DDEnv         = "prod"
)

func ProdEnv() bool {
	// TODO
	return false
}

const (
	FileCollectorComponent = "file-collector"
	IngestorComponent      = "pipeline-ingestor"
	BuilderComponent       = "graph-builder"
)
