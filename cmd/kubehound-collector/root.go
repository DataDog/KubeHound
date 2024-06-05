package main

import (
	"github.com/DataDog/KubeHound/pkg/cmd"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "kubehound-collector",
		Short: "Kubehound collector CLI to collect data from a Kubernetes cluster",
		Long:  `Kubehound collector CLI to collect data and push it to KHaaS through an s3.`,
		PersistentPostRunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.CloseKubehoundConfig()
		},
	}
)

func init() {
	cmd.InitRootCmd(rootCmd)
}
