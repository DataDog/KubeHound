package config

import (
	"bytes"
	"fmt"
	"os"

	embedconfig "github.com/DataDog/KubeHound/configs"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/viper"
)

var (
	BuildVersion string // This should be overwritten by the go build -X flags
)

const (
	DefaultConfigType  = "yaml"
	DefaultClusterName = "unknown"
)

// KubehoundConfig defines the top-level application configuration for KubeHound.
type KubehoundConfig struct {
	Debug      bool             `mapstructure:"debug"`      // Debug mode
	Collector  CollectorConfig  `mapstructure:"collector"`  // Collector configuration
	MongoDB    MongoDBConfig    `mapstructure:"mongodb"`    // MongoDB configuration
	JanusGraph JanusGraphConfig `mapstructure:"janusgraph"` // JanusGraph configuration
	Storage    StorageConfig    `mapstructure:"storage"`    // Global param for all storage provider
	Telemetry  TelemetryConfig  `mapstructure:"telemetry"`  // telemetry configuration, contains statsd and other sub structures
	Builder    BuilderConfig    `mapstructure:"builder"`    // Graph builder  configuration
	Ingestor   IngestorConfig   `mapstructure:"ingestor"`   // Ingestor configuration
	Dynamic    DynamicConfig    `mapstructure:"dynamic"`    // Dynamic (i.e runtime generated) configuration
}

// MustLoadEmbedConfig loads the embedded default application configuration, treating all errors as fatal.
func MustLoadEmbedConfig() *KubehoundConfig {
	cfg, err := NewEmbedConfig(embedconfig.DefaultPath)
	if err != nil {
		log.I.Fatalf("embed config load: %v", err)
	}

	return cfg
}

// MustLoadConfig loads the application configuration from the provided path, treating all errors as fatal.
func MustLoadConfig(configPath string) *KubehoundConfig {
	cfg, err := NewConfig(configPath)
	if err != nil {
		log.I.Fatalf("config load: %v", err)
	}

	return cfg
}

// MustLoadConfig loads the application configuration from the provided path, treating all errors as fatal.
func MustLoadInlineConfig() *KubehoundConfig {
	cfg, err := NewInlineConfig()
	if err != nil {
		log.I.Fatalf("config load: %v", err)
	}

	return cfg
}

func NewKubehoundConfig(configPath string, inLine bool) *KubehoundConfig {
	// Configuration initialization
	var cfg *KubehoundConfig
	switch {
	case len(configPath) != 0:
		log.I.Infof("Loading application configuration from file %s", configPath)
		cfg = MustLoadConfig(configPath)
	case inLine:
		cfg = MustLoadInlineConfig()
	default:
		log.I.Infof("Loading application configuration from default embedded")
		cfg = MustLoadEmbedConfig()
	}

	return cfg
}

// SetDefaultValues loads the default value from the different modules
func SetDefaultValues(c *viper.Viper) {
	// K8s Live collector module
	c.SetDefault(CollectorLivePageSize, DefaultK8sAPIPageSize)
	c.SetDefault(CollectorLivePageBufferSize, DefaultK8sAPIPageBufferSize)
	c.SetDefault(CollectorLiveRate, DefaultK8sAPIRateLimitPerSecond)

	// Default values for storage provider
	c.SetDefault("storage.wipe", true)
	c.SetDefault("storage.retry", DefaultRetry)
	c.SetDefault("storage.retry_delay", DefaultRetryDelay)

	// Disable Datadog telemetry by default
	c.SetDefault("telemetry.enabled", false)

	// Default value for MongoDB
	c.SetDefault("mongodb.url", DefaultMongoUrl)
	c.SetDefault("mongodb.connection_timeout", DefaultConnectionTimeout)

	// Defaults values for JanusGraph
	c.SetDefault("janusgraph.url", DefaultJanusGraphUrl)
	c.SetDefault("janusgraph.connection_timeout", DefaultConnectionTimeout)

	// Profiler values
	c.SetDefault("telemetry.profiler.period", DefaultProfilerPeriod)
	c.SetDefault("telemetry.profiler.cpu_duration", DefaultProfilerCPUDuration)

	// Default values for graph builder
	c.SetDefault("builder.vertex.batch_size", DefaultVertexBatchSize)
	c.SetDefault("builder.vertex.batch_size_small", DefaultVertexBatchSizeSmall)
	c.SetDefault("builder.edge.worker_pool_size", DefaultEdgeWorkerPoolSize)
	c.SetDefault("builder.edge.worker_pool_capacity", DefaultEdgeWorkerPoolCapacity)
	c.SetDefault("builder.edge.batch_size", DefaultEdgeBatchSize)
	c.SetDefault("builder.edge.batch_size_small", DefaultEdgeBatchSizeSmall)
	c.SetDefault("builder.edge.batch_size_cluster_impact", DefaultEdgeBatchSizeClusterImpact)
	c.SetDefault("builder.stop_on_error", DefaultStopOnError)

	c.SetDefault("ingestor.api.port", DefaultIngestorAPIPort)
	c.SetDefault("ingestor.api.address", DefaultIngestorAPIAddr)
	c.SetDefault("ingestor.bucket_name", DefaultBucketName)
	c.SetDefault("ingestor.temp_dir", DefaultTempDir)
	c.SetDefault("ingestor.max_archive_size", DefaultMaxArchiveSize)
	c.SetDefault("ingestor.archive_name", DefaultArchiveName)
}

