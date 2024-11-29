package config

import (
	"time"
)

const (
	DefaultJanusGraphUrl = "ws://localhost:8182/gremlin"

	defaultJanusGraphWriterTimeout     = 60 * time.Second
	defaultJanusGraphWriterMaxRetry    = 3
	defaultJanusGraphWriterWorkerCount = 1

	JanusGraphUrl               = "janusgraph.url"
	JanusGrapTimeout            = "janusgraph.connection_timeout"
	JanusGraphWriterTimeout     = "janusgraph.writer_timeout"
	JanusGraphWriterMaxRetry    = "janusgraph.writer_max_retry"
	JanusGraphWriterWorkerCount = "janusgraph.writer_worker_count"
)

// JanusGraphConfig configures JanusGraph specific parameters.
type JanusGraphConfig struct {
	URL               string        `mapstructure:"url"` // JanusGraph specific configuration
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`

	// JanusGraph vertex/edge writer configuration
	WriterTimeout     time.Duration `mapstructure:"writer_timeout"`
	WriterMaxRetry    int           `mapstructure:"writer_max_retry"`
	WriterWorkerCount int           `mapstructure:"writer_worker_count"`
}
