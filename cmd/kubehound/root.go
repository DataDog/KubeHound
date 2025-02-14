package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/DataDog/KubeHound/pkg/cmd"
	"github.com/DataDog/KubeHound/pkg/kubehound/core"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
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
		PreRunE: func(cobraCmd *cobra.Command, _ []string) error {
			return cmd.InitializeKubehoundConfig(cobraCmd.Context(), cfgFile, true, false)
		},
		RunE: func(cobraCmd *cobra.Command, _ []string) error {
			l := log.Logger(cobraCmd.Context())
			// auto spawning the backend stack
			if !skipBackend {
				err := runBackend(cobraCmd.Context())
				if err != nil {
					return err
				}
			}

			// Passing the Kubehound config from viper
			khCfg, err := cmd.GetConfig()
			if err != nil {
				return fmt.Errorf("get config: %w", err)
			}

			err = core.CoreInitLive(cobraCmd.Context(), khCfg)
			if err != nil {
				return err
			}

			err = core.CoreLive(cobraCmd.Context(), khCfg)
			if err != nil {
				return err
			}

			l.Warn("KubeHound as finished ingesting and building the graph successfully.")
			l.Warn("Please visit the UI to view the graph by clicking the link below:")
			l.Warn("http://localhost:8888")
			// Yes, we should change that :D
			l.Warn("Default password being 'admin'")

			return nil
		},
		PersistentPreRunE: func(cobraCmd *cobra.Command, _ []string) error {
			return rootPreRun(cobraCmd, nil)
		},
		PersistentPostRunE: func(cobraCmd *cobra.Command, _ []string) error {
			defer rootPostRun(cobraCmd.Context())
			return cmd.CloseKubehoundConfig(cobraCmd.Context())
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Profiling flags
	cpuProfile   string
	memProfile   string
	blockProfile string

	// cpuProfileCleanup is called after the root command is executed to
	// cleanup a running cpu profile.
	cpuProfileCleanup func()

	// blockProfileCleanup is called after the root command is executed to
	// cleanup a running block profile.
	blockProfileCleanup func()
)

// rootPreRun is executed before the root command runs and sets up cpu
// profiling.
//
// Bassed on https://golang.org/pkg/runtime/pprof/#hdr-Profiling_a_Go_program
func rootPreRun(cobraCmd *cobra.Command, _ []string) error {
	l := log.Logger(cobraCmd.Context())

	if err := startCPUProfile(l); err != nil {
		return err
	}

	if err := startBlockProfile(l); err != nil {
		return err
	}

	return nil
}

func startCPUProfile(l log.LoggerI) error {
	if cpuProfile == "" {
		return nil
	}

	l.Info("starting cpu profile")

	f, err := os.Create(path.Clean(cpuProfile))
	if err != nil {
		return fmt.Errorf("%w: unable to create CPU profile file", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		if err := f.Close(); err != nil {
			l.Error("error while closing cpu profile file", log.ErrorField(err))
		}
		return err
	}

	cpuProfileCleanup = func() {
		l.Info("stopping cpu profile")
		pprof.StopCPUProfile()
		if err := f.Close(); err != nil {
			l.Error("error while closing cpu profile file", log.ErrorField(err))
		}
	}

	return nil
}

func startBlockProfile(l log.LoggerI) error {
	if blockProfile == "" {
		return nil
	}

	l.Info("starting block profile")

	runtime.SetBlockProfileRate(1)
	f, err := os.Create(path.Clean(blockProfile))
	if err != nil {
		return fmt.Errorf("%w: unable to create block profile file", err)
	}

	p := pprof.Lookup("block")
	blockProfileCleanup = func() {
		l.Info("writing block profile")
		if err := p.WriteTo(f, 0); err != nil {
			l.Error("error while writing block profile file", log.ErrorField(err))
		}
		if err := f.Close(); err != nil {
			l.Error("error while closing block profile file", log.ErrorField(err))
		}
	}

	return nil
}

// rootPostRun is executed after the root command runs and performs memory
// profiling.
func rootPostRun(ctx context.Context) {
	l := log.Logger(ctx)

	if cpuProfileCleanup != nil {
		cpuProfileCleanup()
	}

	if blockProfileCleanup != nil {
		blockProfileCleanup()
	}

	if memProfile != "" {
		l.Info("writing mem profiles")

		// writeProfile writes the profile to a file.
		writeProfile := func(profileName string, profile *pprof.Profile) {
			file, err := os.Create(path.Clean(memProfile + profileName))
			if err != nil {
				l.Error(fmt.Sprintf("error while creating mem-profile %s file", profileName), log.ErrorField(err))
				return
			}
			defer func() {
				if err := file.Close(); err != nil {
					l.Error(fmt.Sprintf("error while closing mem-profile %s file", profileName), log.ErrorField(err))
				}
			}()
			if err := profile.WriteTo(file, 0); err != nil {
				l.Error(fmt.Sprintf("error while writing mem-profile %s profile", profileName), log.ErrorField(err))
			}
		}

		// If the memProfile does not have a suffix, add a dot to the end.
		if !strings.HasSuffix(memProfile, ".") {
			memProfile += "."
		}

		// Trigger a garbage collection to get up-to-date statistics.
		runtime.GC()

		// Write the profiles.
		writeProfile("heap", pprof.Lookup("heap"))
		writeProfile("allocs", pprof.Lookup("allocs"))
		writeProfile("goroutine", pprof.Lookup("goroutine"))
	}
}

// Execute the root command.
func Execute() error {
	tag.SetupBaseTags()
	return rootCmd.Execute()
}

func init() {
	rootFlags := rootCmd.PersistentFlags()

	rootFlags.StringVarP(&cfgFile, "config", "c", cfgFile, "application config file")

	rootFlags.BoolVar(&skipBackend, "skip-backend", skipBackend, "skip the auto deployment of the backend stack (janusgraph, mongodb, and UI)")

	// Profiling flags
	rootFlags.StringVar(&cpuProfile, "cpu-profile", "", "Save the pprof cpu profile in the specified file")
	rootFlags.StringVar(&memProfile, "mem-profile", "", "Save the pprof mem profile in the specified file")
	rootFlags.StringVar(&blockProfile, "block-profile", "", "Save the pprof block profile in the specified file")

	cmd.InitRootCmd(rootCmd)
}
