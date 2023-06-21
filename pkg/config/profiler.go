package config

import "time"

// ProfilerConfig configures profiler specific parameters.
type ProfilerConfig struct {
	Period      time.Duration `mapstructure:"period"`
	CPUDuration time.Duration `mapstructure:"cpu_duration"`
}
