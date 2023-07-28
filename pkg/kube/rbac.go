package kube

import (
	"context"
	"fmt"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	DefaultNodeGroup     = "system:nodes"
	DefaultNodeNamespace = ""
)

var (
	defaultOnce sync.Once
	defaultNid  primitive.ObjectID
	defaultErr  error
)

// NodeUser will return the full name of the dedicated node user.
// See reference for details: https://kubernetes.io/docs/reference/access-authn-authz/node/
func NodeUser(nodeName string) string {
	return fmt.Sprintf("system:node:%s", nodeName)
}

// DefaultNodeIdentity will return the store id of the default system:nodes group.
// See reference for details: https://kubernetes.io/docs/reference/access-authn-authz/node/.
func DefaultNodeIdentity(ctx context.Context, c cache.CacheReader) (primitive.ObjectID, error) {
	defaultOnce.Do(func() {
		var err error
		ck := cachekey.Identity(DefaultNodeGroup, DefaultNodeNamespace)
		defaultNid, err = c.Get(ctx, ck).ObjectID()
		if err != nil {
			defaultErr = fmt.Errorf("resolving default node user")
		}
	})

	return defaultNid, defaultErr
}

// NodeIdentity will either return the store id of the dedicated node user or store id of the default system:nodes group if a dedicated user is not present.
// See reference for details: https://kubernetes.io/docs/reference/access-authn-authz/node/.
func NodeIdentity(ctx context.Context, c cache.CacheReader, nodeName string) (primitive.ObjectID, error) {
	// Lookup whether the node has a dedicated user
	ck := cachekey.Identity(NodeUser(nodeName), DefaultNodeNamespace)
	nid, err := c.Get(ctx, ck).ObjectID()
	switch err {
	case nil:
		// We have a dedicated user, return its id
		return nid, nil
	case cache.ErrNoEntry:
		// Return the default user id
		return DefaultNodeIdentity(ctx, c)
	}

	return primitive.NilObjectID, fmt.Errorf("resolving node identity (%s): %w", nodeName, err)
}
