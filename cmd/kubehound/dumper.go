package main

import (
	"context"
	"fmt"
	"os"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/core"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
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
			return cmd.Help()
		},
	}

	s3Cmd = &cobra.Command{
		Use:    "s3",
		Short:  "Push collected k8s resources to an s3 bucket of a targeted cluster",
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

			_, err = core.DumpCore(context.Background(), cmd)

			return err
		},
	}

	localCmd = &cobra.Command{
		Use:    "local",
		Short:  "Dump locally the k8s resources of a targeted cluster",
		Long:   `Collect all Kubernetes resources needed to build the attack path in an offline format locally (compressed or flat)`,
		PreRun: toggleDebug,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := core.DumpCore(context.Background(), cmd)

			return err
		},
	}
)

func initDumpCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().String("statsd", config.DefaultTelemetryStatsdUrl, "URL of the statsd endpoint")
	viper.BindPFlag(config.TelemetryStatsdUrl, cmd.PersistentFlags().Lookup("statsd")) //nolint: errcheck

	cmd.PersistentFlags().String("profiler", config.DefaultTelemetryProfilerUrl, "URL of the profiler endpoint")
	viper.BindPFlag(config.TelemetryTracerUrl, cmd.PersistentFlags().Lookup("profiler")) //nolint: errcheck

	cmd.PersistentFlags().Bool("telemetry", false, "Enable telemetry with default settings")
	viper.BindPFlag(config.TelemetryEnabled, cmd.PersistentFlags().Lookup("telemetry")) //nolint: errcheck

	cmd.PersistentFlags().Duration("period", config.DefaultProfilerPeriod, "Period specifies the interval at which to collect profiles")
	viper.BindPFlag(config.TelemetryProfilerPeriod, cmd.PersistentFlags().Lookup("period")) //nolint: errcheck

	cmd.PersistentFlags().Duration("cpu-duration", config.DefaultProfilerCPUDuration, "CPU Duration specifies the length at which to collect CPU profiles")
	viper.BindPFlag(config.TelemetryProfilerCPUDuration, cmd.PersistentFlags().Lookup("cpu-duration")) //nolint: errcheck

	cmd.PersistentFlags().Int("rate", config.DefaultK8sAPIRateLimitPerSecond, "Rate limit of requests/second to the Kubernetes API")
	viper.BindPFlag(config.CollectorLiveRate, cmd.PersistentFlags().Lookup("rate")) //nolint: errcheck

	cmd.PersistentFlags().Int64("page-size", config.DefaultK8sAPIPageSize, "Number of entries retrieved by each call on the API (same for all Kubernetes entry types)")
	viper.BindPFlag(config.CollectorLivePageSize, cmd.PersistentFlags().Lookup("page-size")) //nolint: errcheck

	cmd.PersistentFlags().Int32("page-buffer-count", config.DefaultK8sAPIPageBufferSize, "Number of pages to buffer")
	viper.BindPFlag(config.CollectorLivePageBufferSize, cmd.PersistentFlags().Lookup("page-buffer-count")) //nolint: errcheck

	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logs")
}

func initLocalCmd(cmd *cobra.Command) {
	cmd.Flags().Bool("compress", false, "Enable compression for the dumped data (generates a tar.gz file)")
	viper.BindPFlag(config.CollectorLocalCompress, cmd.Flags().Lookup("compress")) //nolint: errcheck

	cmd.Flags().String("output-dir", "", "Directory to dump the data")
	viper.BindPFlag(config.CollectorLocalOutputDir, cmd.Flags().Lookup("output-dir")) //nolint: errcheck
	cmd.MarkFlagRequired("output-dir")
}

func initS3Cmd(cmd *cobra.Command) {
	cmd.Flags().String("bucket", "", "Bucket to use to push k8s resources (e.g.: s3://kubehound-dumps)")
	viper.BindPFlag(config.CollectorS3Bucket, cmd.Flags().Lookup("bucket")) //nolint: errcheck
	cmd.MarkFlagRequired("bucket")                                          //nolint: errcheck

	cmd.Flags().String("region", "", "Region to use to push k8s resources")
	viper.BindPFlag(config.CollectorS3Region, cmd.Flags().Lookup("region")) //nolint: errcheck
	cmd.MarkFlagRequired("region")                                          //nolint: errcheck
}

func init() {

	initDumpCmd(dumpCmd)

	initLocalCmd(localCmd)

	initS3Cmd(s3Cmd)

	dumpCmd.AddCommand(s3Cmd)
	dumpCmd.AddCommand(localCmd)
	rootCmd.AddCommand(dumpCmd)
}

func toggleDebug(cmd *cobra.Command, args []string) {
	if debug {
		log.I.Logger.SetLevel(logrus.DebugLevel)
	}
}
