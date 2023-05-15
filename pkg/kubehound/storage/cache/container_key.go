package cache

const (
	containerCacheName = "k8s-container"
)

type containerCacheKey struct {
	pod       string
	container string
}

var _ CacheKey = (*containerCacheKey)(nil) // Ensure interface compliance

func ContainerKey(pod string, container string) *containerCacheKey {
	return &containerCacheKey{
		pod:       pod,
		container: container,
	}
}

func (k *containerCacheKey) Namespace() string {
	return containerCacheName
}

func (k *containerCacheKey) Key() string {
	return k.pod + "##" + k.container
}
