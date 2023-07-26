package config

type TelemetryConfig struct {
	Enabled  bool           `mapstructure:"enabled"`
	Statsd   StatsdConfig   `mapstructure:"statsd"`
	Tracer   TracerConfig   `mapstructure:"tracer"`
	Profiler ProfilerConfig `mapstructure:"profiler"`
}

// StatsdConfig configures statsd specific parameters.
type StatsdConfig struct {
	URL string `mapstructure:"url"`
}
