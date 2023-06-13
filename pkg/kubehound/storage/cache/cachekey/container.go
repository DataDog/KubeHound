package cachekey

const (
	containerCacheName = "k8s-container"
)

type containerCacheKey struct {
	podName       string
	containerName string
	namespace     string
}

var _ CacheKey = (*containerCacheKey)(nil) // Ensure interface compliance

func Container(podName string, containerName string, namespace string) *containerCacheKey {
	return &containerCacheKey{
		podName:       podName,
		containerName: containerName,
		namespace:     namespace,
	}
}

func (k *containerCacheKey) Shard() string {
	return containerCacheName
}

func (k *containerCacheKey) Key() string {
	return k.namespace + "##" + k.podName + "##" + k.containerName
}
