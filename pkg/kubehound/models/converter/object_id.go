package converter

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
)

// ObjectIDConverter enables converting between an store object ID and an existing graph vertex ID.
type ObjectIDConverter struct {
	cache cache.CacheReader
}

// NewObjectID creates a new ObjectIdConverter instance from the provided cache reader.
func NewObjectID(cache cache.CacheReader) *ObjectIDConverter {
	return &ObjectIDConverter{
		cache: cache,
	}
}

// GraphID will return the graph vertex ID corresponding to the provided storer ID.
func (c *ObjectIDConverter) GraphID(ctx context.Context, storeID string) (int64, error) {
	if c.cache == nil {
		return -1, ErrNoCacheInitialized
	}

	vid, err := c.cache.Get(ctx, cachekey.ObjectID(storeID)).Int64()
	if err != nil {
		return -1, fmt.Errorf("graph id cache fetch (storeID=%s): %w", storeID, err)
	}

	return vid, nil
}
