package config

const (
	CollectorTypeFile   = "file-collector"
	CollectorTypeK8sAPI = "live-k8s-api-collector"
)

const (
	DefaultK8sAPIPageSize           int64 = 500
	DefaultK8sAPIPageBufferSize     int32 = 10
	DefaultK8sAPIRateLimitPerSecond int   = 100
	DefaultK8sAPINonInteractive     bool  = false

	CollectorLiveRate           = "collector.live.rate_limit_per_second"
	CollectorLivePageSize       = "collector.live.page_size"
	CollectorLivePageBufferSize = "collector.live.page_buffer_size"
	CollectorNonInteractive     = "collector.non_interactive"
	CollectorFileArchiveFormat  = "collector.file.archive.format"
	CollectorFileDirectory      = "collector.file.directory"
	CollectorFileClusterName    = "collector.file.cluster_name"
	CollectorFileBlobRegion     = "collector.file.blob.region"
	CollectorFileBlobBucket     = "collector.file.blob.bucket"
)

// CollectorConfig configures collector specific parameters.
type CollectorConfig struct {
	Type           string                 `mapstructure:"type"`            // Collector type
	File           *FileCollectorConfig   `mapstructure:"file"`            // File collector specific configuration
	Live           *K8SAPICollectorConfig `mapstructure:"live"`            // File collector specific configuration
	NonInteractive bool                   `mapstructure:"non_interactive"` // Skip confirmation
}

// K8SAPICollectorConfig configures the K8sAPI collector.
type K8SAPICollectorConfig struct {
	PageSize           int64 `mapstructure:"page_size"`             // Number of entry being retrieving by each call on the API (same for all Kubernetes entry types)
	PageBufferSize     int32 `mapstructure:"page_buffer_size"`      // Number of pages to buffer
	RateLimitPerSecond int   `mapstructure:"rate_limit_per_second"` // Rate limiting per second across all calls (same for all kubernetes entry types) against the Kubernetes API
}

// FileCollectorConfig configures the file collector.
type FileCollectorConfig struct {
	ClusterName string             `mapstructure:"cluster_name"` // Target cluster (must be specified in config as not present in JSON files)
	Directory   string             `mapstructure:"directory"`    // Base directory holding the K8s data JSON files
	Archive     *FileArchiveConfig `mapstructure:"archive"`      // Archive configuration
	Blob        *BlobConfig        `mapstructure:"blob"`         // Blob storage configuration
}

type FileArchiveConfig struct {
	ArchiveName string `mapstructure:"archive_name"` // Name of the output archive
	Format      bool   `mapstructure:"format"`       // Enable compression for the dumped data (generates a tar.gz file)
}

type BlobConfig struct {
	Bucket string `mapstructure:"bucket"` // Bucket to use to push k8s resources (e.g.: s3://<your_bucket>)
	Region string `mapstructure:"region"` // Region to use for the bucket (only for s3)
}
