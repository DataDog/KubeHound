package edge

// import (
// 	"context"
// 	"fmt"

// 	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
// 	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
// 	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
// 	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
// 	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
// 	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// func init() {
// 	Register(&EndpointExposePublic{}, RegisterDefault)
// }

// type EndpointExposePublic struct {
// 	BaseEdge
// }

// type endpointExposeGroup struct {
// 	Endpoint  primitive.ObjectID `bson:"_id" json:"endpoint_id"`
// 	Container primitive.ObjectID `bson:"container_id" json:"container_id"`
// }

// func (e *EndpointExposePublic) Label() string {
// 	return "ENDPOINT_EXPOSE"
// }

// func (e *EndpointExposePublic) Name() string {
// 	return "EndpointExposePublic"
// }

// func (e *EndpointExposePublic) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
// 	typed, ok := entry.(*endpointExposeGroup)
// 	if !ok {
// 		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
// 	}

// 	return adapter.GremlinEdgeProcessor(ctx, oic, e.Label(), typed.Endpoint, typed.Container)
// }

// func (e *EndpointExposePublic) Stream(ctx context.Context, store storedb.Provider, c cache.CacheReader,
// 	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

// 	endpoints := adapter.MongoDB(store).Collection(collections.EndpointName)

// 	// Find endpoints with has_slice
// 	// Find the container with same pod name / namespace / port
// 	// return container id
// 	pipeline := []bson.M{
// 		// Match element with an associated slice
// 		{
// 			"$match": bson.M{
// 				"has_slice": true,
// 			},
// 		},
// 		// Retrieve the target pod name and namespace as well as the container port
// 		{
// 			"$lookup": bson.M{
// 				"as":   "matchContainers",
// 				"from": "containers",
// 				"let": bson.M{
// 					"pod":   "$pod_name",
// 					"podNS": "$pod_namespace",
// 					"port":  "$port.port",
// 					"proto": "$port.protocol",
// 				},
// 				"pipeline": []bson.M{
// 					// Find the container matching our specification
// 					{
// 						"$match": bson.M{"$or": bson.A{
// 							bson.M{"$expr": bson.M{
// 								"$eq": bson.A{
// 									"$k8.objectmeta.namespace", "$$roleNamespace",
// 								},
// 							}},
// 							bson.M{"is_namespaced": false},
// 						}},
// 					},
// 					{
// 						"$project": bson.M{
// 							"_id": 1,
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			"$unwind": "$matchContainers",
// 		},
// 		{
// 			"$project": bson.M{$e
// 				"_id":          1,
// 				"container_id": "$matchContainers._id",
// 			},
// 		},
// 	}

// 	cur, err := endpoints.Aggregate(context.Background(), pipeline)
// 	if err != nil {
// 		return err
// 	}
// 	defer cur.Close(ctx)

// 	return adapter.MongoCursorHandler[endpointExposeGroup](ctx, cur, callback, complete)
// }
