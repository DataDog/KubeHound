package config

import "time"

// MongoDBConfig configures mongodb specific parameters.
type JanusGraphConfig struct {
	URL               string        `mapstructure:"url"` // Mongodb specific configuration
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
}
