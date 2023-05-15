package cache

const (
	nodeCacheName = "k8s-node"
)

type nodeCacheKey struct {
	name string
}

var _ CacheKey = (*nodeCacheKey)(nil) // Ensure interface compliance

func NodeKey(name string) *nodeCacheKey {
	return &nodeCacheKey{
		name: name,
	}
}

func (k *nodeCacheKey) Namespace() string {
	return nodeCacheName
}

func (k *nodeCacheKey) Key() string {
	return k.name
}
