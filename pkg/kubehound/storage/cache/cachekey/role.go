package cachekey

const (
	roleCacheName = "k8s-role"
)

type roleCacheKey struct {
	roleName  string
	namespace string
}

var _ CacheKey = (*roleCacheKey)(nil) // Ensure interface compliance

func Role(roleName string, namespace string) *roleCacheKey {
	return &roleCacheKey{
		roleName:  roleName,
		namespace: namespace,
	}
}

func (k *roleCacheKey) Shard() string {
	return roleCacheName
}

func (k *roleCacheKey) Key() string {
	return k.namespace + "##" + k.roleName
}
