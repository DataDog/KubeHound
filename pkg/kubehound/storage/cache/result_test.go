package cache

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCacheResult_ObjectID(t *testing.T) {
	type fields struct {
		Value any
		Err   error
	}
	tests := []struct {
		name    string
		fields  fields
		want    primitive.ObjectID
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CacheResult{
				Value: tt.fields.Value,
				Err:   tt.fields.Err,
			}
			got, err := r.ObjectID()
			if (err != nil) != tt.wantErr {
				t.Errorf("CacheResult.ObjectID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CacheResult.ObjectID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacheResult_Int64(t *testing.T) {
	type fields struct {
		Value any
		Err   error
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CacheResult{
				Value: tt.fields.Value,
				Err:   tt.fields.Err,
			}
			got, err := r.Int64()
			if (err != nil) != tt.wantErr {
				t.Errorf("CacheResult.Int64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CacheResult.Int64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacheResult_Text(t *testing.T) {
	type fields struct {
		Value any
		Err   error
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CacheResult{
				Value: tt.fields.Value,
				Err:   tt.fields.Err,
			}
			got, err := r.Text()
			if (err != nil) != tt.wantErr {
				t.Errorf("CacheResult.Text() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CacheResult.Text() = %v, want %v", got, tt.want)
			}
		})
	}
}
