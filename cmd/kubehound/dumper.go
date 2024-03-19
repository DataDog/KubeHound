package main

import (
	"context"
	"fmt"
	"os"

	"github.com/DataDog/KubeHound/pkg/cmd"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/core"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var debug bool

var (
	dumpCmd = &cobra.Command{
		Use:    "dump",
		Short:  "Collect Kubernetes resources of a targeted cluster",
		Long:   `Collect all Kubernetes resources needed to build the attack path. This will be dumped in an offline format (s3 or locally)`,
		PreRun: toggleDebug,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	s3Cmd = &cobra.Command{
		Use:    "s3",
		Short:  "Push collected k8s resources to an s3 bucket of a targeted cluster",
		Long:   `Collect all Kubernetes resources needed to build the attack path in an offline format on a s3 bucket`,
		PreRun: toggleDebug,
		RunE: func(cmd *cobra.Command, args []string) error {
			// using compress feature
			viper.Set(config.CollectorLocalCompress, true)

			// Create a temporary directory
			tmpDir, err := os.MkdirTemp("", "kubehound")
			if err != nil {
				return fmt.Errorf("create temporary directory: %w", err)
			}

			log.I.Debugf("Temporary directory created: %s", tmpDir)
			viper.Set(config.CollectorLocalOutputDir, tmpDir)

			_, err = core.DumpCore(context.Background(), cmd)

			return err
		},
	}

	localCmd = &cobra.Command{
		Use:    "local",
		Short:  "Dump locally the k8s resources of a targeted cluster",
		Long:   `Collect all Kubernetes resources needed to build the attack path in an offline format locally (compressed or flat)`,
		PreRun: toggleDebug,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := core.DumpCore(context.Background(), cmd)

			return err
		},
	}
)

func init() {

	cmd.InitDumpCmd(dumpCmd)
	cmd.InitLocalCmd(localCmd)
	cmd.InitS3Cmd(s3Cmd)

	dumpCmd.AddCommand(s3Cmd)
	dumpCmd.AddCommand(localCmd)
	rootCmd.AddCommand(dumpCmd)
}

func toggleDebug(cmd *cobra.Command, args []string) {
	if debug {
		log.I.Logger.SetLevel(logrus.DebugLevel)
	}
}
