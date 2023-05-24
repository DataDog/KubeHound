package system

import (
	"context"
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

func TestMain(m *testing.M) {
	cmdCtx, cmdCancel := context.WithTimeout(context.Background(), CollectorTimeout)
	defer cmdCancel()

	// Run the kind-collect.sh script. NOTE: this is a temporary solution until the K8s API collector is
	// completed, at which point it should be invoked here.
	cmd := exec.CommandContext(cmdCtx, CollectorScriptPath, CollectorOutputDir)
	if err := cmd.Run(); err != nil {
		log.I.Fatalf("Collector script execution: %v", err)
	}

	// Run the ingest
	err := core.Launch(context.Background(), core.WithConfigPath(KubeHoundConfigPath))
	if err != nil {
		log.I.Fatalf("KubeHound core: %v", err)
	}

	// Run the test suite
	os.Exit(m.Run())
}
