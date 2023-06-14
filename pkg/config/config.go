package config

import (
	"fmt"
	"os"

	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/spf13/viper"
)

var (
	BuildVersion string // This should be overwritten by the go build -X flags
)

const (
	DefaultConfigType = "yaml"
	DefaultConfigPath = "etc/kubehound.yaml"
)

// KubehoundConfig defines the top-level application configuration for KubeHound.
type KubehoundConfig struct {
	Collector  CollectorConfig  `mapstructure:"collector"`  // Collector configuration
	MongoDB    MongoDBConfig    `mapstructure:"mongodb"`    // MongoDB configuration
	JanusGraph JanusGraphConfig `mapstructure:"janusgraph"` // JanusGraph configuration
	Storage    StorageConfig    `mapstructure:"janusgraph"` // Global param for all storage provider
	Telemetry  TelemetryConfig  `mapstructure:"telemetry"`  // telemetry configuration, contains statsd and other sub structures
}

// MustLoadDefaultConfig loads the default application configuration, treating all errors as fatal.
func MustLoadDefaultConfig() *KubehoundConfig {
	return MustLoadConfig(DefaultConfigPath)
}

// MustLoadtConfig loads the application configuration from the provided path, treating all errors as fatal.
func MustLoadConfig(configPath string) *KubehoundConfig {
	cfg, err := NewConfig(configPath)
	if err != nil {
		log.I.Fatalf("config load: %v", err)
	}

	return cfg
}

// SetDefaultValues loads the default value from the different modules
func SetDefaultValues(c *viper.Viper) {
	// K8s Live collector module
	c.SetDefault("collector.live.page_size", globals.DefaultK8sAPIPageSize)
	c.SetDefault("collector.live.page_buffer_size", globals.DefaultK8sAPIPageBufferSize)
	c.SetDefault("collector.live.rate_limit_per_second", globals.DefaultK8sAPIRateLimitPerSecond)

	// Connector Retry default for storedb
	c.SetDefault("storage.retry", globals.DefaultRetry)
	c.SetDefault("storage.retry_delay", globals.DefaultRetryDelay)
}

// NewConfig creates a new config instance from the provided file using viper.
func NewConfig(configPath string) (*KubehoundConfig, error) {
	c := viper.New()
	c.SetConfigType(DefaultConfigType)
	c.SetConfigFile(configPath)
	SetDefaultValues(c)
	if err := c.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	kc := KubehoundConfig{}
	if err := c.Unmarshal(&kc); err != nil {
		return nil, fmt.Errorf("unmarshaling config data: %w", err)
	}

	return &kc, nil
}

func IsCI() bool {
	return os.Getenv("CI") != ""
}
