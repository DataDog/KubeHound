package tracer

import (
	"strings"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func Initialize(cfg *config.KubehoundConfig) {
	log.I.Infof("Using %s for tracer URL", cfg.Telemetry.Tracer.URL)

	// Default options
	opts := []tracer.StartOption{
		tracer.WithEnv(globals.DDEnv),
		tracer.WithService(globals.DDServiceName),
		tracer.WithServiceVersion(config.BuildVersion),
		tracer.WithAgentAddr(cfg.Telemetry.Tracer.URL),
		tracer.WithLogStartup(false),
	}

	// Add the base tags
	for _, t := range tag.BaseTags {
		split := strings.Split(t, ":")
		if len(split) != 2 {
			log.I.Fatalf("Invalid base tag in telemtry initialization: %s", t)
		}
		opts = append(opts, tracer.WithGlobalTag(split[0], split[1]))
	}

	// Optional tags from configuration
	for tk, tv := range cfg.Telemetry.Tags {
		opts = append(opts, tracer.WithGlobalTag(tk, tv))
	}

	tracer.Start(opts...)
}

func Shutdown() {
	tracer.Stop()
}
