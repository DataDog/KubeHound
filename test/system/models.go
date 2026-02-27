package system

import "github.com/DataDog/KubeHound/pkg/kubehound/models/shared"

// Test-only model types used for system test assertions.
// These replace the deleted pkg/kubehound/models/graph types.

type Pod struct {
	StoreID               string
	Name                  string
	IsNamespaced          bool
	Namespace             string
	Compromised           shared.CompromiseType
	ServiceAccount        string
	ShareProcessNamespace bool
	Node                  string
	Critical              bool
}

type Node struct {
	StoreID      string
	Name         string
	IsNamespaced bool
	Namespace    string
	Compromised  shared.CompromiseType
	Critical     bool
}

type Container struct {
	StoreID      string
	Name         string
	Image        string
	Command      []string
	Args         []string
	Capabilities []string
	Privileged   bool
	PrivEsc      bool
	HostPID      bool
	HostIPC      bool
	HostNetwork  bool
	RunAsUser    int64
	Ports        []string
	Pod          string
	Node         string
	IsNamespaced bool
	Namespace    string
	Compromised  shared.CompromiseType
}

type Volume struct {
	StoreID      string
	Name         string
	Type         string
	SourcePath   string
	MountPath    string
	Readonly     bool
	IsNamespaced bool
	Namespace    string
}

type Identity struct {
	StoreID      string
	Name         string
	IsNamespaced bool
	Namespace    string
	Type         string
	Critical     bool
}

type PermissionSet struct {
	StoreID      string
	Name         string
	IsNamespaced bool
	Namespace    string
	Role         string
	RoleBinding  string
	Rules        []string
	Critical     bool
}
