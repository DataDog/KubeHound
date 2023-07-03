package cache

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrNoEntry     = errors.New("no matching cache entry")
	ErrInvalidType = errors.New("cache entry value cannot be converted to requested type")
)

type CacheResult struct {
	value any
	err   error
}

func (r *CacheResult) Text() (string, error) {
	if r.err != nil {
		return "", r.err
	}

	if r.value == nil {
		return "", ErrNoEntry
	}

	s, ok := r.value.(string)
	if !ok {
		return "", ErrInvalidType
	}

	return s, nil
}

func (r *CacheResult) Int64() (int64, error) {
	if r.err != nil {
		return -1, r.err
	}

	if r.value == nil {
		return -1, ErrNoEntry
	}

	i, ok := r.value.(int64)
	if !ok {
		return -1, ErrInvalidType
	}

	return i, nil
}

func (r *CacheResult) ObjectID() (primitive.ObjectID, error) {
	if r.err != nil {
		return primitive.NilObjectID, r.err
	}

	if r.value == nil {
		return primitive.NilObjectID, ErrNoEntry
	}

	raw, ok := r.value.(string)
	if !ok {
		return primitive.NilObjectID, ErrInvalidType
	}

	oid, err := primitive.ObjectIDFromHex(raw)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return oid, nil
}
