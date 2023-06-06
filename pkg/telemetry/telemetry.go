package telemetry

import (
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/telemetry/tracer"
)

type TelemetryClient struct {
}

// Initialize all telemetry required
// return client to enable clean shutdown
func Initialize(cfg *config.KubehoundConfig) (*TelemetryClient, error) {
	tracer.Initialize(cfg)
	return nil, nil
}

func (t *TelemetryClient) Shutdown() {
	tracer.Shutdown()
}
