package config

const (
	CollectorTypeFile = "file-collector"
)

type CollectorConfig struct {
	Type string
	File *FileCollectorConfig
}

type FileCollectorConfig struct {
	Directory string
}
