package store

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

type PermissionSet struct {
	Id              int64
	RoleId          int64
	RoleName        string
	RoleBindingId   int64
	RoleBindingName string
	Name            string
	IsNamespaced    bool
	Namespace       string
	Rules           []rbacv1.PolicyRule
	Ownership       OwnershipInfo
	Runtime         RuntimeInfo
}
