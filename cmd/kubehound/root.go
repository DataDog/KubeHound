package main

import (
	"context"

	"github.com/DataDog/Kubehound/pkg/kubehound/core"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "kubehound-local",
		Short: "A local Kubehound instance",
		Long:  `A local instance of Kubehound - a Kubernetes attack path generator`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return core.Launch(context.Background())
		},
	}
)
