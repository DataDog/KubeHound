package store

type Identity struct {
	Id           int64
	Name         string
	IsNamespaced bool
	Namespace    string
	Type         string
	Ownership    OwnershipInfo
	Runtime      RuntimeInfo
}
