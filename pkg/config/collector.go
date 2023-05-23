package config

const (
	CollectorTypeFile = "file-collector"
)

// CollectorConfig configures collector specific parameters.
type CollectorConfig struct {
	Type string               `mapstructure:"type"` // Collector type
	File *FileCollectorConfig `mapstructure:"file"` // File collector specific configuration
}

// FileCollectorConfig configures the file collector.
type FileCollectorConfig struct {
	Directory string `mapstructure:"directory"` // Base directory holding the K8s data JSON files
}
