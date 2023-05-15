package cache

const (
	roleCacheName = "k8s-role"
)

type roleCacheKey struct {
	name string
}

var _ CacheKey = (*roleCacheKey)(nil) // Ensure interface compliance

func RoleKey(name string) *roleCacheKey {
	return &roleCacheKey{
		name: name,
	}
}

func (k *roleCacheKey) Namespace() string {
	return roleCacheName
}

func (k *roleCacheKey) Key() string {
	return k.name
}
