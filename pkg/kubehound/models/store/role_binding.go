package store

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

type BindSubject struct {
	IdentityId int64          `json:"identity_id"`
	Subject    rbacv1.Subject `json:"subject"`
}

type RoleBinding struct {
	Id           int64
	Name         string
	RoleId       int64
	IsNamespaced bool
	Namespace    string
	Subjects     []BindSubject
	RoleRef      rbacv1.RoleRef
	Ownership    OwnershipInfo
	Runtime      RuntimeInfo
}
