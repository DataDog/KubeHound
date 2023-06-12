package config

type TelemetryConfig struct {
	Statsd StatsdConfig `mapstructure:"statsd"`
	Tracer TracerConfig `mapstructure:"tracer"`
}

// StatsdConfig configures statsd specific parameters.
type StatsdConfig struct {
	URL string `mapstructure:"url"`
}
