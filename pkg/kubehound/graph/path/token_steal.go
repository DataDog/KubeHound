package path

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2891284481/TOKEN+STEAL
	tokenStealLabel = "TOKEN_STEAL"

	// @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2880373371/IDENTITY+ASSUME
	identityAssumeLabel = "IDENTITY_ASSUME"
)

var (
	tokenVertexLabel    = vertex.Token{}.Label()
	volumeVertexLabel   = vertex.Volume{}.Label()
	identityVertexLabel = vertex.Identity{}.Label()
	tokenStealPathLabel = fmt.Sprintf("(%s)-[%s]->(%s)-[%s]->(%s)",
		volumeVertexLabel, tokenStealLabel, tokenVertexLabel, identityAssumeLabel, identityVertexLabel)
)

func init() {
	Register(TokenSteal{})
}

type TokenStealPath struct {
	Vertex     *graph.Token `bson:"vertex" json:"vertex"`
	VolumeId   string       `bson:"volume" json:"volume"`
	IdentityId string       `bson:"identity" json:"identity"`
}

type TokenSteal struct {
}

func (v TokenSteal) Label() string {
	return tokenStealLabel
}

func (v TokenSteal) BatchSize() int {
	return DefaultBatchSize
}

func (v TokenSteal) Traversal() PathTraversal {
	return func(g *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		// TODO create the token vertex
		// TODO create a TOKEN_STEAL edge between the volume and the new token
		// TODO create an IDENTITY_ASSUME edge between the token and the associated identity
		return nil
	}
}

func (v TokenSteal) Stream(ctx context.Context, sdb storedb.Provider, cache cache.CacheReader,
	process types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	// Find all volumes with projected service account tokens. The mounts and source fields we need to match on a projected
	// service account token are all deeply nested arrays so matching on the naming convention is the simplest/fastest match
	volumes := adapter.MongoDB(sdb).Collection(collections.VolumeName)
	filter := bson.M{
		"source.volumesource.projected": bson.M{"$exists": true, "$ne": "null"},
		"source.name":                   bson.M{"$regex": "/^kube-api-access/"},
	}

	cur, err := volumes.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	convert := converter.NewGraph()
	var vol store.Volume
	for cur.Next(ctx) {
		err := cur.Decode(&vol)
		if err != nil {
			return err
		}

		// Retrieve the service account name for the pod from the cache
		sa, err := cache.Get(ctx, cachekey.PodIdentity(vol.PodId.Hex()))
		if err != nil {
			log.Trace(ctx).Errorf("cache miss pod identity: %v", err)
			continue
		}

		// Retrieve the associated identity store ID from the cache
		TODO how TF to get the namespc
		said, err := cache.Get(ctx, cachekey.Identity(sa, vol.Name))
		if err != nil {
			// This is completely fine. Most pods will run under a default account with no permissions which we treat
			// as having no identity. As such we do not want to create a token vertex here!
			continue
		}

		// Convert to our graph vertex representation
		v, err := convert.Token(sa, &vol)
		if err != nil {
			return err
		}

		// Create the container that holds all the data required by the traversal function
		err = process(ctx, &TokenStealPath{
			Vertex:     v,
			VolumeId:   vol.Id.Hex(),
			IdentityId: said,
		})
		if err != nil {
			return err
		}
	}

	return complete(ctx)
}
