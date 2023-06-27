package edge

import (
	"sync"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

// Registry holds details of edges (i.e attacks) registered in KubeHound.
type Registry map[string]Builder

// EdgeRegistry singleton support
var registryInstance Registry
var erOnce sync.Once

// Registered returns the edge registry singleton.
func Registered() Registry {
	erOnce.Do(func() {
		registryInstance = make(Registry)
	})

	return registryInstance
}

// Register loads the provided edge into the registry.
func Register(edge Builder) {
	log.I.Infof("Registering edge builder edge %s -> %s", edge.Name(), edge.Label())

	registry := Registered()
	if _, ok := registry[edge.Name()]; ok {
		log.I.Fatalf("edge name collision: %s", edge.Name())
	}

	registry[edge.Name()] = edge
}
