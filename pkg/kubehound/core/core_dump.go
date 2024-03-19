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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DumpConfig struct {
	RunID       *config.RunID
	ClusterName string
	ResPath     string
}

func DumpCore(ctx context.Context, cmd *cobra.Command) (*DumpConfig, error) {
	start := time.Now()
	var err error

	ctx, lc := NewLaunchConfig(ctx, "", true)
	span, ctx := tracer.StartSpanFromContext(ctx, span.DumperLaunch, tracer.Measured())
	err = lc.Initialize(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("initialize config: %w", err)
	}
	defer func() {
		span.Finish(tracer.WithError(err))
		lc.Close()
	}()

	lc.Cfg.Collector.Type = config.CollectorTypeK8sAPI

	clusterName, err := collector.GetClusterName(ctx)
	if err != nil {
		return nil, fmt.Errorf("collector cluster info: %w", err)
	}

	events.PushEvent(
		fmt.Sprintf("Starting KubeHound dump for %s", clusterName),
		fmt.Sprintf("Starting KubeHound dump for %s", clusterName),
		[]string{
			tag.ActionType(events.DumperRun),
		},
	)

	filePath, err := runLocalDump(ctx, lc)
	if err != nil {
		return nil, err
	}
	log.I.Infof("result %s", filePath)

	if cmd.Use == "s3" {
		// Clean up the temporary directory when done
		defer func() {
			err = os.RemoveAll(viper.GetString(config.CollectorLocalOutputDir))
			if err != nil {
				log.I.Errorf("Failed to remove temporary directory: %v", err)
			}
		}()
		lc.Cfg.Ingestor.BucketName = viper.GetString(config.CollectorS3Bucket)
		puller, err := blob.NewBlobStoragePuller(lc.Cfg)
		if err != nil {
			return nil, err
		}

		err = puller.Put(ctx, filePath, clusterName, lc.RunID.String())
		if err != nil {
			return nil, err
		}
	}

	events.PushEvent(
		fmt.Sprintf("Finish KubeHound dump for %s", clusterName),
		fmt.Sprintf("KubeHound dump run has been completed in %s", time.Since(start)),
		[]string{
			tag.ActionType(events.DumperRun),
		},
	)
	log.I.Infof("KubeHound dump run has been completed in %s", time.Since(start))

	var runID config.RunID
	runID = *lc.RunID
	return &DumpConfig{
		RunID:       &runID,
		ClusterName: clusterName,
		ResPath:     filePath,
	}, nil
}

func runLocalDump(ctx context.Context, lc *LaunchConfig) (string, error) {
	log.I.Info("Loading Kubernetes data collector client")
	collect, err := collector.ClientFactory(ctx, lc.Cfg)
	if err != nil {
		return "", fmt.Errorf("collector client creation: %w", err)
	}
	defer collect.Close(ctx)
	log.I.Infof("Loaded %s collector client", collect.Name())

	// Create the dumper instance
	clusterName, _ := collector.GetClusterName(ctx)
	log.I.Infof("Dumping %s to %s", clusterName, viper.GetString(config.CollectorLocalOutputDir))
	dumpIngestor, err := dump.NewDumpIngestor(ctx, collect, viper.GetBool(config.CollectorLocalCompress), viper.GetString(config.CollectorLocalOutputDir), lc.RunID)
	if err != nil {
		return "", fmt.Errorf("create dumper: %w", err)
	}
	defer dumpIngestor.Close(ctx)

	// Dumping all k8s objects using the API
	err = dumpIngestor.DumpK8sObjects(ctx)
	if err != nil {
		return "", fmt.Errorf("dump k8s object: %w", err)
	}

	return dumpIngestor.OutputPath(), nil
}
