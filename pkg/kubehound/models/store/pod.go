package store

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	corev1 "k8s.io/api/core/v1"
)

type Pod struct {
	Id           primitive.ObjectID `bson:"_id"`
	NodeId       primitive.ObjectID `bson:"node_id"`
	IsNamespaced bool               `bson:"is_namespaced"`
	K8           corev1.Pod         `bson:"k8"`
	Ownership    OwnershipInfo      `bson:"ownership"`
	Runtime      RuntimeInfo        `bson:"runtime"`
}
