package store

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Identity struct {
	Id           primitive.ObjectID `bson:"_id"`
	Name         string             `bson:"name"`
	IsNamespaced bool               `bson:"is_namespaced"`
	Namespace    string             `bson:"namespace"`
	Type         string             `bson:"type"`
	Ownership    OwnershipInfo      `bson:"ownership"`
}
