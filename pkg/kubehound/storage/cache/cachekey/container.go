package cachekey

import (
	"strings"
)

const (
	containerCacheName = "k8s-container"
)

type containerCacheKey struct {
	baseCacheKey
}

var _ CacheKey = (*containerCacheKey)(nil) // Ensure interface compliance

func Container(podName string, containerName string, namespace string) *containerCacheKey {
	var sb strings.Builder

	sb.WriteString(namespace)
	sb.WriteString(CacheKeySeparator)
	sb.WriteString(podName)
	sb.WriteString(CacheKeySeparator)
	sb.WriteString(containerName)

	return &containerCacheKey{
		baseCacheKey{sb.String()},
	}
}

func (k *containerCacheKey) Shard() string {
	return containerCacheName
}
