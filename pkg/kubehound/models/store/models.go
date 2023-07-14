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

type Pod struct {
	Id           primitive.ObjectID `bson:"_id"`
	NodeId       primitive.ObjectID `bson:"node_id"`
	IsNamespaced bool               `bson:"is_namespaced"`
	K8           corev1.Pod         `bson:"k8"`
	Ownership    OwnershipInfo      `bson:"ownership"`
}

type Node struct {
	Id           primitive.ObjectID `bson:"_id"`
	IsNamespaced bool               `bson:"is_namespaced"`
	K8           corev1.Node        `bson:"k8"`
	Ownership    OwnershipInfo      `bson:"ownership"`
}

type VolumeMount struct {
	ContainerId primitive.ObjectID `bson:"container_id"`
	K8          corev1.VolumeMount `bson:"k8"`
	Ownership   OwnershipInfo      `bson:"ownership"`
}

// type Volume struct {
// 	Id        primitive.ObjectID `bson:"_id"`
// 	NodeId    primitive.ObjectID `bson:"node_id"`
// 	PodId     primitive.ObjectID `bson:"pod_id"`
// 	Name      string             `bson:"name"`
// 	Type      string             `bson:"type"`
// 	Source    corev1.Volume      `bson:"source"`
// 	Mounts    []VolumeMount      `bson:"mounts"`
// 	ReadOnly  bool               `bson:"readonly"`
// 	Ownership OwnershipInfo      `bson:"ownership"`
// }

type Volume struct {
	Id          primitive.ObjectID `bson:"_id"`
	PodId       primitive.ObjectID `bson:"pod_id"`
	NodeId      primitive.ObjectID `bson:"node_id"`
	ContainerId primitive.ObjectID `bson:"container_id"`
	Name        string             `bson:"name"`
	Type        string             `bson:"type"`
	SourcePath  string             `bson:"source"`
	MountPath   string             `bson:"mount"`
	ReadOnly    bool               `bson:"readonly"`
	Ownership   OwnershipInfo      `bson:"ownership"`
}

type Role struct {
	Id           primitive.ObjectID  `bson:"_id"`
	Name         string              `bson:"name"`
	IsNamespaced bool                `bson:"is_namespaced"`
	Namespace    string              `bson:"namespace"`
	Rules        []rbacv1.PolicyRule `bson:"rules"`
	Ownership    OwnershipInfo       `bson:"ownership"`
}

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
}

type Identity struct {
	Id           primitive.ObjectID `bson:"_id"`
	Name         string             `bson:"name"`
	IsNamespaced bool               `bson:"is_namespaced"`
	Namespace    string             `bson:"namespace"`
	Type         string             `bson:"type"`
	Ownership    OwnershipInfo      `bson:"ownership"`
}
