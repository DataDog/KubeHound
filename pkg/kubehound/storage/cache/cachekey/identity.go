package cachekey

import (
	"strings"
)

const (
	identityCacheName = "k8s-identity"
)

type identityCacheKey struct {
	baseCacheKey
}

var _ CacheKey = (*identityCacheKey)(nil) // Ensure interface compliance

func Identity(identityName string, namespace string) *identityCacheKey {
	var sb strings.Builder

	sb.WriteString(namespace)
	sb.WriteString(CacheKeySeparator)
	sb.WriteString(identityName)

	return &identityCacheKey{
		baseCacheKey{sb.String()},
	}
}

func (k *identityCacheKey) Shard() string {
	return identityCacheName
}
