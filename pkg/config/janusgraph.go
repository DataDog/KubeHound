package config

import "time"

// JanusGraphConfig configures JanusGraph specific parameters.
type JanusGraphConfig struct {
	URL               string        `mapstructure:"url"` // JanusGraph specific configuration
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
}
