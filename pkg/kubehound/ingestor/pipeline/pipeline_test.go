package pipeline

import (
	"github.com/DataDog/KubeHound/pkg/config"
)

var testID = config.NewRunID()
var testConfig = &config.KubehoundConfig{
	Dynamic: config.DynamicConfig{
		RunID:       config.NewRunID(),
		ClusterName: "test-cluster",
	},
}
