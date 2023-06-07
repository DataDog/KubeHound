package cachekey

const (
	identityCacheName = "k8s-identity"
)

type identityCacheKey struct {
	identityName string
}

var _ CacheKey = (*identityCacheKey)(nil) // Ensure interface compliance

func Identity(identityName string) *identityCacheKey {
	return &identityCacheKey{
		identityName: identityName,
	}
}

func (k *identityCacheKey) Namespace() string {
	return identityCacheName
}

func (k *identityCacheKey) Key() string {
	return k.identityName
}
