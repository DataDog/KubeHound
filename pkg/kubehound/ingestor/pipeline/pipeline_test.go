package pipeline

import (
	"github.com/DataDog/KubeHound/pkg/config"
)

var testID = config.NewRunID()
var testConfig = &config.KubehoundConfig{
	Dynamic: config.DynamicConfig{
		RunID: config.NewRunID(),
		Cluster: config.DynamicClusterInfo{
			Name: "test-cluster",
		},
	},
}
