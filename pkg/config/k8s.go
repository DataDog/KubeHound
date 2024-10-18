package config

import (
	"context"
	"fmt"
	"os"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	clusterNameEnvVar = "KH_K8S_CLUSTER_NAME"
)

// ClusterInfo encapsulates the target cluster information for the current run.
type ClusterInfo struct {
	Name string
}

func NewClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	// Testing if running from pod
	// Using an environment variable to get the cluster name as it is not provided in the pod configuration
	l := log.Logger(ctx)
	clusterName := os.Getenv(clusterNameEnvVar)
	if clusterName != "" {
		l.Warn("Using cluster name from environment variable", log.String("env_var", clusterNameEnvVar), log.String(log.FieldClusterKey, clusterName))

		return &ClusterInfo{
			Name: clusterName,
		}, nil
	}

	// Testing if running from outside the cluster
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
