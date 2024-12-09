package main

import (
	"fmt"

	"github.com/DataDog/KubeHound/pkg/cmd"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	runID string
)

var (
	ingestCmd = &cobra.Command{
		Use:   "ingest",
		Short: "Start an ingestion locally or remotely",
		Long:  `Run an ingestion locally (local) or on KHaaS from a bucket to build the attack path (remote)`,
	}

	localIngestCmd = &cobra.Command{
		Use:   "local [directory or tar.gz path]",
		Short: "Ingest data locally from a KubeHound dump",
		Long:  `Run an ingestion locally using a previous dump (directory or tar.gz)`,
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cobraCmd *cobra.Command, args []string) error {
			cmd.BindFlagCluster(cobraCmd)

			return cmd.InitializeKubehoundConfig(cobraCmd.Context(), cfgFile, true, true)
		},
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			// Passing the Kubehound config from viper
			khCfg, err := cmd.GetConfig()
			if err != nil {
				return fmt.Errorf("get config: %w", err)
			}

			return core.CoreLocalIngest(cobraCmd.Context(), khCfg, args[0])
		},
	}

	remoteIngestCmd = &cobra.Command{
		Use:   "remote",
		Short: "Ingest data remotely on a KHaaS instance",
		Long:  `Run an ingestion on KHaaS from a bucket to build the attack path, by default it will rehydrate the latest snapshot previously dumped on a KHaaS instance from all clusters`,
		PreRunE: func(cobraCmd *cobra.Command, args []string) error {
			cmd.BindFlagCluster(cobraCmd)
			viper.BindPFlag(config.IngestorAPIEndpoint, cobraCmd.Flags().Lookup("khaas-server")) //nolint: errcheck
			viper.BindPFlag(config.IngestorAPIInsecure, cobraCmd.Flags().Lookup("insecure"))     //nolint: errcheck

			if !isIngestRemoteDefault() {
				cobraCmd.MarkFlagRequired("run_id")  //nolint: errcheck
				cobraCmd.MarkFlagRequired("cluster") //nolint: errcheck
			}

			return cmd.InitializeKubehoundConfig(cobraCmd.Context(), cfgFile, false, true)
		},
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			// Passing the Kubehound config from viper
			khCfg, err := cmd.GetConfig()
			if err != nil {
				return fmt.Errorf("get config: %w", err)
			}

			if isIngestRemoteDefault() {
				return core.CoreClientGRPCRehydrateLatest(cobraCmd.Context(), khCfg.Ingestor)
			}

			return core.CoreClientGRPCIngest(cobraCmd.Context(), khCfg.Ingestor, khCfg.Dynamic.ClusterName, runID)
		},
	}
)

// If no arg is provided, run the reHydration of the latest snapshots (stored in KHaaS / S3 Bucket)
func isIngestRemoteDefault() bool {
	clusterName := viper.GetString(config.DynamicClusterName)

	return runID == "" && clusterName == ""
}

func init() {

	ingestCmd.AddCommand(localIngestCmd)
	cmd.InitLocalIngestCmd(localIngestCmd)

	ingestCmd.AddCommand(remoteIngestCmd)
	cmd.InitRemoteIngestCmd(remoteIngestCmd, true)
	remoteIngestCmd.Flags().StringVar(&runID, "run_id", "", "KubeHound run id to ingest (e.g.: 01htdgjj34mcmrrksw4bjy2e94)")

	rootCmd.AddCommand(ingestCmd)
}
