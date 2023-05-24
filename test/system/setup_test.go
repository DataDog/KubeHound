package system

import (
	"context"
	"os"
	"testing"

	"github.com/DataDog/KubeHound/pkg/kubehound/core"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

// Create the pods here??
// Init pattern to create a list of resources? Then this consumes and creates all the resources,
// before running the collector and staring the test suite

func TestMain(m *testing.M) {
	// Verify kind is running
	// TODO

	// Configuration file path pointing to the collected kind K8s data
	testConfig := "kubehound.yaml"

	// Run the ingest
	err := core.Launch(context.Background(), core.WithConfigPath(testConfig))
	if err != nil {
		log.I.Fatalf("KubeHound core: %w", err)
	}

	// Run the test suite
	os.Exit(m.Run())
}
