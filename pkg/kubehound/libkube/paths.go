package libkube

import (
	"fmt"
)

// ServiceAccountTokenPath returns the full path of a pod's service account token on the host node.
func ServiceAccountTokenPath(podUid string, volumeName string) string {
	return fmt.Sprintf("/var/lib/kubelet/pods/%s/volumes/kubernetes.io~projected/%s/token",
		podUid, volumeName)
}
