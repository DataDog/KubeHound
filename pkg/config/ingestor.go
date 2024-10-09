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
	IngestorMaxArchiveSize = "ingestor.max_archive_size"
	IngestorTempDir        = "ingestor.temp_dir"
	IngestorArchiveName    = "ingestor.archive_name"

	IngestorBlobBucketURL = "ingestor.blob.bucket_url"
	IngestorBlobRegion    = "ingestor.blob.region"
)

type IngestorConfig struct {
	API            IngestorAPIConfig `mapstructure:"api"`
	Blob           *BlobConfig       `mapstructure:"blob"`
	TempDir        string            `mapstructure:"temp_dir"`
	ArchiveName    string            `mapstructure:"archive_name"`
	MaxArchiveSize int64             `mapstructure:"max_archive_size"`
}

type IngestorAPIConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Insecure bool   `mapstructure:"insecure" validate:"omitempty,boolean"`
}

type BlobConfig struct {
	BucketUrl string `mapstructure:"bucket_url"` // Bucket to use to push k8s resources (e.g.: s3://<your_bucket>)
	Region    string `mapstructure:"region"`     // Region to use for the bucket (only for s3)
}
