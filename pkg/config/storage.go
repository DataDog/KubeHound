package config

import "time"

type StorageConfig struct {
	Retry      int           `mapstructure:"retry"`
	RetryDelay time.Duration `mapstructure:"retry_delay"`
}
