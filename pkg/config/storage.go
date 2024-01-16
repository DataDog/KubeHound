package config

import (
	"time"
)

const (
	DefaultRetry             int           = 10 // number of tries before failing
	DefaultRetryDelay        time.Duration = 10 * time.Second
	DefaultConnectionTimeout time.Duration = 30 * time.Second
)

type StorageConfig struct {
	Retry      int           `mapstructure:"retry"`
	RetryDelay time.Duration `mapstructure:"retry_delay"`
	Wipe       bool          `mapstructure:"wipe"`
}
