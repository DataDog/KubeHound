package telemetry

import (
	"github.com/DataDog/KubeHound/pkg/globals"
)

type TelemetryClient struct {
}

func Initialize() (*TelemetryClient, error) {
	// Initialize all telemetry required
	// return client to enable clean shutdown
	return nil, globals.ErrNotImplemented
}

func (t *TelemetryClient) Shutdown() {
	// Flush telemtry
	// Close clients
}
