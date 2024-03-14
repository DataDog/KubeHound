package config

const (
	DefaultIngestorAPIPort = 9000
	DefaultIngestorAPIAddr = "127.0.0.1"
	DefaultBucketName      = "" // we want to let it empty because we can easily abort if it's not configured
	DefaultTempDir         = "/tmp/kubehound"
	DefaultArchiveName     = "archive.tar.gz"
	DefaultMaxArchiveSize  = int64(1 << 30) // 1GB
)

type IngestorConfig struct {
	API            IngestorAPIConfig `mapstructure:"api"`
	BucketName     string            `mapstructure:"bucket_name"`
	TempDir        string            `mapstructure:"temp_dir"`
	ArchiveName    string            `mapstructure:"archive_name"`
	MaxArchiveSize int64             `mapstructure:"max_archive_size"`
}

type IngestorAPIConfig struct {
	Addr string `mapstructure:"address"`
	Port int    `mapstructure:"port"`
}
