package graph

import "github.com/DataDog/KubeHound/pkg/kubehound/models/shared"

type Container struct {
	StoreId      string                `json:"store_id"`
	Name         string                `json:"name"`
	Image        string                `json:"image"`
	Command      []string              `json:"command"`
	Args         []string              `json:"args"`
	Capabilities []string              `json:"capabilities"`
	Privileged   bool                  `json:"privileged"`
	PrivEsc      bool                  `json:"privesc"`
	HostPID      bool                  `json:"hostPid"`
	HostPath     bool                  `json:"hostPath"`
	HostIPC      bool                  `json:"hostIpc"`
	HostNetwork  bool                  `json:"hostNetwork"`
	RunAsUser    int64                 `json:"runAsUser"`
	Ports        []int                 `json:"ports"`
	Pod          string                `json:"pod"`
	Node         string                `json:"node"`
	Compromised  shared.CompromiseType `json:"compromised,omitempty"`
	Critical     bool                  `json:"critical,omitempty"`
}

type Group struct {
	StoreId string `json:"store_id"`
	Name    string `json:"name"`
}

type Identity struct {
	StoreId      string `json:"store_id"`
	Name         string `json:"name"`
	IsNamespaced bool   `json:"is_namespaced"`
	Namespace    string `json:"namespace"`
	Type         string `json:"type"`
}

type Node struct {
	StoreId      string                `json:"store_id"`
	Name         string                `json:"name"`
	IsNamespaced bool                  `json:"is_namespaced"`
	Namespace    string                `json:"namespace"`
	Compromised  shared.CompromiseType `json:"compromised,omitempty"`
	Critical     bool                  `json:"critical,omitempty"`
}

type Pod struct {
	StoreId                string                `json:"store_id"`
	Name                   string                `json:"name"`
	IsNamespaced           bool                  `json:"is_namespaced"`
	Namespace              string                `json:"namespace"`
	SharedProcessNamespace bool                  `json:"sharedProcessNamespace"`
	ServiceAccount         string                `json:"serviceAccount"`
	Node                   string                `json:"node"`
	Compromised            shared.CompromiseType `json:"compromised,omitempty"`
	Critical               bool                  `json:"critical,omitempty"`
}

type Role struct {
	StoreId      string   `json:"store_id"`
	Name         string   `json:"name"`
	IsNamespaced bool     `json:"is_namespaced"`
	Namespace    string   `json:"namespace"`
	Rules        []string `json:"rules"`
}

type Token struct {
	Name        string                `json:"name"`
	Namespace   string                `json:"namespace"`
	Type        string                `json:"type"`
	Identity    string                `json:"identity"`
	Compromised shared.CompromiseType `json:"compromised,omitempty"`
	Critical    bool                  `json:"critical,omitempty"`
}

type Volume struct {
	StoreId  string `json:"store_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	NodePath string `json:"node_path"`
}
