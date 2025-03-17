package main

import (
	"fmt"
	"os"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "print Kubehound version",
		Long:  `print the current version of Kubehound`,
		Run: func(cobraCmd *cobra.Command, args []string) {
			fmt.Printf("kubehound version: %s (%s/%s)\n", config.BuildVersion, config.BuildArch, config.BuildOs) //nolint:forbidigo
		},
		PersistentPostRun: func(cobraCmd *cobra.Command, args []string) {
			os.Exit(0)
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
