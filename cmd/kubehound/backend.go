//go:build no_backend

package main

import (
	"github.com/DataDog/KubeHound/pkg/backend"
	"github.com/spf13/cobra"
)

var (
	Backend     *backend.Backend
	hard        bool
	composePath []string

	uiProfile = backend.DefaultUIProfile
	uiInvana  bool
)

var (
	backendCmd = &cobra.Command{
		Use:   "backend",
		Short: "Handle the kubehound stack",
		Long:  `Handle the kubehound stack - docker compose based stack for kubehound services (mongodb, graphdb and UI)`,
		PersistentPreRunE: func(cobraCmd *cobra.Command, args []string) error {
			if uiInvana {
				uiProfile = append(uiProfile, "invana")
			}

			return backend.NewBackend(cobraCmd.Context(), composePath, uiProfile)
		},
	}

	backendUpCmd = &cobra.Command{
		Use:   "up",
		Short: "Spawn the kubehound stack",
		Long:  `Spawn the kubehound stack`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return backend.Up(cobraCmd.Context())
		},
	}

	backendWipeCmd = &cobra.Command{
		Use:   "wipe",
		Short: "Wipe the persisted backend data",
		Long:  `Wipe the persisted backend data`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return backend.Wipe(cobraCmd.Context())
		},
	}

	backendDownCmd = &cobra.Command{
		Use:   "down",
		Short: "Tear down the kubehound stack",
		Long:  `Tear down the kubehound stack`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return backend.Down(cobraCmd.Context())
		},
	}

	backendResetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Restart the kubehound stack",
		Long:  `Restart the kubehound stack`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			err := backend.Down(cobraCmd.Context())
			if err != nil {
				return err
			}

			if hard {
				err = backend.Wipe(cobraCmd.Context())
				if err != nil {
					return err
				}
			}

			return backend.Reset(cobraCmd.Context())
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
	backendCmd.PersistentFlags().BoolVar(&uiInvana, "invana", false, "Activate Invana front end as KubeHound UI alternative")
	rootCmd.AddCommand(backendCmd)
}
