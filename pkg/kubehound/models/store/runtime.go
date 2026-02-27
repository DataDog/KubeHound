package store

import "github.com/DataDog/KubeHound/pkg/config"

// Cluster encapsulates information about the target Kubernetes cluster.
type Cluster struct {
	Name         string
	VersionMajor string
	VersionMinor string
}

// RuntimeInfo encapsulates information about the KubeHound run.
type RuntimeInfo struct {
	RunID   string
	Cluster Cluster
}

// Runtime extracts information about the KubeHound run from passed in config.
func Runtime(cfg *config.DynamicConfig) RuntimeInfo {
	return RuntimeInfo{
		RunID: cfg.RunID.String(),
		Cluster: Cluster{
			Name:         cfg.Cluster.Name,
			VersionMajor: cfg.Cluster.VersionMajor,
			VersionMinor: cfg.Cluster.VersionMinor,
		},
	}
}
