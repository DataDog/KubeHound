package profiler

import (
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func Initialize(cfg *config.KubehoundConfig) {
	err := profiler.Start(
		profiler.WithService(globals.DDServiceName),
		profiler.WithEnv(globals.DDEnv),
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
		profiler.WithPeriod(5*time.Second),
	)
	if err != nil {
		log.I.Errorf("start profiler: %v", err)
	}
}

func Shutdown() {
	profiler.Stop()
}
