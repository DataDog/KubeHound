package store

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	rbacv1 "k8s.io/api/rbac/v1"
)

type BindSubject struct {
	IdentityId primitive.ObjectID `bson:"identity_id"`
	Subject    rbacv1.Subject     `bson:"subject"`
}

type RoleBinding struct {
	Id           primitive.ObjectID `bson:"_id"`
	Name         string             `bson:"name"`
	RoleId       primitive.ObjectID `bson:"role_id"`
	IsNamespaced bool               `bson:"is_namespaced"`
	Namespace    string             `bson:"namespace"`
	Subjects     []BindSubject      `bson:"subjects"`
	Ownership    OwnershipInfo      `bson:"ownership"`
	K8           rbacv1.RoleRef     `bson:"k8"`
}

// TODO index
// name, role_id, is_namespaced, namespace,
// subject?
