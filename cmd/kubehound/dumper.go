package main

import (
	"fmt"
	"os"

	"github.com/DataDog/KubeHound/pkg/cmd"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/core"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Collect Kubernetes resources of a targeted cluster",
		Long:  `Collect all Kubernetes resources needed to build the attack path. This will be dumped in an offline format (s3 or locally)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	dumpRemoteCmd = &cobra.Command{
		Use:   "remote",
		Short: "Push collected k8s resources to an s3 bucket of a targeted cluster. Run the ingestion on KHaaS if khaas-server is set.",
		Long:  `Collect all Kubernetes resources needed to build the attack path in an offline format on a s3 bucket. If KubeHound as a Service (KHaaS) is set, then run the ingestion on KHaaS.`,
		PreRunE: func(cobraCmd *cobra.Command, args []string) error {
			viper.BindPFlag(config.IngestorAPIEndpoint, cobraCmd.Flags().Lookup("khaas-server")) //nolint: errcheck
			viper.BindPFlag(config.IngestorAPIInsecure, cobraCmd.Flags().Lookup("insecure"))     //nolint: errcheck

			return cmd.InitializeKubehoundConfig(cobraCmd.Context(), "", true, true)
		},
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			// using compress feature
			viper.Set(config.CollectorFileArchiveFormat, true)

			// Create a temporary directory
			tmpDir, err := os.MkdirTemp("", "kubehound")
			if err != nil {
				return fmt.Errorf("create temporary directory: %w", err)
			}

			log.I.Debugf("Temporary directory created: %s", tmpDir)
			viper.Set(config.CollectorFileDirectory, tmpDir)

			// Passing the Kubehound config from viper
			khCfg, err := cmd.GetConfig()
			if err != nil {
				return fmt.Errorf("get config: %w", err)
			}

			_, err = core.DumpCore(cobraCmd.Context(), khCfg, true)
			if err != nil {
				return fmt.Errorf("dump core: %w", err)
			}
			// Running the ingestion on KHaaS
			if cobraCmd.Flags().Lookup("khaas-server").Value.String() != "" {
				return core.CoreClientGRPCIngest(khCfg.Ingestor, khCfg.Dynamic.ClusterName, khCfg.Dynamic.RunID.String())
			}

			return err
		},
	}
	dumpLocalCmd = &cobra.Command{
		Use:   "local",
		Short: "Dump locally the k8s resources of a targeted cluster",
		Long:  `Collect all Kubernetes resources needed to build the attack path in an offline format locally (compressed or flat)`,
		PreRunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.InitializeKubehoundConfig(cobraCmd.Context(), "", true, true)
		},
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			// Passing the Kubehound config from viper
			khCfg, err := cmd.GetConfig()
			if err != nil {
				return fmt.Errorf("get config: %w", err)
			}
			_, err = core.DumpCore(cobraCmd.Context(), khCfg, false)

			return err
		},
	}
)

func init() {
	cmd.InitDumpCmd(dumpCmd)
	cmd.InitLocalDumpCmd(dumpLocalCmd)
	cmd.InitRemoteDumpCmd(dumpRemoteCmd)
	cmd.InitRemoteIngestCmd(dumpRemoteCmd, false)

	dumpCmd.AddCommand(dumpRemoteCmd)
	dumpCmd.AddCommand(dumpLocalCmd)
	rootCmd.AddCommand(dumpCmd)
}
