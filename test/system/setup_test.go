//nolint:all
package system

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DataDog/KubeHound/pkg/cmd"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/ingestor/api"
	"github.com/DataDog/KubeHound/pkg/ingestor/api/grpc"
	"github.com/DataDog/KubeHound/pkg/ingestor/notifier/noop"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller/blob"
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

func InitSetupTest(ctx context.Context) *providers.ProvidersFactoryConfig {
	err := cmd.InitializeKubehoundConfig(ctx, KubeHoundThroughDumpConfigPath, false, false)
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	khCfg, err := cmd.GetConfig()
	if err != nil {
		log.I.Fatal(err.Error())
	}

	// Initialisation of the p needed for the ingestion and the graph creation
	p, err := providers.NewProvidersFactoryConfig(ctx, khCfg)
	if err != nil {
		log.I.Fatalf("factory config creation: %v", err)
	}

	return p
}

type runArgs struct {
	runID         string
	cluster       string
	collectorPath string
	resultPath    string
}

func Dump(ctx context.Context, compress bool) (*config.KubehoundConfig, string) {
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
	viper.Set(config.CollectorNonInteractive, true)

	// Initialisation of the Kubehound config
	err = cmd.InitializeKubehoundConfig(ctx, "", true, false)
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	khCfg, err := cmd.GetConfig()
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	resultPath, err := core.DumpCore(ctx, khCfg, false)
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	return khCfg, resultPath
}

