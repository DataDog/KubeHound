package store

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	rbacv1 "k8s.io/api/rbac/v1"
)

type PermissionSet struct {
	Id              primitive.ObjectID  `bson:"_id"`
	RoleId          primitive.ObjectID  `bson:"role_id"`
	RoleName        string              `bson:"role_name"`
	RoleBindingId   primitive.ObjectID  `bson:"role_binding_id"`
	RoleBindingName string              `bson:"role_binding_name"`
	Name            string              `bson:"name"`
	IsNamespaced    bool                `bson:"is_namespaced"`
	Namespace       string              `bson:"namespace"`
	Rules           []rbacv1.PolicyRule `bson:"rules"`
	Ownership       OwnershipInfo       `bson:"ownership"`
}

// TODO index
// role_id, role_name, role_binding_id,role_binding_name, name, is_namespaced, namespace,
