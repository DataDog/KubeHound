package store

type Volume struct {
	Id              int64
	PodId           int64
	NodeId          int64
	ContainerId     int64
	ProjectedId     int64
	Name            string
	Type            string
	SourcePath      string
	MountPath       string
	TargetName      string
	TargetNamespace string
	ReadOnly        bool
	Ownership       OwnershipInfo
	Runtime         RuntimeInfo
}
