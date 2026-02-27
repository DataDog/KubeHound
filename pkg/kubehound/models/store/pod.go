package store

type Pod struct {
	Id                    int64
	NodeId                int64
	IsNamespaced          bool
	Name                  string
	Namespace             string
	NodeName              string
	ServiceAccount        string
	HostPID               bool
	HostIPC               bool
	HostNetwork           bool
	ShareProcessNamespace bool
	PodIP                 string
	UID                   string
	Ownership             OwnershipInfo
	Runtime               RuntimeInfo
}
