package system

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DataDog/KubeHound/pkg/kubehound/core"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	CollectorTimeout = 5 * time.Minute
)

const (
	KubeHoundConfigPath = "kubehound.yaml"
)

// Optional syntactic sugar.
var __ = gremlingo.T__
var P = gremlingo.P

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
