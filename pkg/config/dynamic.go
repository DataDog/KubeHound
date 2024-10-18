package config

import (
	"fmt"
	"sync"
)

const (
	DynamicRunID       = "dynamic.run_id"
	DynamicClusterName = "dynamic.cluster_name"
)

// DynamicConfig represent application configuration that can be updated at runtime.
type DynamicConfig struct {
	mu          sync.Mutex
	RunID       *RunID `mapstructure:"run_id"`
	ClusterName string `mapstructure:"cluster_name"`
	Service     string `mapstructure:"service"`
}

func (d *DynamicConfig) HealthCheck() error {
	if d.ClusterName == "" {
		return fmt.Errorf("missing cluster name")
	}

	if d.RunID == nil {
		return fmt.Errorf("missing run id")
	}

	return nil
}

// DynamicOption is a functional option for configuring the dynamic config.
type DynamicOption func() (func(*DynamicConfig), error)

// Wrapper around the dynamic config to provide error feedback
func success(opt func(*DynamicConfig)) DynamicOption {
	return func() (func(*DynamicConfig), error) {
		return opt, nil
	}
}

// Wrapper around the dynamic config to provide error feedback
func failure(err error) DynamicOption {
	return func() (func(*DynamicConfig), error) {
		return nil, err
	}
}

// WithRunID is a functional option for configuring the runID (using in KHaaS).
func WithRunID(runID string) DynamicOption {
	val, err := LoadRunID(runID)
	if err != nil {
		return failure(fmt.Errorf("loading run id: %w", err))
	}

	return success(func(c *DynamicConfig) {
		c.RunID = val
	})
}

// WithClusterName is a functional option for configuring the cluster name.
func WithClusterName(cluster string) DynamicOption {
	return success(func(c *DynamicConfig) {
		c.ClusterName = cluster
	})
}
