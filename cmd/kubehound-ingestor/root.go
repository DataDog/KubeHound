package main

import (
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/ingestor"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
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
			var cfg *config.KubehoundConfig
			if len(cfgFile) != 0 {
				log.I.Infof("Loading application configuration from file %s", cfgFile)
				cfg = config.MustLoadConfig(cfgFile)
			} else {
				log.I.Infof("Loading application configuration from default embedded")
				cfg = config.MustLoadEmbedConfig()
			}
			err := ingestor.Launch(cmd.Context(), cfg)
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", cfgFile, "application config file")
}
