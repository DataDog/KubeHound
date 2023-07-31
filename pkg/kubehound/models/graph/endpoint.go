package graph

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
)

type Endpoint struct {
	StoreID      string `json:"storeID" mapstructure:"storeID"`
	App          string `json:"app" mapstructure:"app"`
	Team         string `json:"team" mapstructure:"team"`
	Service      string `json:"service" mapstructure:"service"`
	IsNamespaced bool   `json:"isNamespaced" mapstructure:"isNamespaced"`
	Namespace    string `json:"namespace" mapstructure:"namespace"`
	Name         string `json:"name" mapstructure:"name"`
	AddressType  string
	Addresses    []string
	Port         int
	PortName     string
	Protocol     string
	Access       shared.EndpointAccessType
	Compromised  shared.CompromiseType `json:"compromised" mapstructure:"compromised"`
}
