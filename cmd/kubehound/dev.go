package main

import (
	"context"
	"os"

	docker "github.com/DataDog/KubeHound/pkg/backend"
	"github.com/spf13/cobra"
)

var (
	DefaultComposeTestingPath = []string{"./deployments/kubehound/docker-compose.yaml", "./deployments/kubehound/docker-compose.testing.yaml"}
	DefaultComposeDevPath     = []string{"./deployments/kubehound/docker-compose.yaml", "./deployments/kubehound/docker-compose.dev.yaml"}
	DefaultComposeDevPathUI   = "./deployments/kubehound/docker-compose.ui.yaml"
	DefaultDatadogComposePath = "./deployments/kubehound/docker-compose.datadog.yaml"
)

var (
	envCmd = &cobra.Command{
		Use:    "dev",
		Hidden: true,
		Short:  "[devOnly] Spawn the kubehound testing stack",
		Long:   `[devOnly] Spawn the kubehound dev stack for the system-tests (build from dockerfile)`,
		PersistentPreRunE: func(cobraCmd *cobra.Command, args []string) error {
			var err error
			Backend, err = docker.NewBackend(cobraCmd.Context(), composePath)

			return err
		},
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if uiTesting {
				DefaultComposeDevPath = append(DefaultComposeDevPath, DefaultComposeDevPathUI)
			}
			// Adding datadog setup
			_, ddAPIKeyOk := os.LookupEnv("DD_API_KEY")
			_, ddAPPKeyOk := os.LookupEnv("DD_API_KEY")
			if ddAPIKeyOk && ddAPPKeyOk {
				DefaultComposeDevPath = append(DefaultComposeDevPath, DefaultDatadogComposePath)
			}

			return runEnv(cobraCmd.Context(), DefaultComposeDevPath)
		},
	}

	envTestingCmd = &cobra.Command{
		Use:   "system-tests",
		Short: "[devOnly] Spawn the kubehound testing stack for the system-tests",
		Long:  `[devOnly] Spawn the kubehound testing stack for the system-tests (using +1 port for the services)`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return runEnv(cobraCmd.Context(), DefaultComposeTestingPath)
		},
	}
)

func runEnv(ctx context.Context, composePaths []string) error {
	Backend, err := docker.NewBackend(ctx, composePaths)
	if err != nil {
		return err
	}
	if downTesting {
		return Backend.Down(ctx)
	}

	return Backend.BuildUp(ctx)
}

func init() {
	envCmd.AddCommand(envTestingCmd)
	envCmd.PersistentFlags().BoolVar(&downTesting, "down", false, "Tearing down the kubehound dev stack and deleting the data associated with it")
	envCmd.Flags().BoolVar(&uiTesting, "ui", false, "Include the UI in the dev stack")

	rootCmd.AddCommand(envCmd)
}
