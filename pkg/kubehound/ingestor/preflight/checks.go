package preflight

import (
	"errors"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

// Volumes will not be ingested - use with caution!
var SkipVolumes = map[string]bool{
	"/var/run/datadog-agent": true,
}

// CheckNode checks an input K8s npode object and reports whether it should be ingested.
func CheckNode(node types.NodeType) (bool, error) {
	if node == nil {
		return false, errors.New("nil node input in preflight check")
	}

	return true, nil
}

// CheckPod checks an input K8s pod object and reports whether it should be ingested.
func CheckPod(pod types.PodType) (bool, error) {
	if pod == nil {
		return false, errors.New("nil pod input in preflight check")
	}

	// If the pod is not running we don't want to save it
	if pod.Status.Phase != "Running" {
		log.I.Debugf("pod %s::%s not running (status=%s), skipping ingest!",
			pod.Namespace, pod.Name, pod.Status.Phase)
		return false, nil
	}

	return true, nil
}

// CheckVolume checks an input K8s volume object and reports whether it should be ingested.
func CheckVolume(volume types.VolumeMountType) (bool, error) {
	if volume == nil {
		return false, errors.New("nil volume input in preflight check")
	}

	if SkipVolumes[volume.MountPath] {
		return false, nil
	}

	return true, nil
}

// CheckContainer checks an input K8s container object and reports whether it should be ingested.
func CheckContainer(container types.ContainerType) (bool, error) {
	if container == nil {
		return false, errors.New("nil container input in preflight check")
	}

	return true, nil
}

// CheckRole checks an input K8s role object and reports whether it should be ingested.
func CheckRole(role types.RoleType) (bool, error) {
	if role == nil {
		return false, errors.New("nil role input in preflight check")
	}

	return true, nil
}

// CheckClusterRole checks an input K8s cluster role object and reports whether it should be ingested.
func CheckClusterRole(role types.ClusterRoleType) (bool, error) {
	if role == nil {
		return false, errors.New("nil cluster role input in preflight check")
	}

	return true, nil
}

// CheckRoleBinding checks an input K8s role object and reports whether it should be ingested.
func CheckRoleBinding(rb types.RoleBindingType) (bool, error) {
	if rb == nil {
		return false, errors.New("nil role binding input in preflight check")
	}

	return true, nil
}

// CheckClusterRoleBinding checks an input K8s cluster role binding object and reports whether it should be ingested.
func CheckClusterRoleBinding(role types.ClusterRoleBindingType) (bool, error) {
	if role == nil {
		return false, errors.New("nil cluster role binding input in preflight check")
	}

	return true, nil
}
