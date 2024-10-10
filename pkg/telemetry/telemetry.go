package telemetry

import (
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/telemetry/profiler"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/KubeHound/pkg/telemetry/tracer"
)

type State struct {
	Enabled bool
}

// Initialize all telemetry required
// return client to enable clean shutdown
func Initialize(khCfg *config.KubehoundConfig) error {
	if !khCfg.Telemetry.Enabled {
		//log.I..Warnf("Telemetry disabled via configuration")

		return nil
	}

	// Profiling
	profiler.Initialize(khCfg)

	// Tracing
	tracer.Initialize(khCfg)

	// Metrics
	err := statsd.Setup(khCfg)
	if err != nil {
		return err
	}

	return nil
}

func Shutdown(enabled bool) {
	if enabled {
		return
	}

	// Profiling
	profiler.Shutdown()

	// Tracing
	tracer.Shutdown()

	// Metrics
	err := statsd.Flush()
	if err != nil {
		//log.I..Warnf("Failed to flush statsd client: %v", err)
	}

	err = statsd.Close()
	if err != nil {
		//log.I..Warnf("Failed to close statsd client: %v", err)
	}
}
