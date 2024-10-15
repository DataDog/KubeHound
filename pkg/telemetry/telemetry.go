package telemetry

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/profiler"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/KubeHound/pkg/telemetry/tracer"
)

type State struct {
	Enabled bool
}

// Initialize all telemetry required
// return client to enable clean shutdown
func Initialize(ctx context.Context, khCfg *config.KubehoundConfig) error {
	l := log.Logger(ctx)
	if !khCfg.Telemetry.Enabled {
		l.Warn("Telemetry disabled via configuration")

		return nil
	}

	// Profiling
	profiler.Initialize(ctx, khCfg)

	// Tracing
	tracer.Initialize(ctx, khCfg)

	// Metrics
	err := statsd.Setup(ctx, khCfg)
	if err != nil {
		return err
	}

	return nil
}

func Shutdown(ctx context.Context, enabled bool) {
	l := log.Logger(ctx)
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
		l.Warnf("Failed to flush statsd client", log.ErrorField(err))
	}

	err = statsd.Close()
	if err != nil {
		l.Warnf("Failed to close statsd client", log.ErrorField(err))
	}
}
