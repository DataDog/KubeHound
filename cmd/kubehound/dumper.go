package main

import (
	"fmt"
	"os"

	docker "github.com/DataDog/KubeHound/pkg/backend"
	"github.com/DataDog/KubeHound/pkg/cmd"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/core"
	"github.com/DataDog/KubeHound/pkg/telemetry/events"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	runLocalIngest bool
	startBackend   bool
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
			viper.Set(config.CollectorFileArchiveNoCompress, false)

			// Create a temporary directory
			tmpDir, err := os.MkdirTemp("", "kubehound")
			defer cmd.ReportError(cobraCmd.Context(), events.DumpFinished, err)
			if err != nil {
				return fmt.Errorf("create temporary directory: %w", err)
			}
			l := log.Trace(cobraCmd.Context())
			l.Info("Temporary directory created", log.String("path", tmpDir))
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
				err = core.CoreClientGRPCIngest(cobraCmd.Context(), khCfg.Ingestor, khCfg.Dynamic.ClusterName, khCfg.Dynamic.RunID.String())

				return err
			}

			return err
		},
	}
	dumpLocalCmd = &cobra.Command{
		Use:   "local [directory to dump the data]",
		Short: "Dump locally the k8s resources of a targeted cluster",
		Args:  cobra.ExactArgs(1),
		Long:  `Collect all Kubernetes resources needed to build the attack path in an offline format locally (compressed or flat)`,
		PreRunE: func(cobraCmd *cobra.Command, args []string) error {
			viper.Set(config.CollectorFileDirectory, args[0])

			return cmd.InitializeKubehoundConfig(cobraCmd.Context(), "", true, true)
		},
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			// Passing the Kubehound config from viper
			khCfg, err := cmd.GetConfig()
			defer cmd.ReportError(cobraCmd.Context(), events.DumpFinished, err)
			if err != nil {
				return fmt.Errorf("get config: %w", err)
			}
			resultPath, err := core.DumpCore(cobraCmd.Context(), khCfg, false)
			if err != nil {
				return fmt.Errorf("dump core: %w", err)
			}

			if startBackend {
				err = docker.NewBackend(cobraCmd.Context(), composePath, docker.DefaultUIProfile)
				if err != nil {
					return fmt.Errorf("new backend: %w", err)
				}
				err = docker.Up(cobraCmd.Context())
				if err != nil {
					return fmt.Errorf("docker up: %w", err)
				}
			}

			if runLocalIngest {
				err = core.CoreLocalIngest(cobraCmd.Context(), khCfg, resultPath)
			}

			return err
		},
	}
)

func init() {
	cmd.InitDumpCmd(dumpCmd)
	cmd.InitLocalDumpCmd(dumpLocalCmd)
	cmd.InitRemoteDumpCmd(dumpRemoteCmd)
	cmd.InitRemoteIngestCmd(dumpRemoteCmd, false)

	dumpLocalCmd.Flags().BoolVar(&runLocalIngest, "ingest", false, "Run the ingestion after the dump")
	dumpLocalCmd.Flags().BoolVar(&startBackend, "backend", false, "Start the backend after the dump")

	dumpCmd.AddCommand(dumpRemoteCmd)
	dumpCmd.AddCommand(dumpLocalCmd)
	rootCmd.AddCommand(dumpCmd)
}
