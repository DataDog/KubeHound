package edge

import (
	"sync"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

type RegistrationFlag uint8

const (
	RegisterDefault       RegistrationFlag = 1 << iota // Default edge
	RegisterGraphMutation                              // Edge can mutate the graph
)

// Registry holds details of edges (i.e attacks) registered in KubeHound.
type Registry struct {
	mutating map[string]Builder
	simple   map[string]Builder
}

// newRegistry creates a new registry instance. This should not be called directly.
func newRegistry() *Registry {
	r := &Registry{
		mutating: make(map[string]Builder),
		simple:   make(map[string]Builder),
	}

	return r
}

// Registry singleton support
var registryInstance *Registry
var erOnce sync.Once

// Registered returns the edge registry singleton.
func Registered() *Registry {
	erOnce.Do(func() {
		registryInstance = newRegistry()
	})

	return registryInstance
}

// Mutating returns the map of registered mutating edge builders.
func (r *Registry) Mutating() map[string]Builder {
	return r.mutating
}

// Simple returns the map of registered edge builders.
func (r *Registry) Simple() map[string]Builder {
	return r.simple
}

// Register loads the provided edge into the registry.
func Register(edge Builder, flags RegistrationFlag) {
	registry := Registered()
	if flags&RegisterGraphMutation != 0 {
		log.I.Debugf("Registering mutating edge builder %s -> %s", edge.Name(), edge.Label())

		if _, ok := registry.mutating[edge.Name()]; ok {
			log.I.Fatalf("edge name collision: %s", edge.Name())
		}

		registry.mutating[edge.Name()] = edge
	} else {
		log.I.Debugf("Registering default edge builder %s -> %s", edge.Name(), edge.Label())
		if _, ok := registry.simple[edge.Name()]; ok {
			log.I.Fatalf("edge name collision: %s", edge.Name())
		}

		registry.simple[edge.Name()] = edge
	}
}
