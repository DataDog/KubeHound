package graph

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
)

type Container struct {
	StoreID      string                `json:"storeID" mapstructure:"storeID"`
	Name         string                `json:"name" mapstructure:"name"`
	Image        string                `json:"image" mapstructure:"image"`
	Command      []string              `json:"command" mapstructure:"command"`
	Args         []string              `json:"args" mapstructure:"args"`
	Capabilities []string              `json:"capabilities" mapstructure:"capabilities"`
	Privileged   bool                  `json:"privileged" mapstructure:"privileged"`
	PrivEsc      bool                  `json:"privesc" mapstructure:"privesc"`
	HostPID      bool                  `json:"hostPid" mapstructure:"hostPid"`
	HostPath     bool                  `json:"hostPath" mapstructure:"hostPath"`
	HostIPC      bool                  `json:"hostIpc" mapstructure:"hostIpc"`
	HostNetwork  bool                  `json:"hostNetwork" mapstructure:"hostNetwork"`
	RunAsUser    int64                 `json:"runAsUser" mapstructure:"runAsUser"`
	Ports        []string              `json:"ports" mapstructure:"ports"`
	Pod          string                `json:"pod" mapstructure:"pod"`
	Node         string                `json:"node" mapstructure:"node"`
	Compromised  shared.CompromiseType `json:"compromised" mapstructure:"compromised"`
}

type Group struct {
	StoreID  string `json:"storeID" mapstructure:"storeID"`
	Name     string `json:"name" mapstructure:"name"`
	Critical bool   `json:"critical" mapstructure:"critical"`
}

type Identity struct {
	StoreID      string `json:"storeID" mapstructure:"storeID"`
	Name         string `json:"name" mapstructure:"name"`
	IsNamespaced bool   `json:"isNamespaced" mapstructure:"isNamespaced"`
	Namespace    string `json:"namespace" mapstructure:"namespace"`
	Type         string `json:"type" mapstructure:"type"`
	Critical     bool   `json:"critical" mapstructure:"critical"`
}

type Node struct {
	StoreID      string                `json:"storeID" mapstructure:"storeID"`
	Name         string                `json:"name" mapstructure:"name"`
	IsNamespaced bool                  `json:"isNamespaced" mapstructure:"isNamespaced"`
	Namespace    string                `json:"namespace" mapstructure:"namespace"`
	Compromised  shared.CompromiseType `json:"compromised" mapstructure:"compromised"`
	Critical     bool                  `json:"critical" mapstructure:"critical"`
}

type Pod struct {
	StoreID                string                `json:"storeID" mapstructure:"storeID"`
	Name                   string                `json:"name" mapstructure:"name"`
	IsNamespaced           bool                  `json:"isNamespaced" mapstructure:"isNamespaced"`
	Namespace              string                `json:"namespace" mapstructure:"namespace"`
	SharedProcessNamespace bool                  `json:"sharedProcessNamespace" mapstructure:"sharedProcessNamespace"`
	ServiceAccount         string                `json:"serviceAccount" mapstructure:"serviceAccount"`
	Node                   string                `json:"node" mapstructure:"node"`
	Compromised            shared.CompromiseType `json:"compromised" mapstructure:"compromised"`
	Critical               bool                  `json:"critical" mapstructure:"critical"`
}

type Role struct {
	StoreID      string   `json:"storeID" mapstructure:"storeID"`
	Name         string   `json:"name" mapstructure:"name"`
	IsNamespaced bool     `json:"isNamespaced" mapstructure:"isNamespaced"`
	Namespace    string   `json:"namespace" mapstructure:"namespace"`
	Rules        []string `json:"rules" mapstructure:"rules"`
	Critical     bool     `json:"critical" mapstructure:"critical"`
}

type Volume struct {
	StoreID string `json:"storeID" mapstructure:"storeID"`
	Name    string `json:"name" mapstructure:"name"`
	Type    string `json:"type" mapstructure:"type"`
	Path    string `json:"path" mapstructure:"path"`
}
