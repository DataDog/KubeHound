package store

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

type Role struct {
	Id           int64
	Name         string
	IsNamespaced bool
	Namespace    string
	Rules        []rbacv1.PolicyRule
	Ownership    OwnershipInfo
	Runtime      RuntimeInfo
}
