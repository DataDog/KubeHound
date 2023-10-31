package pipeline

import (
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/oklog/ulid/v2"
)

var testID = ulid.Make()
var testConfig = &config.KubehoundConfig{
	Dynamic: config.DynamicConfig{
		RunID:   testID,
		Cluster: "test-cluster",
	},
}
