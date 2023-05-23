package storedb

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMongoProvider_BulkWriter(t *testing.T) {
	provider, err := NewMongoProvider(MongoDatabaseURL)
	// TODO: add another check (env var maybe?)
	// "integration test checks"
	if err != nil {
		t.Error("FAILED TO CONNECT TO LOCAL MONGO DB DURING TESTS, SKIPPING")
	}

	fakeCollection := collections.FakeCollection{}

	type fields struct {
		client     *mongo.Client
		db         *mongo.Database
		collection *mongo.Collection
	}
	type args struct {
		ctx        context.Context
		collection collections.Collection
		opts       []WriterOption
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Bulk writer test with valid collection",
			fields: fields{
				client: provider.client,
				db:     provider.db,
			},
			args: args{
				ctx:        context.Background(),
				collection: fakeCollection,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MongoProvider{
				client: tt.fields.client,
				db:     tt.fields.db,
			}
			writer, err := mp.BulkWriter(tt.args.ctx, tt.args.collection, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("MongoProvider.BulkWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if writer == nil {
				t.Errorf("writer returned by BulkWriter is nil")
			}
		})
	}
}
