package graph

type CompromiseType int

const (
	CompromiseNone CompromiseType = iota
	CompromiseSimulated
	CompromiseKnown
)

type Container struct {
	StoreId      string         `json:"store_id" mapstructure:"store_id"`
	Name         string         `json:"name" mapstructure:"name"`
	Image        string         `json:"image" mapstructure:"image"`
	Command      []string       `json:"command" mapstructure:"command"`
	Args         []string       `json:"args" mapstructure:"args"`
	Capabilities []string       `json:"capabilities" mapstructure:"capabilities"`
	Privileged   bool           `json:"privileged" mapstructure:"privileged"`
	PrivEsc      bool           `json:"privesc" mapstructure:"privesc"`
	HostPID      bool           `json:"hostPid" mapstructure:"hostPid"`
	HostPath     bool           `json:"hostPath" mapstructure:"hostPath"`
	HostIPC      bool           `json:"hostIpc" mapstructure:"hostIpc"`
	HostNetwork  bool           `json:"hostNetwork" mapstructure:"hostNetwork"`
	RunAsUser    int64          `json:"runAsUser" mapstructure:"runAsUser"`
	Ports        []int          `json:"ports" mapstructure:"ports"`
	Pod          string         `json:"pod" mapstructure:"pod"`
	Node         string         `json:"node" mapstructure:"node"`
	Compromised  CompromiseType `json:"compromised,omitempty" mapstructure:"compromised"`
	Critical     bool           `json:"critical,omitempty" mapstructure:"critical"`
}

type Group struct {
	StoreId string `json:"store_id" mapstructure:"store_id"`
	Name    string `json:"name" mapstructure:"name"`
}

type Identity struct {
	StoreId      string `json:"store_id" mapstructure:"store_id"`
	Name         string `json:"name" mapstructure:"name"`
	IsNamespaced bool   `json:"is_namespaced" mapstructure:"is_namespaced"`
	Namespace    string `json:"namespace" mapstructure:"namespace"`
	Type         string `json:"type" mapstructure:"type"`
}

type Node struct {
	StoreId      string         `json:"store_id" mapstructure:"store_id"`
	Name         string         `json:"name" mapstructure:"name"`
	IsNamespaced bool           `json:"is_namespaced" mapstructure:"is_namespaced"`
	Namespace    string         `json:"namespace" mapstructure:"namespace"`
	Compromised  CompromiseType `json:"compromised,omitempty" mapstructure:"compromised"`
	Critical     bool           `json:"critical,omitempty" mapstructure:"critical"`
}

type Pod struct {
	StoreId                string         `json:"store_id" mapstructure:"store_id"`
	Name                   string         `json:"name" mapstructure:"name"`
	IsNamespaced           bool           `json:"is_namespaced" mapstructure:"is_namespaced"`
	Namespace              string         `json:"namespace" mapstructure:"namespace"`
	SharedProcessNamespace bool           `json:"sharedProcessNamespace" mapstructure:"sharedProcessNamespace"`
	ServiceAccount         string         `json:"serviceAccount" mapstructure:"serviceAccount"`
	Node                   string         `json:"node" mapstructure:"node"`
	Compromised            CompromiseType `json:"compromised,omitempty" mapstructure:"compromised"`
	Critical               bool           `json:"critical,omitempty" mapstructure:"critical"`
}

type Role struct {
	StoreId      string   `json:"store_id" mapstructure:"store_id"`
	Name         string   `json:"name" mapstructure:"name"`
	IsNamespaced bool     `json:"is_namespaced" mapstructure:"is_namespaced"`
	Namespace    string   `json:"namespace" mapstructure:"namespace"`
	Rules        []string `json:"rules" mapstructure:"rules"`
}

type Token struct {
	StoreId     string         `json:"store_id" mapstructure:"store_id"`
	Name        string         `json:"name" mapstructure:"name"`
	Type        string         `json:"type" mapstructure:"type"`
	Identity    string         `json:"identity" mapstructure:"identity"`
	Path        string         `json:"path" mapstructure:"path"`
	Compromised CompromiseType `json:"compromised,omitempty" mapstructure:"compromised"`
	Critical    bool           `json:"critical,omitempty" mapstructure:"critical"`
}

const (
	VolumeTypeHost      = "HostPath"
	VolumeTypeProjected = "Projected"
)

type Volume struct {
	StoreId string `json:"store_id" mapstructure:"store_id"`
	Name    string `json:"name" mapstructure:"name"`
	Type    string `json:"type" mapstructure:"type"`
	Path    string `json:"path" mapstructure:"path"`
}
