package store

// ContainerInherited holds properties inherited from the parent Pod.
type ContainerInherited struct {
	Namespace      string
	PodName        string
	NodeName       string
	HostPID        bool
	HostIPC        bool
	HostNetwork    bool
	ServiceAccount string
	RunAsUser      int64
}

type Container struct {
	Id           int64
	PodId        int64
	NodeId       int64
	Name         string
	Image        string
	Command      []string
	Args         []string
	Capabilities []string
	Privileged   bool
	PrivEsc      bool
	Ports        []int32
	Inherited    ContainerInherited
	Ownership    OwnershipInfo
	Runtime      RuntimeInfo
}
