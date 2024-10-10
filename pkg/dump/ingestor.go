package dump

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/dump/pipeline"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DumpIngestor struct {
	collector collector.CollectorClient
	writer    writer.DumperWriter
}

func NewDumpIngestor(ctx context.Context, collector collector.CollectorClient, compression bool, directoryOutput string, runID *config.RunID) (*DumpIngestor, error) {
	// Generate path for the dump
	clusterName, err := getClusterName(ctx, collector)
	if err != nil {
		return nil, err
	}

	dumpResult, err := NewDumpResult(clusterName, runID.String(), compression)
	if err != nil {
		return nil, fmt.Errorf("create dump result: %w", err)
	}

	dumpWriter, err := writer.DumperWriterFactory(ctx, compression, directoryOutput, dumpResult.GetFullPath())
	if err != nil {
		return nil, fmt.Errorf("create collector writer: %w", err)
	}

	return &DumpIngestor{
		collector: collector,
		writer:    dumpWriter,
	}, nil
}

func getClusterName(ctx context.Context, collector collector.CollectorClient) (string, error) {
	cluster, err := collector.ClusterInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("collector cluster info: %w", err)
	}

	return cluster.Name, nil
}

func (d *DumpIngestor) Metadata() (collector.Metadata, error) {
	path := filepath.Join(d.writer.OutputPath(), collector.MetadataPath)
	data, err := os.ReadFile(path)
	if err != nil {
		return collector.Metadata{}, err
	}

	md := collector.Metadata{}
	err = json.Unmarshal(data, &md)
	if err != nil {
		return collector.Metadata{}, err
	}

	return md, nil
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

	return pipeline.WaitAndClose(ctx)
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

// Backward Compatibility: Extracting the metadata from the path
const (
	DumpResultFilenameRegex = DumpResultPrefix + DumpResultClusterNameRegex + "_" + DumpResultRunIDRegex + DumpResultExtensionRegex
	DumpResultPathRegex     = DumpResultClusterNameRegex + "/" + DumpResultFilenameRegex
)

func ParsePath(path string) (*DumpResult, error) {
	//log.I..Warnf("[Backward Compatibility] Extracting the metadata from the path: %s", path)

	// ./<clusterName>/kubehound_<clusterName>_<run_id>[.tar.gz]
	// re := regexp.MustCompile(`([a-z0-9\.\-_]+)/kubehound_([a-z0-9\.-_]+)_([a-z0-9]{26})\.?([a-z0-9\.]+)?`)
	re := regexp.MustCompile(DumpResultPathRegex)
	if !re.MatchString(path) {
		return nil, fmt.Errorf("Invalid path provided: %q", path)
	}

	matches := re.FindStringSubmatch(path)
	// The cluster name should match (parent dir and in the filename)
	if matches[1] != matches[2] {
		return nil, fmt.Errorf("Cluster name does not match in the path provided: %q", path)
	}

	clusterName := matches[1]
	runID := matches[3]
	extension := matches[4]

	isCompressed := false
	if extension != "" {
		isCompressed = true
	}
	result, err := NewDumpResult(clusterName, runID, isCompressed)
	if err != nil {
		return nil, err
	}

	return result, nil
}
