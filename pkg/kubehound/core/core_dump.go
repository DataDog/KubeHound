package core

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/dump"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller/blob"
	"github.com/DataDog/KubeHound/pkg/telemetry/events"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// DumpCore is the main entry point for the KubeHound dump process.
// It collects all the necessary data from the Kubernetes API and dumps it to a file.
// If upload is true, it will upload the file to the configured blob storage.
// It returns the path to the dumped file/dir (only used for the system tests)
func DumpCore(ctx context.Context, khCfg *config.KubehoundConfig, upload bool) (string, error) {
	l := log.Logger(ctx)
	start := time.Now()
	var err error

	span, ctx := tracer.StartSpanFromContext(ctx, span.DumperLaunch, tracer.Measured())
	defer func() {
		span.Finish(tracer.WithError(err))
	}()

	khCfg.Collector.Type = config.CollectorTypeK8sAPI

	clusterName, err := config.GetClusterName(ctx)
	if err != nil {
		return "", fmt.Errorf("collector cluster info: %w", err)
	}
	khCfg.Dynamic.ClusterName = clusterName

	events.PushEvent(
		fmt.Sprintf("Starting KubeHound dump for %s", clusterName),
		fmt.Sprintf("Starting KubeHound dump for %s", clusterName),
		[]string{
			tag.ActionType(events.DumperRun),
		},
	)

	filePath, err := runLocalDump(ctx, khCfg)
	if err != nil {
		return "", err
	}
	l.Info("result saved to file", log.String("path", filePath))

	if upload {
		// Clean up the temporary directory when done
		defer func() {
			err = os.RemoveAll(khCfg.Collector.File.Directory)
			if err != nil {
				l.Error("Failed to remove temporary directory", log.ErrorField(err))
			}
		}()
		puller, err := blob.NewBlobStorage(khCfg, khCfg.Ingestor.Blob)
		if err != nil {
			return "", err
		}

		err = puller.Put(ctx, filePath, clusterName, khCfg.Dynamic.RunID.String())
		if err != nil {
			return "", err
		}
	}

	events.PushEvent(
		fmt.Sprintf("Finish KubeHound dump for %s", clusterName),
		fmt.Sprintf("KubeHound dump run has been completed in %s", time.Since(start)),
		[]string{
			tag.ActionType(events.DumperRun),
		},
	)
	l.Info("KubeHound dump run has been completed", log.Duration("duration", time.Since(start)))

	return filePath, nil
}

// Running the local dump of the k8s objects (dumper pipeline)
// It returns the path to the dumped file/dir (only used for the system tests)
func runLocalDump(ctx context.Context, khCfg *config.KubehoundConfig) (string, error) {
	l := log.Logger(ctx)
	l.Info("Loading Kubernetes data collector client")
	collect, err := collector.ClientFactory(ctx, khCfg)
	if err != nil {
		return "", fmt.Errorf("collector client creation: %w", err)
	}
	defer func() { collect.Close(ctx) }()
	l.Info("Loaded collector client", log.String("collector", collect.Name()))

	// Create the dumper instance
	collectorLocalOutputDir := khCfg.Collector.File.Directory
	collectorLocalCompress := !khCfg.Collector.File.Archive.NoCompress
	l.Info("Dumping cluster info to directory", log.String("cluster_name", khCfg.Dynamic.ClusterName), log.String("path", collectorLocalOutputDir))
	dumpIngestor, err := dump.NewDumpIngestor(ctx, collect, collectorLocalCompress, collectorLocalOutputDir, khCfg.Dynamic.RunID)
	if err != nil {
		return "", fmt.Errorf("create dumper: %w", err)
	}
	defer func() { dumpIngestor.Close(ctx) }()

	// Dumping all k8s objects using the API
	err = dumpIngestor.DumpK8sObjects(ctx)
	if err != nil {
		return "", fmt.Errorf("dump k8s object: %w", err)
	}

	return dumpIngestor.OutputPath(), nil
}
