package config

type StorageConfig struct {
	Retry      int `mapstructure:"retry"`
	RetryDelay int `mapstructure:"retry_delay"`
}
