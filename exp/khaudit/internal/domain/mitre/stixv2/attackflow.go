package stixv2

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AttackFlow is a type of object that represents an attack flow.
type AttackFlow struct {
	DomainObject `json:",inline"`
	Scope        string   `json:"scope,omitempty"`
	StartRefs    []string `json:"start_refs,omitempty"`
}

// NewAttackFlow creates a new attack flow.
func NewAttackFlow() *AttackFlow {
	// Get the current time.
	now := time.Now().UTC()

	// Create the attack flow.
	flow := &AttackFlow{
		DomainObject: DomainObject{
			Type:         "attack-flow",
			SpecVersion:  "2.1",
			ID:           fmt.Sprintf("attack-flow--%s", uuid.New().String()),
			Created:      now.Format(time.RFC3339Nano),
			Modified:     now.Format(time.RFC3339Nano),
			CreatedByRef: fmt.Sprintf("identity--%s", khUUID),
			Name:         "Kubehound Attack Flow",
			Description:  "The Kubehound Attack Flow is a collection of attack actions that are executed to gain access to a target environment.",
			ExternalReferences: []ExternalReference{
				{
					SourceName:  "Datadog Kubehound",
					Description: "Kubernetes attack path discovery tool",
					URL:         "https://kubehound.io",
				},
			},
			Extensions: map[string]Extension{
				"extension-definition--fb9c968a-745b-4ade-9b25-c324172197f4": {
					ExtensionType: "new-sdo",
				},
			},
		},
		Scope:     "threat-actor",
		StartRefs: []string{},
	}

	// Return the attack flow.
	return flow
}

// AttackAction is a type of object that represents an attack action.
type AttackAction struct {
	DomainObject   `json:",inline"`
	TechniqueID    string   `json:"technique_id,omitempty"`
	TechniqueRef   string   `json:"technique_ref,omitempty"`
	TacticID       string   `json:"tactic_id,omitempty"`
	TacticRef      string   `json:"tactic_ref,omitempty"`
	ExecutionStart string   `json:"execution_start,omitempty"`
	ExecutionEnd   string   `json:"execution_end,omitempty"`
	CommandRef     string   `json:"command_ref,omitempty"`
	AssetRefs      []string `json:"asset_refs,omitempty"`
	EffectRefs     []string `json:"effect_refs,omitempty"`
}

// NewAttackAction creates a new attack action.
func NewAttackAction(name, description string) *AttackAction {
	// Get the current time.
	now := time.Now().UTC()

	// Create the attack action.
	action := &AttackAction{
		DomainObject: DomainObject{
			Type:        "attack-action",
			SpecVersion: "2.1",
			ID:          fmt.Sprintf("attack-action--%s", uuid.New().String()),
			Created:     now.Format(time.RFC3339Nano),
			Modified:    now.Format(time.RFC3339Nano),
			Name:        name,
			Description: description,
			Extensions: map[string]Extension{
				"extension-definition--fb9c968a-745b-4ade-9b25-c324172197f4": {
					ExtensionType: "new-sdo",
				},
			},
		},
	}

	// Return the attack action.
	return action
}

// AttackAsset is a type of object that represents an attack asset.
type AttackAsset struct {
	DomainObject `json:",inline"`
	ObjectRef    string `json:"object_ref,omitempty"`
}

// NewAttackAsset creates a new attack asset.
func NewAttackAsset(name, description string) *AttackAsset {
	// Get the current time.
	now := time.Now().UTC()

	// Create the attack asset.
	asset := &AttackAsset{
		DomainObject: DomainObject{
			Type:        "attack-asset",
			SpecVersion: "2.1",
			ID:          fmt.Sprintf("attack-asset--%s", uuid.New().String()),
			Created:     now.Format(time.RFC3339Nano),
			Modified:    now.Format(time.RFC3339Nano),
			Name:        name,
			Description: description,
			Extensions: map[string]Extension{
				"extension-definition--fb9c968a-745b-4ade-9b25-c324172197f4": {
					ExtensionType: "new-sdo",
				},
			},
		},
	}

	// Return the attack asset.
	return asset
}
