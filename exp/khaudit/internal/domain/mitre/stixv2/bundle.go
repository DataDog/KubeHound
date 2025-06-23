package stixv2

import (
	"encoding/json"
	"fmt"
	"time"
)

// Bundle is a type of object that contains a collection of STIX objects.
type Bundle struct {
	Type        string    `json:"type"`
	SpecVersion string    `json:"spec_version"`
	ID          string    `json:"id"`
	Created     string    `json:"created"`
	Modified    string    `json:"modified"`
	Objects     ObjectSet `json:"objects"`
}

// AddObject adds an object to the bundle.
func (b *Bundle) AddObject(objects ...Object) {
	for _, object := range objects {
		b.Objects[object.GetID()] = object
	}
}

// NewBundle creates a new bundle.
func NewBundle() *Bundle {
	// Get the current time.
	now := time.Now().UTC()

	// Register the MITRE identity.
	mitreIdentity := &Identity{
		DomainObject: DomainObject{
			Type:        "identity",
			SpecVersion: "2.1",
			ID:          fmt.Sprintf("identity--%s", mitreUUID),
			Created:     now.Format(time.RFC3339Nano),
			Modified:    now.Format(time.RFC3339Nano),
			Name:        "MITRE Engenuity Center for Threat-Informed Defense",
		},
		IdentityClass: "organization",
	}

	// Register the extension definition.
	extensionDefinition := &ExtensionDefinition{
		DomainObject: DomainObject{
			Type:         "extension-definition",
			SpecVersion:  "2.1",
			ID:           fmt.Sprintf("extension-definition--%s", mitreUUID),
			CreatedByRef: mitreIdentity.ID,
			Created:      now.Format(time.RFC3339Nano),
			Modified:     now.Format(time.RFC3339Nano),
			Name:         "Attack Flow",
			Description:  "Extends STIX 2.1 with features to create Attack Flows.",
			ExternalReferences: []ExternalReference{
				{
					SourceName:  "GitHub",
					Description: "Source code repository for Attack Flow",
					URL:         "https://github.com/center-for-threat-informed-defense/attack-flow",
				},
				{
					SourceName:  "Documentation",
					Description: "Documentation for Attack Flow",
					URL:         "https://center-for-threat-informed-defense.github.io/attack-flow",
				},
			},
		},
		Schema:         "https://center-for-threat-informed-defense.github.io/attack-flow/stix/attack-flow-schema-2.0.0.json",
		Version:        "2.0.0",
		ExtensionTypes: []string{"new-sdo"},
	}

	// Register the Datadog Kubehound identity.
	khIdentity := &Identity{
		DomainObject: DomainObject{
			Type:        "identity",
			SpecVersion: "2.1",
			ID:          fmt.Sprintf("identity--%s", khUUID),
			Created:     now.Format(time.RFC3339Nano),
			Modified:    now.Format(time.RFC3339Nano),
			Name:        "Datadog Kubehound",
		},
		IdentityClass: "organization",
	}

	// Add the objects to the object set.
	objects := ObjectSet{
		extensionDefinition.ID: extensionDefinition,
		mitreIdentity.ID:       mitreIdentity,
		khIdentity.ID:          khIdentity,
	}

	return &Bundle{
		Type:        "bundle",
		SpecVersion: "2.1",
		ID:          fmt.Sprintf("bundle--%s", mitreUUID),
		Created:     now.Format(time.RFC3339Nano),
		Modified:    now.Format(time.RFC3339Nano),
		Objects:     objects,
	}
}

// ObjectSet is a map of STIX objects by their ID.
// This is serialized as a JSON list of objects.
type ObjectSet map[string]Object

// MarshalJSON implements the json.Marshaler interface.
func (o ObjectSet) MarshalJSON() ([]byte, error) {
	objects := make([]Object, 0, len(o))
	for _, object := range o {
		objects = append(objects, object)
	}

	return json.Marshal(objects)
}
