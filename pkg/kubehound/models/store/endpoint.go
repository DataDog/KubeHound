package store

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Endpoint struct {
	Id           primitive.ObjectID        `bson:"_id"`
	ContainerId  primitive.ObjectID        `bson:"container_id"`
	PodId        primitive.ObjectID        `bson:"pod_id"`
	NodeId       primitive.ObjectID        `bson:"node_id"`
	IsNamespaced bool                      `bson:"is_namespaced"`
	Namespace    string                    `bson:"namespace"`
	Name         *string                   `bson:"name"`
	K8           metav1.ObjectMeta         `bson:"k8"`
	AddressType  discoveryv1.AddressType   `bson:"address_type"`
	Addresses    []discoveryv1.Endpoint    `bson:"endpoint_address"`
	Port         discoveryv1.EndpointPort  `bson:"port"`
	Ownership    OwnershipInfo             `bson:"ownership"`
	Access       shared.EndpointAccessType `bson:"access"`
}
