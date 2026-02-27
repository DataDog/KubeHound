package store

type Node struct {
	Id           int64
	UserId       int64
	IsNamespaced bool
	Name         string
	Namespace    string
	Ownership    OwnershipInfo
	Runtime      RuntimeInfo
}
