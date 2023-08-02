package cachekey

import (
	"strings"
)

const (
	roleCacheName = "k8s-role"
)

type roleCacheKey struct {
	baseCacheKey
}

var _ CacheKey = (*roleCacheKey)(nil) // Ensure interface compliance

func Role(roleName string, namespace string) *roleCacheKey {
	var sb strings.Builder

	sb.WriteString(namespace)
	sb.WriteString(CacheKeySeparator)
	sb.WriteString(roleName)

	return &roleCacheKey{
		baseCacheKey{sb.String()},
	}
}

func (k *roleCacheKey) Shard() string {
	return roleCacheName
}
