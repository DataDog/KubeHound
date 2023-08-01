package graph

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
)

type Endpoint struct {
	StoreID             string                    `json:"storeID" mapstructure:"storeID"`
	App                 string                    `json:"app" mapstructure:"app"`
	Team                string                    `json:"team" mapstructure:"team"`
	Service             string                    `json:"service" mapstructure:"service"`
	IsNamespaced        bool                      `json:"isNamespaced" mapstructure:"isNamespaced"`
	Namespace           string                    `json:"namespace" mapstructure:"namespace"`
	Name                string                    `json:"name" mapstructure:"name"`
	ServiceEndpointName string                    `json:"serviceEndpoint" mapstructure:"serviceEndpoint"`
	AddressType         string                    `json:"addressType" mapstructure:"addressType"`
	Addresses           []string                  `json:"addresses" mapstructure:"addresses"`
	Port                int                       `json:"port" mapstructure:"port"`
	PortName            string                    `json:"portName" mapstructure:"portName"`
	Protocol            string                    `json:"protocol" mapstructure:"protocol"`
	Access              shared.EndpointAccessType `json:"access" mapstructure:"access"`
	Compromised         shared.CompromiseType     `json:"compromised" mapstructure:"compromised"`
}
