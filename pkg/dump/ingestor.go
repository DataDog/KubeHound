package dump

import (
	"context"
	"fmt"
	"path"

	"time"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/dump/pipeline"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DumpIngestor struct {
	directoryOutput string
	ResultName      string
	collector       collector.CollectorClient
	writer          writer.DumperWriter
}

const (
	OfflineDumpDateFormat = "2006-01-02-15-04-05"
	OfflineDumpPrefix     = "kubehound_"
)

// ./<clusterName>/kubehound_<clusterName>_<run_id>
func DumpIngestorResultName(clusterName string, runID string) string {
	return path.Join(clusterName, fmt.Sprintf("%s%s_%s", OfflineDumpPrefix, clusterName, runID))
}

// TODO: Rmove this function after the code split
func dumpIngestorResultName(clusterName string) string {
	return path.Join(clusterName, fmt.Sprintf("%s%s_%s", OfflineDumpPrefix, clusterName, time.Now().Format(OfflineDumpDateFormat)))
}

// func NewDumpIngestor(ctx context.Context, collector collector.CollectorClient, compression bool, directoryOutput string) (*DumpIngestor, error) {
func NewDumpIngestor(ctx context.Context, collector collector.CollectorClient, compression bool, directoryOutput string, runID *config.RunID) (*DumpIngestor, error) {
	// Generate path for the dump
	clusterName, err := getClusterName(ctx, collector)
	if err != nil {
		return nil, err
	}

	resultName := DumpIngestorResultName(clusterName, runID.String())

	dumpWriter, err := writer.DumperWriterFactory(ctx, compression, directoryOutput, resultName)
	if err != nil {
		return nil, fmt.Errorf("create collector writer: %w", err)
	}

	return &DumpIngestor{
		directoryOutput: directoryOutput,
		collector:       collector,
		writer:          dumpWriter,
		ResultName:      resultName,
	}, nil
}

func getClusterName(ctx context.Context, collector collector.CollectorClient) (string, error) {
	cluster, err := collector.ClusterInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("collector cluster info: %w", err)
	}

	return cluster.Name, nil
}

func (d *DumpIngestor) OutputPath() string {
	return d.writer.OutputPath()
}

func (d *DumpIngestor) DumpK8sObjects(ctx context.Context) error {
	spanDump, ctx := tracer.StartSpanFromContext(ctx, span.CollectorDump, tracer.Measured())
	var err error
	defer func() { spanDump.Finish(tracer.WithError(err)) }()

	ctx, pipeline, err := pipeline.NewPipelineDumpIngestor(ctx, d.collector, d.writer)
	if err != nil {
		return fmt.Errorf("create pipeline ingestor: %w", err)
	}

	spanDump.SetTag(tag.DumperWorkerNumberTag, pipeline.WorkerNumber)

	err = pipeline.Run(ctx)
	if err != nil {
		return fmt.Errorf("run pipeline ingestor: %w", err)
	}

	return pipeline.Wait(ctx)
}

// Close() is invoked by the collector to close all handlers used to dump k8s objects.
// The function flushes all writers and close all the handlers.
func (d *DumpIngestor) Close(ctx context.Context) error {
	err := d.writer.Flush(ctx)
	if err != nil {
		return fmt.Errorf("flush writer: %w", err)
	}

	return d.writer.Close(ctx)
}
