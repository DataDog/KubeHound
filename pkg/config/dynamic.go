package config

import (
	"sync"
)

// DynamicConfig represent application configuration that can be updated at runtime.
type DynamicConfig struct {
	mu      sync.Mutex
	RunID   *RunID
	Cluster string
}

// DynamicOption is a functional option for configuring the dynamic config.
type DynamicOption func(c *DynamicConfig)

// WithClusterName is a functional option for configuring the cluster name.
func WithClusterName(cluster string) DynamicOption {
	return func(c *DynamicConfig) {
		c.Cluster = cluster
	}
}
