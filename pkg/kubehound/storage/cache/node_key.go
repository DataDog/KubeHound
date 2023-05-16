package cache

const (
	nodeCacheName = "k8s-node"
)

type nodeCacheKey struct {
	nodeName string
}

var _ CacheKey = (*nodeCacheKey)(nil) // Ensure interface compliance

func NodeKey(nodeName string) *nodeCacheKey {
	return &nodeCacheKey{
		nodeName: nodeName,
	}
}

func (k *nodeCacheKey) Namespace() string {
	return nodeCacheName
}

func (k *nodeCacheKey) Key() string {
	return k.nodeName
}
