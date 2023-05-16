package store

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StoreObjectID returns a MongoDB object ID.
// See: https://www.mongodb.com/docs/manual/reference/method/ObjectId/
func StoreObjectID() primitive.ObjectID {
	return primitive.NewObjectIDFromTimestamp(time.Now())
}
