package system

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DataDog/KubeHound/pkg/kubehound/core"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

const (
	CollectorTimeout = 5 * time.Minute
)

const (
	KubeHoundConfigPath = "kubehound.yaml"
	CollectorScriptPath = "./kind-collect.sh"
	CollectorOutputDir  = "kind-collect"
)

// runKubeHound runs the collector against the local kind cluster, then runs KubeHound to create
// an attack graph that can be queried in the individual system tests.
func runKubeHound() error {
	// Run the ingest
	err := core.Launch(context.Background(), core.WithConfigPath(KubeHoundConfigPath))
	if err != nil {
		return fmt.Errorf("KubeHound launch: %v", err)
	}

	return nil
}

func TestMain(m *testing.M) {
	if err := runKubeHound(); err != nil {
		log.I.Fatalf(err.Error())
	}

	// Run the test suite
	os.Exit(m.Run())
}
