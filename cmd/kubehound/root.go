package main

import (
	"fmt"

	"github.com/DataDog/KubeHound/pkg/backend"
	"github.com/DataDog/KubeHound/pkg/cmd"
	"github.com/DataDog/KubeHound/pkg/kubehound/core"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/spf13/cobra"
)

var (
	cfgFile     = ""
	skipBackend = false
)

var (
	rootCmd = &cobra.Command{
		Use:   "kubehound",
		Short: "A local Kubehound instance",
		Long:  `A local instance of Kubehound - a Kubernetes attack path generator`,
		PreRunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.InitializeKubehoundConfig(cobraCmd.Context(), cfgFile, true, false)
		},
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			// auto spawning the backend stack
			if !skipBackend {
				// Forcing the embed docker config to be loaded
				Backend, err := backend.NewBackend(cobraCmd.Context(), []string{""})
				if err != nil {
					return err
				}
				res, err := Backend.IsStackRunning(cobraCmd.Context())
				if err != nil {
					return err
				}
				if !res {
					err = Backend.Up(cobraCmd.Context())
					if err != nil {
						return err
					}
				} else {
					log.I.Info("Backend stack is already running")
				}
			}

			// Passing the Kubehound config from viper
			khCfg, err := cmd.GetConfig()
			if err != nil {
				return fmt.Errorf("get config: %w", err)
			}

			return core.CoreLive(cobraCmd.Context(), khCfg)
		},
		PersistentPostRunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.CloseKubehoundConfig()
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", cfgFile, "application config file")

	rootCmd.Flags().BoolVar(&skipBackend, "skip-backend", skipBackend, "skip the auto deployment of the backend stack (janusgraph, mongodb, and UI)")

	cmd.InitRootCmd(rootCmd)
}
