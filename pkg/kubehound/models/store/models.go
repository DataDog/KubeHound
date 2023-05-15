package store

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// Properties that are interesting to attackers can be set at a Pod level such as hostPid, or container level such
// as capabilities. To simplify the graph model, the container node is chosen as the single source of truth for all host
// security related information. Any capabilities derived from the containing Pod are set ONLY on the container (and
// inheritance/override rules applied)
type ContainerInherited struct {
	PodName        string `bson:"pod_name,omitempty"`
	NodeName       string `bson:"node_name,omitempty"`
	HostPID        bool   `bson:"host_pid,omitempty"`
	HostIPC        bool   `bson:"host_ipc,omitempty"`
	HostNetwork    bool   `bson:"host_net,omitempty"`
	ServiceAccount string `bson:"service_account"`
}

type Container struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	PodId     primitive.ObjectID `bson:"pod_id,omitempty"`
	NodeId    primitive.ObjectID `bson:"node_id,omitempty"`
	Inherited ContainerInherited `bson:"inherited,omitempty"`
	K8        corev1.Container   `bson:"k8,omitempty"`
}

type Pod struct {
	Id     primitive.ObjectID `bson:"_id,omitempty"`
	NodeId primitive.ObjectID `bson:"node_id,omitempty"`
	K8     corev1.Pod         `bson:"k8,omitempty"`
}

type Node struct {
	Id primitive.ObjectID `bson:"_id,omitempty"`
	K8 corev1.Node        `bson:"k8,omitempty"`
}

type VolumeMount struct {
	ContainerId primitive.ObjectID `bson:"container_id"`
	K8          corev1.VolumeMount `bson:"k8"`
}

type Volume struct {
	Id     primitive.ObjectID `bson:"_id"`
	NodeId primitive.ObjectID `bson:"node_id"`
	PodId  primitive.ObjectID `bson:"pod_id"`
	Name   string             `bson:"name"`
	Source corev1.Volume      `bson:"source"`
	Mounts []VolumeMount      `bson:"mounts"`
}

type Role struct {
	Id        primitive.ObjectID  `bson:"_id"`
	Name      string              `bson:"name"`
	Global    bool                `bson:"global"`
	Namespace string              `bson:"namespace"`
	Rules     []rbacv1.PolicyRule `bson:"rules"`
}

type BindSubject struct {
	IdentityId primitive.ObjectID `bson:"identity_id"`
	Subject    rbacv1.Subject     `bson:"subject"`
}

type RoleBinding struct {
	Id        primitive.ObjectID `bson:"_id"`
	Name      string             `bson:"name"`
	RoleId    primitive.ObjectID `bson:"role_id"`
	Global    bool               `bson:"global"`
	Namespace string             `bson:"namespace"`
	Subjects  []BindSubject      `bson:"subjects"`
}

type Identity struct {
	Id        primitive.ObjectID `bson:"_id"`
	Name      string             `bson:"name"`
	Namespace string             `bson:"namespace"`
	Type      string             `bson:"type"`
}
