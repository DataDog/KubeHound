package store

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultEndpointProtocol = "TCP"
	DefaultPortName         = ""
)

// Ok do not do the extra work

// Endpoint Ingest -> Write cache entry. Write without Ids
// Pod Ingest -> Check cache, if no matyching entry write everything. Set flag for dervived. Otherwise skip

// Edge calculatiopn

// Internal -> simply lookup derived flag = true and link endpoint -> container directly via Id field
// External -> lookup derived flag = fasle, match pod name sin endpoint trarget to pod ids, link edges

// TODO remove podid nodeid fields!
type Endpoint struct {
	Id           primitive.ObjectID        `bson:"_id"`
	ContainerId  primitive.ObjectID        `bson:"container_id"`
	PodName      string                    `bson:"pod_name"`
	NodeName     string                    `bson:"node_name"`
	IsNamespaced bool                      `bson:"is_namespaced"`
	Namespace    string                    `bson:"namespace"`
	Name         string                    `bson:"name"`
	HasSlice     bool                      `bson:"has_slice"`
	ServiceName  string                    `bson:"service_name"`
	K8           metav1.ObjectMeta         `bson:"k8"`
	AddressType  discoveryv1.AddressType   `bson:"address_type"`
	Backend      discoveryv1.Endpoint      `bson:"backend"`
	Port         discoveryv1.EndpointPort  `bson:"port"`
	Ownership    OwnershipInfo             `bson:"ownership"`
	Access       shared.EndpointAccessType `bson:"access"`
}

func (e *Endpoint) SafePort() int {
	if e.Port.Port == nil {
		return 0
	}

	return int(*e.Port.Port)
}

func (e *Endpoint) SafeProtocol() string {
	if e.Port.Protocol == nil {
		return DefaultEndpointProtocol
	}

	return string(*e.Port.Protocol)
}

func (e *Endpoint) SafePortName() string {
	if e.Port.Name == nil {
		return DefaultPortName
	}

	return string(*e.Port.Name)
}
