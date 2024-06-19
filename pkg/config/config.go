package config

import (
	"bytes"
	"fmt"
	"os"

	embedconfig "github.com/DataDog/KubeHound/configs"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/viper"
)

var (
	BuildVersion string // This should be overwritten by the go build -X flags
	BuildArch    string // This should be overwritten by the go build -X flags
	BuildOs      string // This should be overwritten by the go build -X flags
)

const (
	DefaultConfigType  = "yaml"
	DefaultClusterName = "unknown"

	GlobalDebug = "debug"
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
	cfg, err := NewEmbedConfig(viper.GetViper(), embedconfig.DefaultPath)
	if err != nil {
		log.I.Fatalf("embed config load: %v", err)
	}

	return cfg
}

// MustLoadConfig loads the application configuration from the provided path, treating all errors as fatal.
func MustLoadConfig(configPath string) *KubehoundConfig {
	cfg, err := NewConfig(viper.GetViper(), configPath)
	if err != nil {
		log.I.Fatalf("config load: %v", err)
	}

	return cfg
}

// MustLoadConfig loads the application configuration from the provided path, treating all errors as fatal.
func MustLoadInlineConfig() *KubehoundConfig {
	cfg, err := NewInlineConfig(viper.GetViper())
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
		log.I.Info("Loading application from inline command")
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
	c.SetDefault(CollectorNonInteractive, DefaultK8sAPINonInteractive)

	// Default values for storage provider
	c.SetDefault("storage.wipe", true)
	c.SetDefault("storage.retry", DefaultRetry)
	c.SetDefault("storage.retry_delay", DefaultRetryDelay)

	// Disable Datadog telemetry by default
	c.SetDefault(TelemetryEnabled, false)

	// Default value for MongoDB
	c.SetDefault("mongodb.url", DefaultMongoUrl)
	c.SetDefault("mongodb.connection_timeout", DefaultConnectionTimeout)

	// Defaults values for JanusGraph
	c.SetDefault("janusgraph.url", DefaultJanusGraphUrl)
	c.SetDefault("janusgraph.connection_timeout", DefaultConnectionTimeout)

	// Profiler values
	c.SetDefault(TelemetryProfilerPeriod, DefaultProfilerPeriod)
	c.SetDefault(TelemetryProfilerCPUDuration, DefaultProfilerCPUDuration)

	// Default values for graph builder
	c.SetDefault("builder.vertex.batch_size", DefaultVertexBatchSize)
	c.SetDefault("builder.vertex.batch_size_small", DefaultVertexBatchSizeSmall)
	c.SetDefault("builder.edge.worker_pool_size", DefaultEdgeWorkerPoolSize)
	c.SetDefault("builder.edge.worker_pool_capacity", DefaultEdgeWorkerPoolCapacity)
	c.SetDefault("builder.edge.batch_size", DefaultEdgeBatchSize)
	c.SetDefault("builder.edge.batch_size_small", DefaultEdgeBatchSizeSmall)
	c.SetDefault("builder.edge.batch_size_cluster_impact", DefaultEdgeBatchSizeClusterImpact)
	c.SetDefault("builder.stop_on_error", DefaultStopOnError)
	c.SetDefault("builder.edge.large_cluster_optimizations", DefaultLargeClusterOptimizations)

	c.SetDefault(IngestorAPIEndpoint, DefaultIngestorAPIEndpoint)
	c.SetDefault(IngestorAPIInsecure, DefaultIngestorAPIInsecure)
	c.SetDefault(IngestorBlobBucketName, DefaultBucketName)
	c.SetDefault(IngestorTempDir, DefaultTempDir)
	c.SetDefault(IngestorMaxArchiveSize, DefaultMaxArchiveSize)
	c.SetDefault(IngestorArchiveName, DefaultArchiveName)
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

func unmarshalConfig(v *viper.Viper) (*KubehoundConfig, error) {
	kc := KubehoundConfig{}
	kc.Dynamic.mu.Lock()
	defer kc.Dynamic.mu.Unlock()

	err := v.Unmarshal(&kc)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling config data: %w", err)
	}

	// Validating the object
	validateConfig := validator.New(validator.WithRequiredStructEnabled())
	err = validateConfig.Struct(&kc)
	if err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &kc, nil
}

// NewConfig creates a new config instance from the provided file using viper.
func NewConfig(v *viper.Viper, configPath string) (*KubehoundConfig, error) {
	v.SetConfigType(DefaultConfigType)
	v.SetConfigFile(configPath)

	// Configure default values
	SetDefaultValues(v)

	// Configure environment variable override
	SetEnvOverrides(v)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	kc, err := unmarshalConfig(v)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling config data: %w", err)
	}

	return kc, nil
}

// NewConfig creates a new config instance from the provided file using viper.
func NewInlineConfig(v *viper.Viper) (*KubehoundConfig, error) {
	// Load default embedded config file
	SetDefaultValues(v)

	// Configure environment variable override
	SetEnvOverrides(v)

	kc, err := unmarshalConfig(v)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling config data: %w", err)
	}

	return kc, nil
}

// NewEmbedConfig creates a new config instance from an embedded config file using viper.
func NewEmbedConfig(v *viper.Viper, configPath string) (*KubehoundConfig, error) {
	v.SetConfigType(DefaultConfigType)
	SetDefaultValues(v)

	data, err := embedconfig.F.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading embed config: %w", err)
	}

	err = viper.ReadConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	kc, err := unmarshalConfig(v)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling config data: %w", err)
	}

	return kc, nil
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
	kc.Dynamic.ClusterName = DefaultClusterName

	for _, option := range opts {
		opt, err := option()
		if err != nil {
			return err
		}
		opt(&kc.Dynamic)
	}

	return nil
}
