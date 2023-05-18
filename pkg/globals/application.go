package globals

const (
	DDServiceName = "kubehound"
	DDTeamName    = "ase"
)

func ProdEnv() bool {
	// TODO
	return false
}

const (
	IngestorComponent = "pipeline-ingestor"
	BuilderComponent  = "graph-builder"
)
