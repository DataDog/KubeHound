package cachekey

const (
	identityCacheName = "k8s-identity"
)

type identityCacheKey struct {
	identityName string
	namespace    string
}

var _ CacheKey = (*identityCacheKey)(nil) // Ensure interface compliance

func Identity(identityName string, namespace string) *identityCacheKey {
	return &identityCacheKey{
		identityName: identityName,
		namespace:    namespace,
	}
}

func (k *identityCacheKey) Shard() string {
	return identityCacheName
}

func (k *identityCacheKey) Key() string {
	return k.namespace + "##" + k.identityName
}