// SetEnvOverrides enables environment variable overrides for the config.
func SetEnvOverrides(c *viper.Viper) {
	var res *multierror.Error

	// Enable changing file collector fields via environment variables
	res = multierror.Append(res, c.BindEnv("collector.type", "KH_COLLECTOR"))
	res = multierror.Append(res, c.BindEnv("collector.file.directory", "KH_COLLECTOR_DIR"))
	res = multierror.Append(res, c.BindEnv("collector.file.cluster", "KH_COLLECTOR_TARGET"))

	if res.ErrorOrNil() != nil {
		log.I.Fatalf("config environment override: %v", res.ErrorOrNil())
	}
}

// NewConfig creates a new config instance from the provided file using viper.
func NewConfig(configPath string) (*KubehoundConfig, error) {
	viper.SetConfigType(DefaultConfigType)
	viper.SetConfigFile(configPath)

	// Configure default values
	SetDefaultValues(viper.GetViper())

	// Configure environment variable override
	SetEnvOverrides(viper.GetViper())
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	kc := KubehoundConfig{}
	if err := viper.Unmarshal(&kc); err != nil {
		return nil, fmt.Errorf("unmarshaling config data: %w", err)
	}

	return &kc, nil
}

// NewConfig creates a new config instance from the provided file using viper.
func NewInlineConfig() (*KubehoundConfig, error) {
	// Configure environment variable override
	SetEnvOverrides(viper.GetViper())
	kc := KubehoundConfig{}
	if err := viper.Unmarshal(&kc); err != nil {
		return nil, fmt.Errorf("unmarshaling config data: %w", err)
	}

	return &kc, nil
}

// NewEmbedConfig creates a new config instance from an embedded config file using viper.
func NewEmbedConfig(configPath string) (*KubehoundConfig, error) {
	viper.SetConfigType(DefaultConfigType)
	SetDefaultValues(viper.GetViper())

	data, err := embedconfig.F.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading embed config: %w", err)
	}

	err = viper.ReadConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	kc := KubehoundConfig{}
	if err := viper.Unmarshal(&kc); err != nil {
		return nil, fmt.Errorf("unmarshaling config data: %w", err)
	}

	return &kc, nil
}

// IsCI determines whether the application is running within a CI action
func IsCI() bool {
	return os.Getenv("CI") != ""
}

// ComputeDynamic sets the dynamic components of the config from the provided options.
func (kc *KubehoundConfig) ComputeDynamic(opts ...DynamicOption) error {
	kc.Dynamic.mu.Lock()
	defer kc.Dynamic.mu.Unlock()

	kc.Dynamic.RunID = NewRunID()
	kc.Dynamic.Cluster = DefaultClusterName

	for _, option := range opts {
		opt, err := option()
		if err != nil {
			return err
		}
		opt(&kc.Dynamic)
	}

	return nil
}
