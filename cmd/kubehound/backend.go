package main

import (
	docker "github.com/DataDog/KubeHound/pkg/backend"
	"github.com/spf13/cobra"
)

var (
	Backend *docker.Backend
	hard    bool
)

var (
	backendCmd = &cobra.Command{
		Use:   "backend",
		Short: "Handle the kubehound stack",
		Long:  `Handle the kubehound stack - docker compose based stack for kubehound services (mongodb, graphdb and UI)`,
		PersistentPreRunE: func(cobraCmd *cobra.Command, args []string) error {
			var err error
			Backend, err = docker.NewBackend(cobraCmd.Context())

			return err
		},
	}

	backendUpCmd = &cobra.Command{
		Use:   "up",
		Short: "Spawn the kubehound stack",
		Long:  `Spawn the kubehound stack`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return Backend.Up(cobraCmd.Context())
		},
	}

	backendWipeCmd = &cobra.Command{
		Use:   "wipe",
		Short: "Wipe the persisted backend data",
		Long:  `Wipe the persisted backend data`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return Backend.Wipe(cobraCmd.Context())
		},
	}

	backendDownCmd = &cobra.Command{
		Use:   "down",
		Short: "Tear down the kubehound stack",
		Long:  `Tear down the kubehound stack`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return Backend.Down(cobraCmd.Context())
		},
	}

	backendResetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Restart the kubehound stack",
		Long:  `Restart the kubehound stack`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			err := Backend.Down(cobraCmd.Context())
			if err != nil {
				return err
			}

			return Backend.Reset(cobraCmd.Context())
		},
	}
)

func init() {
	backendCmd.AddCommand(backendUpCmd)
	backendCmd.AddCommand(backendWipeCmd)
	backendCmd.AddCommand(backendResetCmd)
	backendResetCmd.Flags().BoolVar(&skipBackend, "hard", false, "Also wipe all data before restarting the stack")

	backendCmd.AddCommand(backendDownCmd)

	rootCmd.AddCommand(backendCmd)
}
