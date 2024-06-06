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
	grpcClientIngestCmd = &cobra.Command{
		Use:   "ingest",
		Short: "Start an ingestion on KubeHoud as a Service server",
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
	cmd.InitGrpcClientCmd(grpcClientIngestCmd, true)

	ingestorCmd.AddCommand(grpcClientIngestCmd)
}
