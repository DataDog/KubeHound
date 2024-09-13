package config

import (
	"time"
)

const (
	DefaultJanusGraphUrl = "ws://localhost:8182/gremlin"

	JanusGraphUrl    = "janusgraph.url"
	JanusGrapTimeout = "janusgraph.connection_timeout"
)

// JanusGraphConfig configures JanusGraph specific parameters.
type JanusGraphConfig struct {
	URL               string        `mapstructure:"url"` // JanusGraph specific configuration
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
}
