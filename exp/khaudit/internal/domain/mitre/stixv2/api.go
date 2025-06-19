package stixv2

import "github.com/google/uuid"

// Object is the interface for all STIX objects.
type Object interface {
	// GetID returns the ID of the object.
	GetID() string
	// GetType returns the type of the object.
	GetType() string
}

// -----------------------------------------------------------------------------

var (
	// MITRE UUID.
	mitreUUID = uuid.Must(uuid.Parse("fb9c968a-745b-4ade-9b25-c324172197f4"))
	// Datadog Kubehound UUID.
	khUUID = uuid.Must(uuid.Parse("161ee6dd-10a8-47a8-952b-f225b3dad2ce"))
)
