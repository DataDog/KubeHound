package main

import (
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "print Kubehound version",
		Long:  `print the current version of Kubehound`,
		Run: func(cobraCmd *cobra.Command, args []string) {
			fmt.Printf("kubehound version: %s (%s/%s)", config.BuildVersion, config.BuildArch, config.BuildOs) // ignore:forbidigo
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
