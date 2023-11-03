package store

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	corev1 "k8s.io/api/core/v1"
)

type Volume struct {
	Id          primitive.ObjectID `bson:"_id"`
	PodId       primitive.ObjectID `bson:"pod_id"`
	NodeId      primitive.ObjectID `bson:"node_id"`
	ContainerId primitive.ObjectID `bson:"container_id"`
	ProjectedId primitive.ObjectID `bson:"projected_id"`
	Name        string             `bson:"name"`
	Type        string             `bson:"type"`
	SourcePath  string             `bson:"source"`
	MountPath   string             `bson:"mount"`
	ReadOnly    bool               `bson:"readonly"`
	Ownership   OwnershipInfo      `bson:"ownership"`
	Runtime     RuntimeInfo        `bson:"runtime"`
	K8          corev1.Volume      `bson:"k8"`
}
