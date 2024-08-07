package main

import (
	"fmt"
	"os"

	"github.com/DataDog/KubeHound/pkg/cmd"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	configPath string
)

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Show the current configuration",
		Long:  `[devOnly] Spawn the kubehound dev stack for the system-tests (build from dockerfile)`,
		PreRunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.InitializeKubehoundConfig(cobraCmd.Context(), "", true, true)
		},
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			// Adding datadog setup
			khCfg, err := cmd.GetConfig()
			if err != nil {
				return fmt.Errorf("get config: %w", err)
			}

			yamlData, err := yaml.Marshal(&khCfg)

			if err != nil {
				return fmt.Errorf("marshaling khCfg: %v", err)
			}

			if configPath != "" {
				f, err := os.Create(configPath)
				if err != nil {
					return fmt.Errorf("creating file: %v", err)
				}

				_, err = f.Write(yamlData)
				if err != nil {
					return fmt.Errorf("writing to file: %v", err)
				}

				log.I.Infof("Configuration saved to %s\n", configPath)

				return nil
			}

			fmt.Println(" --- YAML ---")  //nolint:forbidigo
			fmt.Println(string(yamlData)) //nolint:forbidigo

			return nil
		},
	}
)

func init() {
	configCmd.Flags().StringVar(&configPath, "path", "", "path to dump current KubeHound configuration")

	rootCmd.AddCommand(configCmd)
}
