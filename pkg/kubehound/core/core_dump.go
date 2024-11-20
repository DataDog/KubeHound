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
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// DumpCore is the main entry point for the KubeHound dump process.
// It collects all the necessary data from the Kubernetes API and dumps it to a file.
// If upload is true, it will upload the file to the configured blob storage.
// It returns the path to the dumped file/dir (only used for the system tests)
func DumpCore(ctx context.Context, khCfg *config.KubehoundConfig, upload bool) (string, error) {
	l := log.Logger(ctx)

	clusterName, err := config.GetClusterName(ctx)
	defer func() {
		if err != nil {
			errMsg := fmt.Errorf("fatal error: %w", err)
			l.Error("Error occurred", log.ErrorField(errMsg))
			_ = events.PushEvent(ctx, events.DumpFailed, fmt.Sprintf("%s", errMsg))
		}
	}()
	if err != nil {
		return "", fmt.Errorf("collector cluster info: %w", err)
	}
	khCfg.Dynamic.ClusterName = clusterName
	ctx = context.WithValue(ctx, log.ContextFieldCluster, clusterName)
	ctx = context.WithValue(ctx, log.ContextFieldRunID, khCfg.Dynamic.RunID.String())

	start := time.Now()

	span, ctx := span.SpanRunFromContext(ctx, span.DumperLaunch)
	span.SetTag(ext.ManualKeep, true)
	l = log.Logger(ctx)
	defer func() {
		span.Finish(tracer.WithError(err))
	}()

	khCfg.Collector.Type = config.CollectorTypeK8sAPI

	_ = events.PushEvent(ctx, events.DumpStarted, "")

	filePath, err := runLocalDump(ctx, khCfg)
	if err != nil {
		return "", err
	}
	l.Info("result saved to file", log.String(log.FieldPathKey, filePath))

	if upload {
		// Clean up the temporary directory when done
		defer func() {
			// This error is scope to the defer and not be handled by the other defer function
			err := os.RemoveAll(khCfg.Collector.File.Directory)
			if err != nil {
				errMsg := fmt.Errorf("Failed to remove temporary directory: %w", err)
				l.Error("Failed to remove temporary directory", log.ErrorField(err))
				_ = events.PushEvent(ctx, events.DumpFailed, fmt.Sprintf("%s", errMsg))
			}
		}()
		var puller *blob.BlobStore
		puller, err = blob.NewBlobStorage(khCfg, khCfg.Ingestor.Blob)
		if err != nil {
			return "", err
		}

		err = puller.Put(ctx, filePath, clusterName, khCfg.Dynamic.RunID.String())
		if err != nil {
			return "", err
		}
	}

	text := fmt.Sprintf("KubeHound dump run has been completed in %s", time.Since(start))
	_ = events.PushEvent(ctx, events.DumpFinished, text)
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
	ctx = context.WithValue(ctx, log.ContextFieldComponent, collect.Name())
	l.Info("Loaded collector client")

	// Create the dumper instance
	collectorLocalOutputDir := khCfg.Collector.File.Directory
	collectorLocalCompress := !khCfg.Collector.File.Archive.NoCompress
	l.Info("Dumping cluster info to directory", log.String(log.FieldPathKey, collectorLocalOutputDir))
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
