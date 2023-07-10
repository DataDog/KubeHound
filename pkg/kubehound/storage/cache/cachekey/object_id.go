package cachekey

const (
	objectIdCacheName = "store-graph-id"
)

type objectIDCacheKey struct {
	storeID string
}

var _ CacheKey = (*objectIDCacheKey)(nil) // Ensure interface compliance

func ObjectID(storeID string) *objectIDCacheKey {
	return &objectIDCacheKey{
		storeID: storeID,
	}
}

func (k *objectIDCacheKey) Shard() string {
	return objectIdCacheName
}

func (k *objectIDCacheKey) Key() string {
	return k.storeID
}
