package edge

import (
	"sync"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

// EdgeRegistry holds details of edges (i.e attacks) registered in KubeHound.
type EdgeRegistry map[string]Builder

// EdgeRegistry singleton support
var registryInstance EdgeRegistry
var erOnce sync.Once

// Registry returns the EdgeRegistry singleton.
func Registry() EdgeRegistry {
	erOnce.Do(func() {
		registryInstance = make(EdgeRegistry)
	})

	return registryInstance
}

// Register loads the provided edge into the registry.
func Register(edge Builder) {
	log.I.Infof("Registering edge %s", edge.Label())
	Registry()[edge.Label()] = edge
}
