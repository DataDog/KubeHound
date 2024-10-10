package tracer

import (
	"context"
	"strings"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func Initialize(cfg *config.KubehoundConfig) {
	l := log.Logger(context.TODO())
	// Default options
	opts := []tracer.StartOption{
		tracer.WithEnv(globals.GetDDEnv()),
		tracer.WithService(globals.DDServiceName),
		tracer.WithServiceVersion(config.BuildVersion),
		tracer.WithLogStartup(false),
	}
	if cfg.Telemetry.Tracer.URL != "" {
		l.Infof("Using %s for tracer URL", cfg.Telemetry.Tracer.URL)
		opts = append(opts, tracer.WithAgentAddr(cfg.Telemetry.Tracer.URL))
	}

	// Add the base tags
	for _, t := range tag.GetBaseTags() {
		const tagSplitLen = 2
		split := strings.Split(t, ":")
		if len(split) != tagSplitLen {
			l.Fatal("Invalid base tag in telemetry initialization", log.String("tag", t))
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
