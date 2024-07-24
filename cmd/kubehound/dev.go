package main

import (
	"context"
	"os"

	"github.com/DataDog/KubeHound/pkg/backend"
	docker "github.com/DataDog/KubeHound/pkg/backend"
	"github.com/spf13/cobra"
)

var (
	DefaultComposeTestingPath = []string{"./deployments/kubehound/docker-compose.yaml", "./deployments/kubehound/docker-compose.testing.yaml"}
	DefaultComposeDevPath     = []string{"./deployments/kubehound/docker-compose.yaml", "./deployments/kubehound/docker-compose.dev.graph.yaml", "./deployments/kubehound/docker-compose.dev.mongo.yaml"}
	DefaultComposeDevPathUI   = "./deployments/kubehound/docker-compose.dev.ui.yaml"
	DefaultComposeDevPathGRPC = "./deployments/kubehound/docker-compose.dev.ingestor.yaml"
	DefaultDatadogComposePath = "./deployments/kubehound/docker-compose.datadog.yaml"
)

var (
	uiTesting   bool
	grpcTesting bool
	downTesting bool
	profiles    []string
)

var (
	envCmd = &cobra.Command{
		Use:    "dev",
		Hidden: true,
		Short:  "[devOnly] Spawn the kubehound testing stack",
		Long:   `[devOnly] Spawn the kubehound dev stack for the system-tests (build from dockerfile)`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if uiTesting {
				DefaultComposeDevPath = append(DefaultComposeDevPath, DefaultComposeDevPathUI)
			}
			if grpcTesting {
				DefaultComposeDevPath = append(DefaultComposeDevPath, DefaultComposeDevPathGRPC)
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
	if uiTesting {
		profiles = append(profiles, backend.DevUIProfile)
	}

	err := docker.NewBackend(ctx, composePaths, profiles)
	if err != nil {
		return err
	}
	if downTesting {
		return docker.Down(ctx)
	}

	return docker.BuildUp(ctx)
}

func init() {
	envCmd.AddCommand(envTestingCmd)
	envCmd.PersistentFlags().BoolVar(&downTesting, "down", false, "Tearing down the kubehound dev stack and deleting the data associated with it")
	envCmd.Flags().BoolVar(&uiTesting, "ui", false, "Include the UI in the dev stack")
	envCmd.Flags().BoolVar(&grpcTesting, "grpc", false, "Include Grpc Server (ingestor) in the dev stack")

	rootCmd.AddCommand(envCmd)
}
