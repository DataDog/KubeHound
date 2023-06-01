package config

// StatsdConfig configures mongodb specific parameters.
type StatsdConfig struct {
	URL string `mapstructure:"url"` // Mongodb specific configuration
}
