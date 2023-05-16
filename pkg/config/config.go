package config

import (
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

// KubehoundConfig defines the top-level application configuration for KubeHound.
type KubehoundConfig struct {
}

// MustLoadDefaultConfig loads the default application configuration, treating all errors as fatal.
func MustLoadDefaultConfig() *KubehoundConfig {
	log.I.Fatal(globals.ErrNotImplemented.Error())
	return nil
}
