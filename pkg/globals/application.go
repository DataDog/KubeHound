package globals

import (
	"os"
)

const (
	DDServiceName    = "kubehound"
	DefaultDDEnv     = "dev"
	DefaultComponent = "kubehound-ingestor"
)

func GetDDEnv() string {
	env := os.Getenv("DD_ENV")
	if env == "" {
		return DefaultDDEnv
	}

	return env
}

const (
	FileCollectorComponent = "file-collector"
	IngestorComponent      = "pipeline-ingestor"
	BuilderComponent       = "graph-builder"
)
