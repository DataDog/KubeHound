package cache

import (
	"errors"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCacheResult_ObjectID(t *testing.T) {
	objectId := primitive.NewObjectID()
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
		{
			name: "success case",
			fields: fields{
				Value: objectId.Hex(),
				Err:   nil,
			},
			want:    objectId,
			wantErr: false,
		},
		{
			name: "error result case",
			fields: fields{
				Value: primitive.NilObjectID,
				Err:   errors.New("test error"),
			},
			want:    primitive.NilObjectID,
			wantErr: true,
		},
		{
			name: "type error case",
			fields: fields{
				Value: -1,
				Err:   nil,
			},
			want:    primitive.NilObjectID,
			wantErr: true,
		},
		{
			name: "nil value case",
			fields: fields{
				Value: nil,
				Err:   nil,
			},
			want:    primitive.NilObjectID,
			wantErr: true,
		},
		{
			name: "object ID convert test",
			fields: fields{
				Value: "notandobjid",
				Err:   nil,
			},
			want:    primitive.NilObjectID,
			wantErr: true,
		},
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
		{
			name: "success case",
			fields: fields{
				Value: int64(1337),
				Err:   nil,
			},
			want:    1337,
			wantErr: false,
		},
		{
			name: "error result case",
			fields: fields{
				Value: int64(-1),
				Err:   errors.New("test error"),
			},
			want:    int64(-1),
			wantErr: true,
		},
		{
			name: "type error case",
			fields: fields{
				Value: "hello",
				Err:   nil,
			},
			want:    int64(-1),
			wantErr: true,
		},
		{
			name: "nil value case",
			fields: fields{
				Value: nil,
				Err:   nil,
			},
			want:    int64(-1),
			wantErr: true,
		},
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
		{
			name: "success case",
			fields: fields{
				Value: "hello world",
				Err:   nil,
			},
			want:    "hello world",
			wantErr: false,
		},
		{
			name: "error result case",
			fields: fields{
				Value: "",
				Err:   errors.New("test error"),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "type error case",
			fields: fields{
				Value: -1,
				Err:   nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "nil value case",
			fields: fields{
				Value: nil,
				Err:   nil,
			},
			want:    "",
			wantErr: true,
		},
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

func TestCacheResult_Bool(t *testing.T) {
	type fields struct {
		Value any
		Err   error
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{
			name: "success case",
			fields: fields{
				Value: true,
				Err:   nil,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "error result case",
			fields: fields{
				Value: false,
				Err:   errors.New("test error"),
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "type error case",
			fields: fields{
				Value: -1,
				Err:   nil,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "nil value case",
			fields: fields{
				Value: nil,
				Err:   nil,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CacheResult{
				Value: tt.fields.Value,
				Err:   tt.fields.Err,
			}
			got, err := r.Bool()
			if (err != nil) != tt.wantErr {
				t.Errorf("CacheResult.Bool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CacheResult.Bool() = %v, want %v", got, tt.want)
			}
		})
	}
}
