package config

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
)

// ClusterInfo encapsulates the target cluster information for the current run.
type ClusterInfo struct {
	Name string
}

func NewClusterInfo(_ context.Context) (*ClusterInfo, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	raw, err := kubeConfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("raw config get: %w", err)
	}

	return &ClusterInfo{
		Name: raw.CurrentContext,
	}, nil
}

func GetClusterName(ctx context.Context) (string, error) {
	cluster, err := NewClusterInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("collector cluster info: %w", err)
	}

	return cluster.Name, nil
}
