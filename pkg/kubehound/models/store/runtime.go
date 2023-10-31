package store

import "github.com/DataDog/KubeHound/pkg/config"

// RuntimeInfo encapsulates information about the KubeHound run.
type RuntimeInfo struct {
	RunID   string `bson:"runID"`
	Cluster string `bson:"cluster"`
}

// Runtime extracts information about the KubeHound run from passed in config.
func Runtime(cfg *config.DynamicConfig) RuntimeInfo {
	return RuntimeInfo{
		RunID:   cfg.RunID.String(),
		Cluster: cfg.Cluster,
	}
}
