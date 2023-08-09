package cachekey

const (
	nodeCacheName = "k8s-node"
)

type nodeCacheKey struct {
	baseCacheKey
}

var _ CacheKey = (*nodeCacheKey)(nil) // Ensure interface compliance

func Node(nodeName string) *nodeCacheKey {
	return &nodeCacheKey{
		baseCacheKey{nodeName},
	}
}

func (k *nodeCacheKey) Shard() string {
	return nodeCacheName
}
