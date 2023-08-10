package cachekey

import (
	"strconv"
	"strings"
)

const (
	endpointCacheName = "k8s-endpoint"
)

type endpointCacheKey struct {
	baseCacheKey
}

var _ CacheKey = (*endpointCacheKey)(nil) // Ensure interface compliance

func Endpoint(namespace string, podName string, protocol string, port int) *endpointCacheKey {
	var sb strings.Builder

	sb.WriteString(namespace)
	sb.WriteString(CacheKeySeparator)
	sb.WriteString(podName)
	sb.WriteString(CacheKeySeparator)
	sb.WriteString(protocol)
	sb.WriteString(CacheKeySeparator)
	sb.WriteString(strconv.Itoa(port))

	return &endpointCacheKey{
		baseCacheKey{sb.String()},
	}
}

func (k *endpointCacheKey) Shard() string {
	return endpointCacheName
}
