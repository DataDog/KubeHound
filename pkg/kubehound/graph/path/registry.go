package path

import (
	"sync"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

// Registry holds details of paths registered in KubeHound to be constructed via
// store queries (not directly via pipeline ingest).
type Registry map[string]Builder

// PathRegistry singleton support.
var registryInstance Registry
var erOnce sync.Once

// Registered returns the path registry singleton.
func Registered() Registry {
	erOnce.Do(func() {
		registryInstance = make(Registry)
	})

	return registryInstance
}

// Register loads the provided path into the registry.
func Register(path Builder) {
	log.I.Infof("Registering path %s", path.Label())
	Registered()[path.Label()] = path
}
