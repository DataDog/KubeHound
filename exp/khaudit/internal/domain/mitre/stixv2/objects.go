// Package stixv2 contains the types for STIX v2 objects.
package stixv2

// -----------------------------------------------------------------------------

// DomainObject is the base type for all STIX objects.
type DomainObject struct {
	Type               string              `json:"type"`
	ID                 string              `json:"id"`
	SpecVersion        string              `json:"spec_version"`
	CreatedByRef       string              `json:"created_by_ref,omitempty"`
	Created            string              `json:"created"`
	Modified           string              `json:"modified"`
	Name               string              `json:"name"`
	Confidence         int                 `json:"confidence,omitempty"`
	Description        string              `json:"description,omitempty"`
	ExternalReferences []ExternalReference `json:"external_references,omitempty"`
	Extensions         Extensions          `json:"extensions,omitempty"`
}

// GetID returns the ID of the object.
func (o *DomainObject) GetID() string { return o.ID }

// GetType returns the type of the object.
func (o *DomainObject) GetType() string { return o.Type }

// ExternalReference is a reference to an external source.
type ExternalReference struct {
	SourceName  string `json:"source_name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// Extensions is a map of extension definitions.
type Extensions map[string]Extension

// ExtensionDefinition is a type of object that defines an extension.
type ExtensionDefinition struct {
	DomainObject   `json:",inline"`
	Schema         string   `json:"schema,omitempty"`
	Version        string   `json:"version,omitempty"`
	ExtensionTypes []string `json:"extension_types,omitempty"`
}

// Extension is a type of object that extends a STIX object.
type Extension struct {
	ExtensionType string `json:"extension_type,omitempty"`
}

// -----------------------------------------------------------------------------

// Identity is a type of object that represents an identity.
type Identity struct {
	DomainObject       `json:",inline"`
	IdentityClass      string `json:"identity_class"`
	ContactInformation string `json:"contact_information,omitempty"`
}
