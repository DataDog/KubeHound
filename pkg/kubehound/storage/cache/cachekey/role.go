package cachekey

const (
	roleCacheName = "k8s-role"
)

type roleCacheKey struct {
	roleName string
}

var _ CacheKey = (*roleCacheKey)(nil) // Ensure interface compliance

func Role(roleName string) *roleCacheKey {
	return &roleCacheKey{
		roleName: roleName,
	}
}

func (k *roleCacheKey) Namespace() string {
	return roleCacheName
}

func (k *roleCacheKey) Key() string {
	return k.roleName
}
