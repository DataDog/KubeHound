package tracer

import (
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func Initialize(cfg *config.KubehoundConfig) {
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
