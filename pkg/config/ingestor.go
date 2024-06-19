package config

const (
	DefaultIngestorAPIEndpoint = "127.0.0.1:9000"
	DefaultIngestorAPIInsecure = false
	DefaultBucketName          = "" // we want to let it empty because we can easily abort if it's not configured
	DefaultTempDir             = "/tmp/kubehound"
	DefaultArchiveName         = "archive.tar.gz"
	DefaultMaxArchiveSize      = int64(2 << 30) // 2GB

	IngestorAPIEndpoint    = "ingestor.api.endpoint"
	IngestorAPIInsecure    = "ingestor.api.insecure"
	IngestorClusterName    = "ingestor.cluster_name"
	IngestorRunID          = "ingestor.run_id"
	IngestorMaxArchiveSize = "ingestor.max_archive_size"
	IngestorTempDir        = "ingestor.temp_dir"
	IngestorArchiveName    = "ingestor.archive_name"

	IngestorBlobBucketName = "ingestor.blob.bucket_name"
	IngestorBlobRegion     = "ingestor.blob.region"
)

type IngestorConfig struct {
	API            IngestorAPIConfig `mapstructure:"api"`
	Blob           *BlobConfig       `mapstructure:"blob"`
	TempDir        string            `mapstructure:"temp_dir"`
	ArchiveName    string            `mapstructure:"archive_name"`
	MaxArchiveSize int64             `mapstructure:"max_archive_size"`
	ClusterName    string            `mapstructure:"cluster_name"`
	RunID          string            `mapstructure:"run_id"`
}

type IngestorAPIConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Insecure bool   `mapstructure:"insecure" validate:"omitempty,boolean"`
}
