package converter

import (
	"context"
	"fmt"

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

func (c *ObjectIdConverter) GraphId(ctx context.Context, storeID string) (int64, error) {
	if c.cache == nil {
		return -1, ErrNoCacheInitialized
	}

	vid, err := c.cache.Get(ctx, cachekey.ObjectId(storeID)).Int64()
	if err != nil {
		return -1, fmt.Errorf("graph id cache fetch: %w", err)
	}

	return vid, nil
}
