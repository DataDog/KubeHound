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

	cloudCmd = &cobra.Command{
		Use:   "cloud",
		Short: "Push collected k8s resources to an s3 bucket of a targeted cluster",
		Long:  `Collect all Kubernetes resources needed to build the attack path in an offline format on a s3 bucket`,
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

			return err
		},
	}

	localCmd = &cobra.Command{
		Use:   "local",
		Short: "Dump locally the k8s resources of a targeted cluster",
		Long:  `Collect all Kubernetes resources needed to build the attack path in an offline format locally (compressed or flat)`,
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
	cmd.InitLocalCmd(localCmd)
	cmd.InitCloudCmd(cloudCmd)

	dumpCmd.AddCommand(cloudCmd)
	dumpCmd.AddCommand(localCmd)
	rootCmd.AddCommand(dumpCmd)
}
