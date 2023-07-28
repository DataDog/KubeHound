package store

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	rbacv1 "k8s.io/api/rbac/v1"
)

type Role struct {
	Id           primitive.ObjectID  `bson:"_id"`
	Name         string              `bson:"name"`
	IsNamespaced bool                `bson:"is_namespaced"`
	Namespace    string              `bson:"namespace"`
	Rules        []rbacv1.PolicyRule `bson:"rules"`
	Ownership    OwnershipInfo       `bson:"ownership"`
}
