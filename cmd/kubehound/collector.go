package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/dumper"
	"github.com/DataDog/KubeHound/pkg/dumper/writer"
	"github.com/DataDog/KubeHound/pkg/kubehound/core"
	"github.com/DataDog/KubeHound/pkg/s3"
	"github.com/DataDog/KubeHound/pkg/telemetry/events"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	kstatsd "github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var debug bool

var (
	dumpCmd = &cobra.Command{
		Use:    "dump",
		Short:  "Collect Kubernetes resources of a targeted cluster",
		Long:   `Collect all Kubernetes resources needed to build the attack path. This will be dumped in an offline format (s3 or locally)`,
		PreRun: toggleDebug,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()
			return nil
		},
	}

	s3Cmd = &cobra.Command{
		Use:    "s3",
		Short:  "Push collected k8s resources to an s3 bucket of a targeted cluste",
		Long:   `Collect all Kubernetes resources needed to build the attack path in an offline format on a s3 bucket`,
		PreRun: toggleDebug,
		RunE: func(cmd *cobra.Command, args []string) error {
			// using compress feature
			viper.Set(config.CollectorLocalCompress, true)

			// Create a temporary directory
			tmpDir, err := os.MkdirTemp("", "kubehound")
			if err != nil {
				return fmt.Errorf("create temporary directory: %w", err)
			}

			log.I.Debugf("Temporary directory created: %s", tmpDir)
			viper.Set(config.CollectorLocalOutputDir, tmpDir)
			return dump(context.Background(), cmd)
		},
	}

	localCmd = &cobra.Command{
		Use:    "local",
		Short:  "Dump locally the k8s resources of a targeted cluster",
		Long:   `Collect all Kubernetes resources needed to build the attack path in an offline format locally (compressed or flat)`,
		PreRun: toggleDebug,
		RunE: func(cmd *cobra.Command, args []string) error {
			return dump(context.Background(), cmd)
		},
	}
)

func init() {
	dumpCmd.PersistentFlags().String("statsd", config.DefaultTelemetryStatsdUrl, "URL of the statsd endpoint")
	viper.BindPFlag(config.TelemetryStatsdUrl, dumpCmd.PersistentFlags().Lookup("statsd")) //nolint: errcheck

	dumpCmd.PersistentFlags().String("profiler", config.DefaultTelemetryProfilerUrl, "URL of the profiler endpoint")
	viper.BindPFlag(config.TelemetryTracerUrl, dumpCmd.PersistentFlags().Lookup("profiler")) //nolint: errcheck

	dumpCmd.PersistentFlags().Bool("telemetry", false, "Enable telemetry with default settings")
	viper.BindPFlag(config.TelemetryEnabled, dumpCmd.PersistentFlags().Lookup("telemetry")) //nolint: errcheck

	dumpCmd.PersistentFlags().Duration("period", config.DefaultProfilerPeriod, "Period specifies the interval at which to collect profiles")
	viper.BindPFlag(config.TelemetryProfilerPeriod, dumpCmd.PersistentFlags().Lookup("period")) //nolint: errcheck

	dumpCmd.PersistentFlags().Duration("cpu-duration", config.DefaultProfilerCPUDuration, "CPU Duration specifies the length at which to collect CPU profiles")
	viper.BindPFlag(config.TelemetryProfilerCPUDuration, dumpCmd.PersistentFlags().Lookup("cpu-duration")) //nolint: errcheck

	dumpCmd.PersistentFlags().Int("rate", config.DefaultK8sAPIRateLimitPerSecond, "Rate limit of requests/second to the Kubernetes API")
	viper.BindPFlag(config.CollectorLiveRate, dumpCmd.PersistentFlags().Lookup("rate")) //nolint: errcheck

	dumpCmd.PersistentFlags().Int64("page-size", config.DefaultK8sAPIPageSize, "Number of entries retrieved by each call on the API (same for all Kubernetes entry types)")
	viper.BindPFlag(config.CollectorLivePageSize, dumpCmd.PersistentFlags().Lookup("page-size")) //nolint: errcheck

	dumpCmd.PersistentFlags().Int32("page-buffer-count", config.DefaultK8sAPIPageBufferSize, "Number of pages to buffer")
	viper.BindPFlag(config.CollectorLivePageBufferSize, dumpCmd.PersistentFlags().Lookup("page-buffer-count")) //nolint: errcheck

	dumpCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logs")

	localCmd.Flags().Bool("compress", false, "Enable compression for the dumped data (generates a tar.gz file)")
	viper.BindPFlag(config.CollectorLocalCompress, localCmd.Flags().Lookup("compress")) //nolint: errcheck

	localCmd.Flags().String("output-dir", "", "Directory to dump the data")
	viper.BindPFlag(config.CollectorLocalOutputDir, localCmd.Flags().Lookup("output-dir")) //nolint: errcheck
	localCmd.MarkFlagRequired("output-dir")

	s3Cmd.Flags().String("bucket", "", "Bucket to use to push k8s resources")
	viper.BindPFlag(config.CollectorS3Bucket, s3Cmd.Flags().Lookup("bucket")) //nolint: errcheck
	s3Cmd.MarkFlagRequired("bucket")

	s3Cmd.Flags().String("region", "", "Region to use to push k8s resources")
	viper.BindPFlag(config.CollectorS3Region, s3Cmd.Flags().Lookup("region")) //nolint: errcheck
	s3Cmd.MarkFlagRequired("region")

	dumpCmd.AddCommand(s3Cmd)
	dumpCmd.AddCommand(localCmd)
	rootCmd.AddCommand(dumpCmd)
}

