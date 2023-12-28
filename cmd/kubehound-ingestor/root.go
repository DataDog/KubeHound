package main

import (
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return core.LaunchRemoteIngestor(cmd.Context(), core.WithConfigPath(cfgFile))
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", cfgFile, "application config file")
}
