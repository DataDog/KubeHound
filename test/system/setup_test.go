//nolint:all
package system

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DataDog/KubeHound/pkg/cmd"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	"github.com/DataDog/KubeHound/pkg/kubehound/core"
	"github.com/DataDog/KubeHound/pkg/kubehound/libkube"
	"github.com/DataDog/KubeHound/pkg/kubehound/providers"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

const (
	CollectorTimeout = 5 * time.Minute
)

const (
	KubeHoundConfigPath            = "kubehound.yaml"
	KubeHoundThroughDumpConfigPath = "kubehound_dump.yaml"
)

// Optional syntactic sugar.
var __ = gremlingo.T__
var P = gremlingo.P

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func RunTestSuites(t *testing.T) {
	suite.Run(t, new(EdgeTestSuite))
	suite.Run(t, new(VertexTestSuite))
	suite.Run(t, new(DslTestSuite))
}

func DumpAndRun(compress bool) {
	ctx := context.Background()
	var err error

	// Setting the base tags
	tag.SetupBaseTags()
	// Reseting the cache mechanism for the creation of the default accounts (system:nodes)
	libkube.ResetOnce()

	// Simulating inline command
	dumpCmd := &cobra.Command{
		Use: "local",
	}
	cmd.InitDumpCmd(dumpCmd)

	viper.Set(config.CollectorFileArchiveFormat, compress)

	tmpDir, err := os.MkdirTemp("/tmp/", "kh-system-tests-*")
	if err != nil {
		log.I.Fatalf(err.Error())
	}
	viper.Set(config.CollectorFileDirectory, tmpDir)

	// Initialisation of the Kubehound config
	cmd.InitializeKubehoundConfig(ctx, "", true)
	khCfg, err := cmd.GetConfig()
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	resultPath, err := core.DumpCore(ctx, khCfg, false)
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	// Saving the clusterName and collectorDir for the ingestion step
	// Those values are needed to run the ingestion pipeline
	collectorDir := khCfg.Collector.File.Directory
	clusterName := khCfg.Dynamic.ClusterName
	runID := khCfg.Dynamic.RunID

	if compress {
		err := puller.ExtractTarGz(resultPath, collectorDir, config.DefaultMaxArchiveSize)
		if err != nil {
			log.I.Fatalf(err.Error())
		}
	}

	// Reseting the context to simulate a new ingestion from scratch
	ctx = context.Background()
	// Reseting the base tags
	tag.SetupBaseTags()
	// Reseting the viper config
	viper.Reset()

	// Setting the collectorDir, clusterName and runID needed for the ingestion step
	// This information is used by the grpc server to run the ingestion
	viper.Set(config.CollectorFileDirectory, collectorDir)
	viper.Set(config.CollectorFileClusterName, clusterName)
	viper.Set(config.DynamicRunID, runID)

	err = cmd.InitializeKubehoundConfig(ctx, KubeHoundThroughDumpConfigPath, false)
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	khCfg, err = cmd.GetConfig()
	if err != nil {
		log.I.Fatal(err.Error())
	}

	err = khCfg.ComputeDynamic(config.WithClusterName(clusterName))
	if err != nil {
		log.I.Fatalf("collector client creation: %v", err)
	}

	// Initialisation of the providers needed for the ingestion and the graph creation
	fc, err := providers.NewProvidersFactoryConfig(ctx, khCfg)
	if err != nil {
		log.I.Fatalf("factory config creation: %v", err)
	}
	defer fc.Close(ctx)

	err = fc.IngestBuildData(ctx, khCfg)
	if err != nil {
		log.I.Fatalf("ingest build data: %v", err)
	}
}

type FlatTestSuite struct {
	suite.Suite
}

func (s *FlatTestSuite) SetupSuite() {
	DumpAndRun(false)
}

func (s *FlatTestSuite) TestRun() {
	RunTestSuites(s.T())
}

type CompressTestSuite struct {
	suite.Suite
}

func (s *CompressTestSuite) SetupSuite() {
	DumpAndRun(true)
}

func (s *CompressTestSuite) TestRun() {
	RunTestSuites(s.T())
}

type LiveTestSuite struct {
	suite.Suite
}

// runKubeHound runs the collector against the local kind cluster, then runs KubeHound to create
// an attack graph that can be queried in the individual system tests.
func (s *LiveTestSuite) SetupSuite() {
	ctx := context.Background()
	libkube.ResetOnce()

	// Initialisation of the Kubehound config
	cmd.InitializeKubehoundConfig(ctx, KubeHoundConfigPath, true)
	khCfg, err := cmd.GetConfig()
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	core.CoreLive(ctx, khCfg)
}

func (s *LiveTestSuite) TestRun() {
	RunTestSuites(s.T())
}

func TestCompressTestSuite(t *testing.T) {
	suite.Run(t, new(CompressTestSuite))
}

func TestLiveTestSuite(t *testing.T) {
	suite.Run(t, new(LiveTestSuite))
}

func TestFlatTestSuite(t *testing.T) {
	suite.Run(t, new(FlatTestSuite))
}
