package vertex

import (
	"sync"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

// VertexRegistry holds details of vertices registered in KubeHound to be constructed via
// store queries (not directly via pipeline ingest).
type VertexRegistry map[string]QueryBuilder

// VertexRegistry singleton support.
var registryInstance VertexRegistry
var erOnce sync.Once

// Registry returns the VertexRegistry singleton.
func Registry() VertexRegistry {
	erOnce.Do(func() {
		registryInstance = make(VertexRegistry)
	})

	return registryInstance
}

// Register loads the provided edge into the registry.
func Register(vertex QueryBuilder) {
	log.I.Infof("Registering vertex %s", vertex.Label())
	Registry()[vertex.Label()] = vertex
}
