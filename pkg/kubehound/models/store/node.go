package store

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	corev1 "k8s.io/api/core/v1"
)

type Node struct {
	Id           primitive.ObjectID `bson:"_id"`
	UserId       primitive.ObjectID `bson:"user_id"`
	IsNamespaced bool               `bson:"is_namespaced"`
	K8           corev1.Node        `bson:"k8"`
	Ownership    OwnershipInfo      `bson:"ownership"`
}

// TODO index
// user_id,
