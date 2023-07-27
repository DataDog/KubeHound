package config

import (
	"time"
)

const (
	DefaultProfilerPeriod      time.Duration = 60 * time.Second
	DefaultProfilerCPUDuration time.Duration = 15 * time.Second
)

type TelemetryConfig struct {
	Enabled  bool              `mapstructure:"enabled"`  // Whether or not to enable Datadog telemetry
	Tags     map[string]string `mapstructure:"tags"`     // Free form tags to be added to all telemetry
	Statsd   StatsdConfig      `mapstructure:"statsd"`   // Statsd configuration (for metrics)
	Tracer   TracerConfig      `mapstructure:"tracer"`   // Tracer configuration (for APM)
	Profiler ProfilerConfig    `mapstructure:"profiler"` // Profiler configuration
}

// StatsdConfig configures statsd specific parameters.
type StatsdConfig struct {
	URL string `mapstructure:"url"` // Statsd endpoint URL
}

// TracerConfig configures tracer specific parameters.
type TracerConfig struct {
	URL string `mapstructure:"url"` // Tracer endpoint URL
}
