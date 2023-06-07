package path

import (
	"sync"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

// PathRegistry holds details of paths registered in KubeHound to be constructed via
// store queries (not directly via pipeline ingest).
type PathRegistry map[string]Builder

// PathRegistry singleton support.
var registryInstance PathRegistry
var erOnce sync.Once

// Registry returns the VertexRegistry singleton.
func Registry() PathRegistry {
	erOnce.Do(func() {
		registryInstance = make(PathRegistry)
	})

	return registryInstance
}

// Register loads the provided path into the registry.
func Register(path Builder) {
	log.I.Infof("Registering path %s", path.Label())
	Registry()[path.Label()] = path
}
