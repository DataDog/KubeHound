package graph

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
)

type Pod struct {
	StoreID               string                `json:"storeID" mapstructure:"storeID"`
	App                   string                `json:"app" mapstructure:"app"`
	Team                  string                `json:"team" mapstructure:"team"`
	Service               string                `json:"service" mapstructure:"service"`
	Name                  string                `json:"name" mapstructure:"name"`
	IsNamespaced          bool                  `json:"isNamespaced" mapstructure:"isNamespaced"`
	Namespace             string                `json:"namespace" mapstructure:"namespace"`
	ShareProcessNamespace bool                  `json:"shareProcessNamespace" mapstructure:"shareProcessNamespace"`
	ServiceAccount        string                `json:"serviceAccount" mapstructure:"serviceAccount"`
	Node                  string                `json:"node" mapstructure:"node"`
	Compromised           shared.CompromiseType `json:"compromised" mapstructure:"compromised"`
	Critical              bool                  `json:"critical" mapstructure:"critical"`
}
