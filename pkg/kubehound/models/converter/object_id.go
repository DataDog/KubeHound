package converter

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
)

// ObjectIdConverter enables converting between an store object ID and an existing graph vertex ID.
type ObjectIdConverter struct {
	cache cache.CacheReader
}

func NewObjectId(cache cache.CacheReader) *ObjectIdConverter {
	return &ObjectIdConverter{
		cache: cache,
	}
}

// NOTE: requires cache access (NodeKey).
func (c *ObjectIdConverter) GraphId(ctx context.Context, storeID string) (int64, error) {
	if c.cache == nil {
		return -1, ErrNoCacheInitialized
	}

	vid, err := c.cache.Get(ctx, cachekey.ObjectId(storeID))
	if err != nil {
		return -1, err
	}

	// TODO typoen convert
	return vid, nil
}
