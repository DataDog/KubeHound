package cachekey

import "strconv"

const (
	endpointCacheName = "k8s-endpoint"
)

type endpointCacheKey struct {
	podName   string
	namespace string
	port      string
	protocol  string
}

var _ CacheKey = (*endpointCacheKey)(nil) // Ensure interface compliance

func Endpoint(podName string, port int, protocol string, namespace string) *endpointCacheKey {
	return &endpointCacheKey{
		podName:   podName,
		namespace: namespace,
		port:      strconv.Itoa(port),
		protocol:  protocol,
	}
}

func (k *endpointCacheKey) Shard() string {
	return endpointCacheName
}

func (k *endpointCacheKey) Key() string {
	return k.namespace + "##" + k.podName + "##" + k.port + "##" + k.protocol
}
