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
	inputFilePath string
)

var (
	ingestCmd = &cobra.Command{
		Use:   "ingest",
		Short: "Start an ingestion locally or remotely",
		Long:  `Run an ingestion locally (local) or on KHaaS from a bucket to build the attack path (remote)`,
	}

	localIngestCmd = &cobra.Command{
		Use:   "local",
		Short: "Ingest data locally from a KubeHound dump",
		Long:  `Run an ingestion locally using a previous dump (directory or tar.gz)`,
		PreRunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.InitializeKubehoundConfig(cobraCmd.Context(), "", true, true)
		},
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			// Passing the Kubehound config from viper
			khCfg, err := cmd.GetConfig()
			if err != nil {
				return fmt.Errorf("get config: %w", err)
			}

			return core.CoreLocalIngest(cobraCmd.Context(), khCfg, inputFilePath)
		},
	}

	remoteIngestCmd = &cobra.Command{
		Use:   "remote",
		Short: "Ingest data remotely on a KHaaS instance",
		Long:  `Run an ingestion on KHaaS from a bucket to build the attack path`,
		PreRunE: func(cobraCmd *cobra.Command, args []string) error {
			viper.BindPFlag(config.IngestorAPIEndpoint, cobraCmd.Flags().Lookup("khaas-server")) //nolint: errcheck
			viper.BindPFlag(config.IngestorAPIInsecure, cobraCmd.Flags().Lookup("insecure"))     //nolint: errcheck

			return cmd.InitializeKubehoundConfig(cobraCmd.Context(), "", false, true)
		},
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			// Passing the Kubehound config from viper
			khCfg, err := cmd.GetConfig()
			if err != nil {
				return fmt.Errorf("get config: %w", err)
			}

			return core.CoreClientGRPCIngest(khCfg.Ingestor, khCfg.Ingestor.ClusterName, khCfg.Ingestor.RunID)
		},
	}
)

func init() {

	ingestCmd.AddCommand(localIngestCmd)
	cmd.InitLocalIngestCmd(localIngestCmd)
	localIngestCmd.Flags().StringVar(&inputFilePath, "data", "", "Filepath for the data to process (directory or tar.gz path)")

	ingestCmd.AddCommand(remoteIngestCmd)
	cmd.InitRemoteIngestCmd(remoteIngestCmd, true)

	rootCmd.AddCommand(ingestCmd)
}
