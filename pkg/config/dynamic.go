package config

import (
	"sync"

	"github.com/oklog/ulid/v2"
)

type DynamicConfig struct {
	mu      sync.Mutex
	RunID   ulid.ULID
	Cluster string
}

type DynamicOption func(c *DynamicConfig)

func WithClusterName(cluster string) DynamicOption {
	return func(c *DynamicConfig) {
		c.Cluster = cluster
	}
}
