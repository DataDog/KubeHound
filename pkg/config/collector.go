package config

const (
	CollectorTypeFile   = "file-collector"
	CollectorTypeK8sAPI = "live-k8s-api-collector"
)

// CollectorConfig configures collector specific parameters.
type CollectorConfig struct {
	Type string                 `mapstructure:"type"` // Collector type
	File *FileCollectorConfig   `mapstructure:"file"` // File collector specific configuration
	Live *K8SAPICollectorConfig `mapstructure:"live"` // File collector specific configuration
}

// K8SAPICollectorConfig configures the K8sAPI collector.
type K8SAPICollectorConfig struct {
	PageSize           int64 `mapstructure:"page_size"`             // Number of entry being retrieving by each call on the API (same for all Kubernetes entry types)
	PageBufferSize     int32 `mapstructure:"page_buffer_size"`      // Number of pages to buffer
	RateLimitPerSecond int   `mapstructure:"rate_limit_per_second"` // Rate limiting per second accross all calls (same for all kubernetes entry types) against the Kubernetes API
}

// FileCollectorConfig configures the file collector.
type FileCollectorConfig struct {
	Directory string `mapstructure:"directory"` // Base directory holding the K8s data JSON files
}