func toggleDebug(cmd *cobra.Command, args []string) {
	if debug {
		log.I.Logger.SetLevel(logrus.DebugLevel)
	}
}

func dump(ctx context.Context, cmd *cobra.Command) error {

	start := time.Now()

	// Configuration initialization
	cfg := config.MustLoadInlineConfig()
	// Hardcoding the collector type to K8sAPI
	cfg.Collector.Type = config.CollectorTypeK8sAPI

	lc := core.NewLaunchConfig(cfg, span.DumperLaunch)

	// Getting current clusterName, needed to set as global tag
	tags := map[string]string{}
	clusterName, err := collector.GetClusterName(ctx)
	if err == nil {
		tags[tag.CollectorCluster] = clusterName
	} else {
		log.I.Errorf("collector cluster info: %v", err)
	}

	// Initiate the telemetry and tags
	ctx, err = lc.Bootstrap(ctx, tags)
	if err != nil {
		return err
	}
	defer lc.Close()

	_ = kstatsd.Event(&statsd.Event{
		Title: fmt.Sprintf("Starting KubeHound dump for %s", clusterName),
		Text:  fmt.Sprintf("Starting KubeHound dump for %s", clusterName),
		Tags:  []string{tag.ActionType(events.DumperRun)},
	})

	// Create the collector instance
	log.I.Info("Loading Kubernetes data collector client")
	collect, err := collector.ClientFactory(ctx, cfg)
	if err != nil {
		return fmt.Errorf("collector client creation: %w", err)
	}
	defer collect.Close(ctx)
	log.I.Infof("Loaded %s collector client", collect.Name())

	dumper := dumper.NewDumper()
	defer dumper.Close(ctx)
	err = dumper.Initialize(ctx, collect, viper.GetBool(config.CollectorLocalCompress), viper.GetString(config.CollectorLocalOutputDir))
	if err != nil {
		return fmt.Errorf("initialize collector: %w", err)
	}

	// Multi-threading the dump with one worker for each types
	// The number of workers is set to 7 to have one thread per k8s object type to pull  fronm the Kubernetes API
	workerPoolSize := 7

	// Using single thread when zipping to avoid concurency issues
	if viper.GetBool(config.CollectorLocalCompress) {
		workerPoolSize = 1
	}

	// Dumping all k8s objects using the API
	err = dumper.DumpK8sObjects(ctx, workerPoolSize)
	if err != nil {
		return fmt.Errorf("dump k8s object: %w", err)
	}

	if cmd.Use == "s3" {
		// Clean up the temporary directory when done
		defer func() {
			err := os.RemoveAll(viper.GetString(config.CollectorLocalOutputDir))
			if err != nil {
				fmt.Println("Failed to remove temporary directory:", err)
			}
		}()

		objectKey := fmt.Sprintf("%s%s", dumper.ResName, writer.TarWriterExtension)
		log.I.Infof("Pushing %s to s3", objectKey)
		err = pushToS3(ctx, dumper.OutputPath(), objectKey)
		if err != nil {
			return fmt.Errorf("push %s to s3: %w", objectKey, err)
		}
	}
	_ = kstatsd.Event(&statsd.Event{
		Title: fmt.Sprintf("Finish KubeHound dump for %s", clusterName),
		Text:  fmt.Sprintf("KubeHound dump run has been completed in %s", time.Since(start)),
		Tags: []string{
			tag.ActionType(events.DumperRun),
		},
	})
	log.I.Infof("KubeHound dump run has been completed in %s", time.Since(start))

	return nil
}

// Push results to the s3 bucket
func pushToS3(ctx context.Context, filePath string, key string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	s3Client, err := s3.NewS3Store(ctx, viper.GetString(config.CollectorS3Region), viper.GetString(config.CollectorS3Bucket))
	if err != nil {
		return err
	}

	return s3Client.Upload(ctx, key, content)
}
