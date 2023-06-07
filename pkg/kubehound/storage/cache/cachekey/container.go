package cachekey

const (
	containerCacheName = "k8s-container"
)

type containerCacheKey struct {
	podName       string
	containerName string
}

var _ CacheKey = (*containerCacheKey)(nil) // Ensure interface compliance

func Container(podName string, containerName string) *containerCacheKey {
	return &containerCacheKey{
		podName:       podName,
		containerName: containerName,
	}
}

func (k *containerCacheKey) Namespace() string {
	return containerCacheName
}

func (k *containerCacheKey) Key() string {
	return k.podName + "##" + k.containerName
}
