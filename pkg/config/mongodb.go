package config

import (
	"time"
)

const (
	DefaultMongoUrl = "mongodb://localhost:27017"

	MongoUrl               = "mongodb.url"
	MongoConnectionTimeout = "mongodb.connection_timeout"
)

// MongoDBConfig configures mongodb specific parameters.
type MongoDBConfig struct {
	URL               string        `mapstructure:"url"` // Mongodb specific configuration
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
}
