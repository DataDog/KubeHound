package core

type launchConfig struct {
	ConfigPath string
}

type LaunchOption func(*launchConfig)

// WithConfigPath sets the path for the KubeHound config file.
func WithConfigPath(configPath string) LaunchOption {
	return func(lc *launchConfig) {
		lc.ConfigPath = configPath
	}
}
