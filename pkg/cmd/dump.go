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
	cmd.PersistentFlags().Bool("telemetry", false, "Enable telemetry with default settings")
	viper.BindPFlag(config.TelemetryEnabled, cmd.PersistentFlags().Lookup("telemetry")) //nolint: errcheck

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
	cmd.Flags().Bool("no-compress", false, "Disable compression for the dumped data (generates a directory)")
	viper.BindPFlag(config.CollectorFileArchiveNoCompress, cmd.Flags().Lookup("no-compress")) //nolint: errcheck
}

func InitRemoteDumpCmd(cmd *cobra.Command) {
	cmd.Flags().String("bucket-url", "", "Bucket to use to push k8s resources (e.g.: s3://<your_bucket>)")
	viper.BindPFlag(config.IngestorBlobBucketURL, cmd.Flags().Lookup("bucket-url")) //nolint: errcheck

	cmd.Flags().String("region", "", "Region to retrieve the configuration (only for s3) (e.g.: us-east-1)")
	viper.BindPFlag(config.IngestorBlobBucketURL, cmd.Flags().Lookup("region")) //nolint: errcheck
}

func InitLocalIngestCmd(cmd *cobra.Command) {
	InitCluster(cmd)
	cmd.Flags().MarkDeprecated(flagCluster, "Since v1.4.1, KubeHound dump archive contains a metadata file holding the clustername") //nolint: errcheck
}

func InitRemoteIngestCmd(cmd *cobra.Command, standalone bool) {

	cmd.PersistentFlags().String("khaas-server", "", "GRPC endpoint exposed by KubeHound as a Service (KHaaS) server (e.g.: localhost:9000)")
	cmd.PersistentFlags().Bool("insecure", config.DefaultIngestorAPIInsecure, "Allow insecure connection to the KHaaS server grpc endpoint")

	// IngestorAPIEndpoint
	if standalone {
		InitCluster(cmd)
	}
}

const (
	flagCluster = "cluster"
)

func InitCluster(cmd *cobra.Command) {
	cmd.Flags().String(flagCluster, "", "Cluster name to ingest (e.g.: my-cluster-1)")
}

func BindFlagCluster(cmd *cobra.Command) {
	viper.BindPFlag(config.DynamicClusterName, cmd.Flags().Lookup(flagCluster)) //nolint: errcheck
}
