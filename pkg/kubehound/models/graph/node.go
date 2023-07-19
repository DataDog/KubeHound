package graph

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
)

type Node struct {
	StoreID      string                `json:"storeID" mapstructure:"storeID"`
	App          string                `json:"app" mapstructure:"app"`
	Team         string                `json:"team" mapstructure:"team"`
	Service      string                `json:"service" mapstructure:"service"`
	Name         string                `json:"name" mapstructure:"name"`
	IsNamespaced bool                  `json:"isNamespaced" mapstructure:"isNamespaced"`
	Namespace    string                `json:"namespace" mapstructure:"namespace"`
	Compromised  shared.CompromiseType `json:"compromised" mapstructure:"compromised"`
	Critical     bool                  `json:"critical" mapstructure:"critical"`
}
