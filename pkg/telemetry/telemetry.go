package telemetry

import (
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/KubeHound/pkg/telemetry/tracer"
)

// Initialize all telemetry required
// return client to enable clean shutdown
func Initialize(cfg *config.KubehoundConfig) error {
	//Tracing
	tracer.Initialize(cfg)
	// Metrics
	err := statsd.Setup(cfg.Telemetry.Statsd.URL)
	if err != nil {
		return err
	}

	return nil
}

func Shutdown() {
	//Tracing
	tracer.Shutdown()
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
