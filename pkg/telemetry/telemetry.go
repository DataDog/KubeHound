package telemetry

import (
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
)

func Initialize(cfg *config.KubehoundConfig) error {
	// Metrics
	err := statsd.Setup(cfg.Statsd.URL)
	if err != nil {
		return err
	}

	return nil
}

func Shutdown() {
	// Metrics
	err := statsd.Flush()
	if err != nil {
		log.I.Warnf("Failed to flush statsd client: %v", err)
	}
	err = statsd.Close()
	if err != nil {
		log.I.Warnf("Failed to close statsd client: %v", err)
	}
}
