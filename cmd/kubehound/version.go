package main

import (
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "print Kubehound version",
		Long:  `print the current version of Kubehound`,
		Run: func(cobraCmd *cobra.Command, args []string) {
			log.I.Infof("kubehound version: %s (%s/%s)", config.BuildVersion, config.BuildArch, config.BuildOs)
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
