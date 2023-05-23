package config

import (
	"fmt"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/spf13/viper"
)

const (
	DefaultConfigType = "yaml"
	DefaultConfigPath = "etc/kubehound.yaml"
)

// KubehoundConfig defines the top-level application configuration for KubeHound.
type KubehoundConfig struct {
	Collector CollectorConfig `mapstructure:"collector"` // Collector configuration
}

// MustLoadDefaultConfig loads the default application configuration, treating all errors as fatal.
func MustLoadDefaultConfig() *KubehoundConfig {
	cfg, err := NewConfig(DefaultConfigPath)
	if err != nil {
		log.I.Fatalf("config load: %w", err)
	}

	return cfg
}

// NewConfig creates a new config instance from the provided file using viper.
func NewConfig(configPath string) (*KubehoundConfig, error) {
	c := viper.New()
	c.SetConfigType(DefaultConfigType)
	c.SetConfigFile(configPath)

	if err := c.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	kc := KubehoundConfig{}
	if err := c.Unmarshal(&kc); err != nil {
		return nil, fmt.Errorf("unmarshaling config data: %w", err)
	}

	return &kc, nil
}
