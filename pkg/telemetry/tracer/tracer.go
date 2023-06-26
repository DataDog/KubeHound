package tracer

import (
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func Initialize(cfg *config.KubehoundConfig) {
	log.I.Infof("Using %s for tracer URL", cfg.Telemetry.Tracer.URL)
	tracer.Start(
		tracer.WithEnv(globals.DDEnv),
		tracer.WithService(globals.DDServiceName),
		tracer.WithServiceVersion(config.BuildVersion),
		tracer.WithAgentAddr(cfg.Telemetry.Tracer.URL),
	)
}

func Shutdown() {
	tracer.Stop()
}
