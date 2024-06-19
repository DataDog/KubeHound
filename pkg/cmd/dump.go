package cmd

import (
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func InitRootCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("debug", false, "Enable debug logs")
	viper.BindPFlag(config.GlobalDebug, cmd.PersistentFlags().Lookup("debug")) //nolint: errcheck
}

func InitDumpCmd(cmd *cobra.Command) {
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

	cmd.PersistentFlags().BoolP("non-interactive", "y", config.DefaultK8sAPINonInteractive, "Non interactive mode (skip cluster confirmation)")
	viper.BindPFlag(config.CollectorNonInteractive, cmd.PersistentFlags().Lookup("non-interactive")) //nolint: errcheck

	cmd.PersistentFlags().Bool("debug", false, "Enable debug logs")
	viper.BindPFlag(config.GlobalDebug, cmd.PersistentFlags().Lookup("debug")) //nolint: errcheck
}

func InitLocalDumpCmd(cmd *cobra.Command) {
	cmd.Flags().Bool("compress", false, "Enable compression for the dumped data (generates a tar.gz file)")
	viper.BindPFlag(config.CollectorFileArchiveFormat, cmd.Flags().Lookup("compress")) //nolint: errcheck

	cmd.Flags().String("output-dir", "", "Directory to dump the data")
	viper.BindPFlag(config.CollectorFileDirectory, cmd.Flags().Lookup("output-dir")) //nolint: errcheck
	cmd.MarkFlagRequired("output-dir")                                               //nolint: errcheck
}

func InitRemoteDumpCmd(cmd *cobra.Command) {
	cmd.Flags().String("bucket", "", "Bucket to use to push k8s resources (e.g.: s3://<your_bucket>)")
	viper.BindPFlag(config.CollectorFileBlobBucket, cmd.Flags().Lookup("bucket")) //nolint: errcheck
	cmd.MarkFlagRequired("bucket")                                                //nolint: errcheck

	cmd.Flags().String("region", "", "Region to retrieve the configuration (only for s3) (e.g.: us-east-1)")
	viper.BindPFlag(config.CollectorFileBlobRegion, cmd.Flags().Lookup("region")) //nolint: errcheck
}

func InitRemoteIngestCmd(cmd *cobra.Command, standalone bool) {

	cmd.Flags().String("khaas-server", "", "GRPC endpoint exposed by KubeHound as a Service (KHaaS) server (e.g.: localhost:9000)")
	cmd.Flags().Bool("insecure", config.DefaultIngestorAPIInsecure, "Allow insecure connection to the KHaaS server grpc endpoint")

	// IngestorAPIEndpoint
	if standalone {
		cmd.Flags().String("run_id", "", "KubeHound run id to ingest (e.g.: 01htdgjj34mcmrrksw4bjy2e94)")
		viper.BindPFlag(config.IngestorRunID, cmd.Flags().Lookup("run_id")) //nolint: errcheck
		cmd.MarkFlagRequired("run_id")                                      //nolint: errcheck

		cmd.Flags().String("cluster", "", "Cluster name to ingest (e.g.: my-cluster-1)")
		viper.BindPFlag(config.IngestorClusterName, cmd.Flags().Lookup("cluster")) //nolint: errcheck
		cmd.MarkFlagRequired("cluster")                                            //nolint: errcheck

		// Reusing the same flags for the dump cloud and ingest command
		cmd.MarkFlagRequired("khaas-server") //nolint: errcheck
	}
}

func InitLocalIngestCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().String("cluster", "", "Cluster name to ingest (e.g.: my-cluster-1)")
	viper.BindPFlag(config.IngestorClusterName, cmd.Flags().Lookup("cluster")) //nolint: errcheck
	cmd.MarkFlagRequired("cluster")                                            //nolint: errcheck
}
