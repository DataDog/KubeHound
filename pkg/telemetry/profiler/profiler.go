package profiler

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func Initialize(cfg *config.KubehoundConfig) {
	l := log.Logger(context.TODO())
	opts := []profiler.Option{
		profiler.WithService(globals.DDServiceName),
		profiler.WithEnv(globals.GetDDEnv()),
		profiler.WithVersion(config.BuildVersion),
		profiler.WithProfileTypes(
			profiler.CPUProfile,
			profiler.HeapProfile,
			// The profiles below are disabled by default to keep overhead
			// low, but can be enabled as needed.

			// profiler.BlockProfile,
			// profiler.MutexProfile,
			// profiler.GoroutineProfile,
		),
		profiler.WithPeriod(cfg.Telemetry.Profiler.Period),
		profiler.CPUDuration(cfg.Telemetry.Profiler.CPUDuration),
		profiler.WithLogStartup(false),
		profiler.WithTags(tag.GetBaseTags()...),
	}
	if cfg.Telemetry.Tracer.URL != "" {
		opts = append(opts, profiler.WithAgentAddr(cfg.Telemetry.Tracer.URL))
	}

	err := profiler.Start(opts...)
	if err != nil {
		l.Error("start profiler", log.ErrorField(err))
	}
}

func Shutdown() {
	profiler.Stop()
}
