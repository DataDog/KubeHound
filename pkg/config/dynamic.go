package config

import (
	"sync"
)

type DynamicConfig struct {
	mu      sync.Mutex
	RunID   *RunID
	Cluster string
}

type DynamicOption func(c *DynamicConfig)

func WithClusterName(cluster string) DynamicOption {
	return func(c *DynamicConfig) {
		c.Cluster = cluster
	}
}