func RunLocal(ctx context.Context, runArgs *runArgs, compress bool, p *providers.ProvidersFactoryConfig) {
	// Saving the clusterName and collectorDir for the ingestion step
	// Those values are needed to run the ingestion pipeline
	collectorDir := runArgs.collectorPath
	clusterName := runArgs.cluster
	runID := runArgs.runID

	if compress {
		dryRun := false
		err := puller.ExtractTarGz(dryRun, runArgs.resultPath, collectorDir, config.DefaultMaxArchiveSize)
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

	err := cmd.InitializeKubehoundConfig(ctx, KubeHoundThroughDumpConfigPath, false, false)
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	khCfg, err := cmd.GetConfig()
	if err != nil {
		log.I.Fatal(err.Error())
	}

	// We need to flush the cache to prevent warning/error on the overwriting element in cache the  any conflict with the previous ingestion
	err = p.CacheProvider.Prepare(ctx)
	if err != nil {
		log.I.Fatalf("preparing cache provider: %v", err)
	}

	err = khCfg.ComputeDynamic(config.WithClusterName(clusterName), config.WithRunID(runID))
	if err != nil {
		log.I.Fatalf("collector client creation: %v", err)
	}

	err = p.IngestBuildData(ctx, khCfg)
	if err != nil {
		log.I.Fatalf("ingest build data: %v", err)
	}
}

func RunGRPC(ctx context.Context, runArgs *runArgs, p *providers.ProvidersFactoryConfig) {
	// Extracting info from Dump phase
	runID := runArgs.runID
	cluster := runArgs.cluster
	fileFolder := runArgs.collectorPath

	// Reseting the context to simulate a new ingestion from scratch
	ctx = context.Background()
	// Reseting the base tags
	tag.SetupBaseTags()
	// Reseting the viper config
	viper.Reset()
	err := cmd.InitializeKubehoundConfig(ctx, KubeHoundThroughDumpConfigPath, false, false)
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	khCfg, err := cmd.GetConfig()
	if err != nil {
		log.I.Fatal(err.Error())
	}

	khCfg.Ingestor.Blob.Bucket = fmt.Sprintf("file://%s", fileFolder)
	log.I.Info("Creating Blob Storage provider")
	puller, err := blob.NewBlobStorage(khCfg, khCfg.Ingestor.Blob)
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	log.I.Info("Creating Noop Notifier")
	noopNotifier := noop.NewNoopNotifier()

	log.I.Info("Creating Ingestor API")
	ingestorApi := api.NewIngestorAPI(khCfg, puller, noopNotifier, p)

	// Start the GRPC server
	go func() {
		err := grpc.Listen(ctx, ingestorApi)
		log.I.Fatalf(err.Error())
	}()

	// Starting ingestion of the dumped data
	err = core.CoreClientGRPCIngest(khCfg.Ingestor, cluster, runID)
	if err != nil {
		log.I.Fatalf(err.Error())
	}
}
func DumpAndRun(ctx context.Context, compress bool, p *providers.ProvidersFactoryConfig) {
	khCfg, resultPath := Dump(ctx, compress)

	// Extracting info from Dump phase
	runArgs := &runArgs{
		runID:         khCfg.Dynamic.RunID.String(),
		cluster:       khCfg.Dynamic.ClusterName,
		collectorPath: khCfg.Collector.File.Directory,
		resultPath:    resultPath,
	}

	RunLocal(ctx, runArgs, compress, p)
}

type FlatTestSuite struct {
	suite.Suite
}

func (s *FlatTestSuite) SetupSuite() {
	// Reseting the context to simulate a new ingestion from scratch
	ctx := context.Background()

	p := InitSetupTest(ctx)
	defer p.Close(ctx)

	DumpAndRun(ctx, false, p)
}

func (s *FlatTestSuite) TestRun() {
	RunTestSuites(s.T())
}

type CompressTestSuite struct {
	suite.Suite
}

func (s *CompressTestSuite) SetupSuite() {
	// Reseting the context to simulate a new ingestion from scratch
	ctx := context.Background()

	p := InitSetupTest(ctx)
	defer p.Close(ctx)

	DumpAndRun(ctx, true, p)
}

func (s *CompressTestSuite) TestRun() {
	RunTestSuites(s.T())
}

type MultipleIngestioTestSuite struct {
	suite.Suite
}

func (s *MultipleIngestioTestSuite) SetupSuite() {
	// Reseting the context to simulate a new ingestion from scratch
	ctx := context.Background()

	p := InitSetupTest(ctx)
	defer p.Close(ctx)

	// Simulating multiple ingestion (twice the same cluster)
	DumpAndRun(ctx, true, p)
	DumpAndRun(ctx, false, p)
}

func (s *MultipleIngestioTestSuite) TestRun() {
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
	cmd.InitializeKubehoundConfig(ctx, KubeHoundConfigPath, true, false)
	khCfg, err := cmd.GetConfig()
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	core.CoreLive(ctx, khCfg)
}

func (s *LiveTestSuite) TestRun() {
	RunTestSuites(s.T())
}

type GRPCTestSuite struct {
	suite.Suite
}

func (s *GRPCTestSuite) SetupSuite() {
	// Reseting the context to simulate a new ingestion from scratch
	ctx := context.Background()

	p := InitSetupTest(ctx)
	defer p.Close(ctx)

	khCfg, _ := Dump(ctx, true)

	// Extracting info from Dump phase
	runArgs := &runArgs{
		runID:         khCfg.Dynamic.RunID.String(),
		cluster:       khCfg.Dynamic.ClusterName,
		collectorPath: khCfg.Collector.File.Directory,
	}

	RunGRPC(ctx, runArgs, p)

	// Reingesting the same to trigger the error
	// Starting ingestion of the dumped data
	err := cmd.InitializeKubehoundConfig(ctx, KubeHoundThroughDumpConfigPath, false, false)
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	khCfg, err = cmd.GetConfig()
	if err != nil {
		log.I.Fatal(err.Error())
	}

	err = core.CoreClientGRPCIngest(khCfg.Ingestor, runArgs.cluster, runArgs.runID)
	s.ErrorContains(err, api.ErrAlreadyIngested.Error())
}

func (s *GRPCTestSuite) TestRun() {
	RunTestSuites(s.T())
}

// TODO: needs to add support of runID/cluster in all janusgraph requests system-tests to avoid collision
// func TestMultipleIngestioTestSuite(t *testing.T) {
// 	suite.Run(t, new(MultipleIngestioTestSuite))
// }

func TestCompressTestSuite(t *testing.T) {
	suite.Run(t, new(CompressTestSuite))
}

func TestLiveTestSuite(t *testing.T) {
	suite.Run(t, new(LiveTestSuite))
}

func TestFlatTestSuite(t *testing.T) {
	suite.Run(t, new(FlatTestSuite))
}

func TestGRPCTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCTestSuite))
}
