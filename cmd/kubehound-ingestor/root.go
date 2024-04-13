package main

import (
	"fmt"

	"github.com/DataDog/KubeHound/pkg/cmd"
	"github.com/DataDog/KubeHound/pkg/kubehound/core"
	"github.com/spf13/cobra"
)

var (
	cfgFile = ""
)

var (
	rootCmd = &cobra.Command{
		Use:          "kubehound-ingestor",
		Short:        "Kubehound Ingestor Service",
		Long:         `instance of Kubehound that pulls data from cloud storage`,
		SilenceUsage: true,
		PersistentPreRunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.InitializeKubehoundConfig(cobraCmd.Context(), cfgFile, true, false)
		},
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			// Passing the Kubehound config from viper
			khCfg, err := cmd.GetConfig()
			if err != nil {
				return fmt.Errorf("get config: %w", err)
			}

			return core.CoreGrpcApi(cobraCmd.Context(), khCfg)
		},
		PersistentPostRunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.CloseKubehoundConfig()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", cfgFile, "application config file")

	cmd.InitRootCmd(rootCmd)
}
