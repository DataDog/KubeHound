package config

import (
	"bytes"
	"context"
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
	BuildBranch  string // This should be overwritten by the go build -X flags
	BuildArch    string // This should be overwritten by the go build -X flags
	BuildOs      string // This should be overwritten by the go build -X flags
)

const (
	DefaultConfigType  = "yaml"
	DefaultClusterName = "unknown"
	DefaultConfigName  = "kubehound"

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
func MustLoadEmbedConfig(ctx context.Context) *KubehoundConfig {
	l := log.Logger(ctx)
	cfg, err := NewEmbedConfig(ctx, viper.GetViper(), embedconfig.DefaultPath)
	if err != nil {
		l.Fatal("embed config load", log.ErrorField(err))
	}

	return cfg
}

// MustLoadConfig loads the application configuration from the provided path, treating all errors as fatal.
func MustLoadConfig(ctx context.Context, configPath string) *KubehoundConfig {
	l := log.Logger(ctx)
	cfg, err := NewConfig(ctx, viper.GetViper(), configPath)
	if err != nil {
		l.Fatal("config load", log.ErrorField(err))
	}

	return cfg
}

// MustLoadConfig loads the application configuration from the provided path, treating all errors as fatal.
func MustLoadInlineConfig(ctx context.Context) *KubehoundConfig {
	l := log.Logger(ctx)
	cfg, err := NewInlineConfig(ctx, viper.GetViper())
	if err != nil {
		l.Fatal("config load", log.ErrorField(err))
	}

	return cfg
}

func NewKubehoundConfig(ctx context.Context, configPath string, inLine bool) *KubehoundConfig {
	l := log.Logger(ctx)
	// Configuration initialization
	var cfg *KubehoundConfig
	switch {
	case len(configPath) != 0:
		l.Info("Loading application configuration from file", log.String("path", configPath))
		cfg = MustLoadConfig(ctx, configPath)
	case inLine:
		l.Info("Loading application from inline command")
		cfg = MustLoadInlineConfig(ctx)
	default:
		l.Info("Loading application configuration from default embedded")
		cfg = MustLoadEmbedConfig(ctx)
	}

	return cfg
}

// SetDefaultValues loads the default value from the different modules
func SetDefaultValues(ctx context.Context, v *viper.Viper) {
	// K8s Live collector module
	v.SetDefault(CollectorLivePageSize, DefaultK8sAPIPageSize)
	v.SetDefault(CollectorLivePageBufferSize, DefaultK8sAPIPageBufferSize)
	v.SetDefault(CollectorLiveRate, DefaultK8sAPIRateLimitPerSecond)
	v.SetDefault(CollectorNonInteractive, DefaultK8sAPINonInteractive)

	// File collector module
	v.SetDefault(CollectorFileArchiveNoCompress, DefaultArchiveNoCompress)

	// Default values for storage provider
	v.SetDefault("storage.wipe", true)
	v.SetDefault("storage.retry", DefaultRetry)
	v.SetDefault("storage.retry_delay", DefaultRetryDelay)

	// Disable Datadog telemetry by default
	v.SetDefault(TelemetryEnabled, false)

	// Default value for MongoDB
	v.SetDefault(MongoUrl, DefaultMongoUrl)
	v.SetDefault(MongoConnectionTimeout, DefaultConnectionTimeout)

	// Defaults values for JanusGraph
	v.SetDefault(JanusGraphUrl, DefaultJanusGraphUrl)
	v.SetDefault(JanusGrapTimeout, DefaultConnectionTimeout)
	v.SetDefault(JanusGraphWriterTimeout, defaultJanusGraphWriterTimeout)
	v.SetDefault(JanusGraphWriterMaxRetry, defaultJanusGraphWriterMaxRetry)
	v.SetDefault(JanusGraphWriterWorkerCount, defaultJanusGraphWriterWorkerCount)

	// Profiler values
	v.SetDefault(TelemetryProfilerPeriod, DefaultProfilerPeriod)
	v.SetDefault(TelemetryProfilerCPUDuration, DefaultProfilerCPUDuration)

	// Default values for graph builder
	v.SetDefault("builder.vertex.batch_size", DefaultVertexBatchSize)
	v.SetDefault("builder.vertex.batch_size_small", DefaultVertexBatchSizeSmall)
	v.SetDefault("builder.edge.worker_pool_size", DefaultEdgeWorkerPoolSize)
	v.SetDefault("builder.edge.worker_pool_capacity", DefaultEdgeWorkerPoolCapacity)
	v.SetDefault("builder.edge.batch_size", DefaultEdgeBatchSize)
	v.SetDefault("builder.edge.batch_size_small", DefaultEdgeBatchSizeSmall)
	v.SetDefault("builder.edge.batch_size_cluster_impact", DefaultEdgeBatchSizeClusterImpact)
	v.SetDefault("builder.stop_on_error", DefaultStopOnError)
	v.SetDefault("builder.edge.large_cluster_optimizations", DefaultLargeClusterOptimizations)

	v.SetDefault(IngestorAPIEndpoint, DefaultIngestorAPIEndpoint)
	v.SetDefault(IngestorAPIInsecure, DefaultIngestorAPIInsecure)
	v.SetDefault(IngestorBlobBucketURL, DefaultBucketName)
	v.SetDefault(IngestorTempDir, DefaultTempDir)
	v.SetDefault(IngestorMaxArchiveSize, DefaultMaxArchiveSize)
	v.SetDefault(IngestorArchiveName, DefaultArchiveName)

	SetLocalConfig(ctx, v)
}

// SetEnvOverrides enables environment variable overrides for the config.
func SetEnvOverrides(ctx context.Context, c *viper.Viper) {
	var res *multierror.Error
	l := log.Logger(ctx)

	// Enable changing file collector fields via environment variables
	res = multierror.Append(res, c.BindEnv("collector.type", "KH_COLLECTOR"))
	res = multierror.Append(res, c.BindEnv("collector.file.directory", "KH_COLLECTOR_DIR"))
	res = multierror.Append(res, c.BindEnv("collector.file.cluster", "KH_COLLECTOR_TARGET"))

	res = multierror.Append(res, c.BindEnv(MongoUrl, "KH_MONGODB_URL"))
	res = multierror.Append(res, c.BindEnv(JanusGraphUrl, "KH_JANUSGRAPH_URL"))
	res = multierror.Append(res, c.BindEnv(JanusGraphWriterMaxRetry, "KH_JANUSGRAPH_WRITER_MAX_RETRY"))
	res = multierror.Append(res, c.BindEnv(JanusGraphWriterTimeout, "KH_JANUSGRAPH_WRITER_TIMEOUT"))
	res = multierror.Append(res, c.BindEnv(JanusGraphWriterWorkerCount, "KH_JANUSGRAPH_WRITER_WORKER_COUNT"))

	res = multierror.Append(res, c.BindEnv(IngestorAPIEndpoint, "KH_INGESTOR_API_ENDPOINT"))
	res = multierror.Append(res, c.BindEnv(IngestorAPIInsecure, "KH_INGESTOR_API_INSECURE"))
	res = multierror.Append(res, c.BindEnv(IngestorBlobBucketURL, "KH_INGESTOR_BUCKET_URL"))
	res = multierror.Append(res, c.BindEnv(IngestorTempDir, "KH_INGESTOR_TEMP_DIR"))
	res = multierror.Append(res, c.BindEnv(IngestorMaxArchiveSize, "KH_INGESTOR_MAX_ARCHIVE_SIZE"))
	res = multierror.Append(res, c.BindEnv(IngestorArchiveName, "KH_INGESTOR_ARCHIVE_NAME"))
	res = multierror.Append(res, c.BindEnv(IngestorBlobRegion, "KH_INGESTOR_REGION"))

	res = multierror.Append(res, c.BindEnv("builder.vertex.batch_size", "KH_BUILDER_VERTEX_BATCH_SIZE"))
	res = multierror.Append(res, c.BindEnv("builder.vertex.batch_size_small", "KH_BUILDER_VERTEX_BATCH_SIZE_SMALL"))
	res = multierror.Append(res, c.BindEnv("builder.edge.batch_size", "KH_BUILDER_EDGE_BATCH_SIZE"))
	res = multierror.Append(res, c.BindEnv("builder.edge.batch_size_small", "KH_BUILDER_EDGE_BATCH_SIZE_SMALL"))

	res = multierror.Append(res, c.BindEnv(TelemetryStatsdUrl, "STATSD_URL"))
	res = multierror.Append(res, c.BindEnv(TelemetryTracerUrl, "TRACE_AGENT_URL"))

	if res.ErrorOrNil() != nil {
		l.Fatal("config environment override", log.ErrorField(res.ErrorOrNil()))
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
func NewConfig(ctx context.Context, v *viper.Viper, configPath string) (*KubehoundConfig, error) {
	// Configure default values
	SetDefaultValues(ctx, v)

	// Loading inLine config path
	v.SetConfigType(DefaultConfigType)
	v.SetConfigFile(configPath)

	// Configure environment variable override
	SetEnvOverrides(ctx, v)
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
func NewInlineConfig(ctx context.Context, v *viper.Viper) (*KubehoundConfig, error) {
	// Load default embedded config file
	SetDefaultValues(ctx, v)

	// Configure environment variable override
	SetEnvOverrides(ctx, v)

	kc, err := unmarshalConfig(v)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling config data: %w", err)
	}

	return kc, nil
}

// Load local config file if it exists, check for local file in current dir or in $HOME/.config/
// Not returning any error as it is not mandatory to have a local config file
func SetLocalConfig(ctx context.Context, v *viper.Viper) {
	l := log.Logger(ctx)

	v.SetConfigName(DefaultConfigName) // name of config file (without extension)
	v.SetConfigType(DefaultConfigType) // REQUIRED if the config file does not have the extension in the name
	v.AddConfigPath("$HOME/.config/")  // call multiple times to add many search paths
	v.AddConfigPath(".")               // optionally look for config in the working directory

	err := v.ReadInConfig()
	if err != nil {
		fp := fmt.Sprintf("%s.%s", DefaultConfigName, DefaultConfigType)
		l.Warn("No local config file was found", log.String("file", fp))
		l.Debug("Error reading config", log.ErrorField(err), log.String("file", fp))
	}
	l.Info("Using file for default config", log.String("path", viper.ConfigFileUsed()))
}

// NewEmbedConfig creates a new config instance from an embedded config file using viper.
func NewEmbedConfig(ctx context.Context, v *viper.Viper, configPath string) (*KubehoundConfig, error) {
	v.SetConfigType(DefaultConfigType)
	SetDefaultValues(ctx, v)

	// Configure environment variable override
	SetEnvOverrides(ctx, v)
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
