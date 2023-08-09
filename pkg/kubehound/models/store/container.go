package store

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	corev1 "k8s.io/api/core/v1"
)

// Properties that are interesting to attackers can be set at a Pod level such as hostPid, or container level such
// as capabilities. To simplify the graph model, the container node is chosen as the single source of truth for all host
// security related information. Any capabilities derived from the containing Pod are set ONLY on the container (and
// inheritance/override rules applied)
type ContainerInherited struct {
	Namespace      string `bson:"namespace"`
	PodName        string `bson:"pod_name"`
	NodeName       string `bson:"node_name"`
	HostPID        bool   `bson:"host_pid"`
	HostIPC        bool   `bson:"host_ipc"`
	HostNetwork    bool   `bson:"host_net"`
	ServiceAccount string `bson:"service_account"`
}

type Container struct {
	Id        primitive.ObjectID `bson:"_id"`
	PodId     primitive.ObjectID `bson:"pod_id"`
	NodeId    primitive.ObjectID `bson:"node_id"`
	Inherited ContainerInherited `bson:"inherited"`
	K8        corev1.Container   `bson:"k8"`
	Ownership OwnershipInfo      `bson:"ownership"`
}
