package main

import (
	docker "github.com/DataDog/KubeHound/pkg/backend"
	"github.com/spf13/cobra"
)

var (
	Backend     *docker.Backend
	hard        bool
	composePath []string

	downTesting bool
	uiTesting   bool
)

var (
	backendCmd = &cobra.Command{
		Use:   "backend",
		Short: "Handle the kubehound stack",
		Long:  `Handle the kubehound stack - docker compose based stack for kubehound services (mongodb, graphdb and UI)`,
		PersistentPreRunE: func(cobraCmd *cobra.Command, args []string) error {
			return docker.NewBackend(cobraCmd.Context(), composePath)
		},
	}

	backendUpCmd = &cobra.Command{
		Use:   "up",
		Short: "Spawn the kubehound stack",
		Long:  `Spawn the kubehound stack`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return docker.Up(cobraCmd.Context())
		},
	}

	backendWipeCmd = &cobra.Command{
		Use:   "wipe",
		Short: "Wipe the persisted backend data",
		Long:  `Wipe the persisted backend data`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return docker.Wipe(cobraCmd.Context())
		},
	}

	backendDownCmd = &cobra.Command{
		Use:   "down",
		Short: "Tear down the kubehound stack",
		Long:  `Tear down the kubehound stack`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return docker.Down(cobraCmd.Context())
		},
	}

	backendResetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Restart the kubehound stack",
		Long:  `Restart the kubehound stack`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			err := docker.Down(cobraCmd.Context())
			if err != nil {
				return err
			}

			if hard {
				err = docker.Wipe(cobraCmd.Context())
				if err != nil {
					return err
				}
			}

			return docker.Reset(cobraCmd.Context())
		},
	}
)

func init() {
	backendCmd.AddCommand(backendUpCmd)
	backendCmd.AddCommand(backendWipeCmd)
	backendCmd.AddCommand(backendResetCmd)
	backendResetCmd.Flags().BoolVar(&skipBackend, "hard", false, "Also wipe all data before restarting the stack")

	backendCmd.AddCommand(backendDownCmd)
	backendCmd.PersistentFlags().StringSliceVarP(&composePath, "file", "f", composePath, "Compose configuration files")
	rootCmd.AddCommand(backendCmd)
}
