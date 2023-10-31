package graph

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
)

type Container struct {
	StoreID      string                `json:"storeID" mapstructure:"storeID"`
	App          string                `json:"app" mapstructure:"app"`
	Team         string                `json:"team" mapstructure:"team"`
	Service      string                `json:"service" mapstructure:"service"`
	RunID        string                `json:"runID" mapstructure:"runID"`
	Cluster      string                `json:"cluster" mapstructure:"cluster"`
	IsNamespaced bool                  `json:"isNamespaced" mapstructure:"isNamespaced"`
	Namespace    string                `json:"namespace" mapstructure:"namespace"`
	Name         string                `json:"name" mapstructure:"name"`
	Image        string                `json:"image" mapstructure:"image"`
	Command      []string              `json:"command" mapstructure:"command"`
	Args         []string              `json:"args" mapstructure:"args"`
	Capabilities []string              `json:"capabilities" mapstructure:"capabilities"`
	Privileged   bool                  `json:"privileged" mapstructure:"privileged"`
	PrivEsc      bool                  `json:"privesc" mapstructure:"privesc"`
	HostPID      bool                  `json:"hostPid" mapstructure:"hostPid"`
	HostIPC      bool                  `json:"hostIpc" mapstructure:"hostIpc"`
	HostNetwork  bool                  `json:"hostNetwork" mapstructure:"hostNetwork"`
	RunAsUser    int64                 `json:"runAsUser" mapstructure:"runAsUser"`
	Ports        []string              `json:"ports" mapstructure:"ports"`
	Pod          string                `json:"pod" mapstructure:"pod"`
	Node         string                `json:"node" mapstructure:"node"`
	Compromised  shared.CompromiseType `json:"compromised" mapstructure:"compromised"`
}
