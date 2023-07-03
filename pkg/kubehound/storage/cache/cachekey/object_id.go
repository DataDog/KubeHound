package cachekey

const (
	objectIdCacheName = "store-graph-id"
)

type objectIdCacheKey struct {
	storeID string
}

var _ CacheKey = (*nodeCacheKey)(nil) // Ensure interface compliance

func ObjectId(storeID string) *objectIdCacheKey {
	return &objectIdCacheKey{
		storeID: storeID,
	}
}

func (k *objectIdCacheKey) Shard() string {
	return objectIdCacheName
}

func (k *objectIdCacheKey) Key() string {
	return k.storeID
}
