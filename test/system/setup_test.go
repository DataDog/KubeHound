package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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

// cleanupCollected deletes data collected to enable idempotency for local calls.
func cleanupCollected() {
	err := os.RemoveAll(CollectorOutputDir)
	if err != nil {
		log.I.Errorf("Collector data cleanup: %v", err)
	}
}

// runKubeHound runs the collector against the local kind cluster, then runs KubeHound to create
// an attack graph that can be queried in the individual system tests.
func runKubeHound() error {
	cmdCtx, cmdCancel := context.WithTimeout(context.Background(), CollectorTimeout)
	defer cmdCancel()

	// Run the kind-collect.sh script. NOTE: this is a temporary solution until the K8s API collector is
	// completed, at which point it should be invoked here.
	cmd := exec.CommandContext(cmdCtx, CollectorScriptPath, CollectorOutputDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	defer cleanupCollected()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("collector script execution: %v", err)
	}

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
