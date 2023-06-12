package cachekey

const (
	podIdentityCacheName = "k8s-pod-identity"
)

type podIdentityCacheKey struct {
	podId string
}

var _ CacheKey = (*podIdentityCacheKey)(nil) // Ensure interface compliance

func PodIdentity(podId string) *podIdentityCacheKey {
	return &podIdentityCacheKey{
		podId: podId,
	}
}

func (k *podIdentityCacheKey) Shard() string {
	return podIdentityCacheName
}

func (k *podIdentityCacheKey) Key() string {
	return k.podId
}
