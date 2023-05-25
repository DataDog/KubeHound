package storedb

import (
	"context"
	"testing"
	"time"

	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"go.mongodb.org/mongo-driver/mongo"
)

// We need a "complex" object to store in MongoDB
type FakeElement struct {
	FieldA int
	FieldB string
}

func TestMongoAsyncWriter_Queue(t *testing.T) {
	t.Parallel()
	fakeElem := FakeElement{
		FieldA: 123,
		FieldB: "lol",
	}

	ctx := context.Background()
	mongoProvider, err := NewMongoProvider(ctx, MongoDatabaseURL, 1*time.Second)

	// TODO: add another check (env var maybe?)
	// "integration test checks"
	if err != nil {
		t.Error("FAILED TO CONNECT TO LOCAL MONGO DB DURING TESTS, SKIPPING")
		return
	}

	type args struct {
		ctx   context.Context
		model any
	}
	tests := []struct {
		name      string
		args      []args
		wantErr   bool
		queueSize int
	}{
		{
			name:      "test adding one item in mongo db queue",
			args:      []args{},
			wantErr:   false,
			queueSize: 0,
		},
		{
			name: "test adding one item in mongo db queue",
			args: []args{
				{
					ctx:   context.TODO(),
					model: fakeElem,
				},
			},
			wantErr:   false,
			queueSize: 1,
		},
		{
			name: "test adding multiple items in mongo db queue",
			args: []args{
				{
					ctx:   context.TODO(),
					model: fakeElem,
				},
				{
					ctx:   context.TODO(),
					model: "test random string insert 2",
				},
			},
			wantErr:   false,
			queueSize: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			writer := NewMongoAsyncWriter(ctx, mongoProvider, collections.FakeCollection{})
			// insert multiple times if needed
			for _, args := range tt.args {
				if err := writer.Queue(args.ctx, args.model); (err != nil) != tt.wantErr {
					t.Errorf("MongoAsyncWriter.Queue() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			// We can't really know when the queue as effectively enqueued all the items
			// 100 ms should be a very large buffer
			// all these tests are running in parallel, so it should be mostly always end in ~100ms
			// (depending on the core count / parallel count)
			time.Sleep(100 * time.Millisecond)

			gotSize := len(writer.ops)
			if gotSize != tt.queueSize {
				t.Errorf("MongoAsyncWriter.Queue() didn't inserted items, got size: %d, wanted: %d", gotSize, tt.queueSize)
			}
		})
	}
}

func TestMongoAsyncWriter_Flush(t *testing.T) {
	t.Parallel()
	fakeElem := FakeElement{
		FieldA: 123,
		FieldB: "lol",
	}

	type fields struct {
		mongodb *MongoProvider
		ops     []mongo.WriteModel
	}
	type argsQueue struct {
		ctx   context.Context
		model any
	}
	type argsFlush struct {
		ctx context.Context
	}

	ctx := context.Background()
	mongoProvider, err := NewMongoProvider(ctx, MongoDatabaseURL, 1*time.Second)
	// TODO: add another check (env var maybe?)
	// "integration test checks"
	if err != nil {
		t.Error("FAILED TO CONNECT TO LOCAL MONGO DB DURING TESTS, SKIPPING")
		return
	}

	tests := []struct {
		name      string
		fields    fields
		argsQueue []argsQueue
		argsFlush argsFlush
		want      chan struct{}
		queueSize int
		wantErr   bool
	}{
		{
			name: "test flushing multiple items from mongo db queue",
			fields: fields{
				mongodb: mongoProvider,
				ops:     []mongo.WriteModel{},
			},
			argsQueue: []argsQueue{
				{
					ctx:   context.Background(),
					model: fakeElem,
				},
				{
					ctx:   context.Background(),
					model: fakeElem,
				},
			},
			argsFlush: argsFlush{
				ctx: context.Background(),
			},
			queueSize: 0,
			wantErr:   false,
		},
		{
			name: "test flushing 0 items from mongo db queue",
			fields: fields{
				mongodb: mongoProvider,
				ops:     make([]mongo.WriteModel, 0),
			},
			argsQueue: []argsQueue{},
			argsFlush: argsFlush{
				ctx: context.Background(),
			},
			queueSize: 0,
			wantErr:   false,
		},
		{
			name: "test flushing 6 items from mongo db queue (with TestBatchSize = 5)",
			fields: fields{
				mongodb: mongoProvider,
				ops:     make([]mongo.WriteModel, 0),
			},
			argsQueue: []argsQueue{
				{
					ctx:   context.Background(),
					model: fakeElem,
				},
				{
					ctx:   context.Background(),
					model: fakeElem,
				},
				{
					ctx:   context.Background(),
					model: fakeElem,
				},
				{
					ctx:   context.Background(),
					model: fakeElem,
				},
				{
					ctx:   context.Background(),
					model: fakeElem,
				},
				{
					ctx:   context.Background(),
					model: fakeElem,
				},
			},
			argsFlush: argsFlush{
				ctx: context.Background(),
			},
			queueSize: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			maw := NewMongoAsyncWriter(ctx, tt.fields.mongodb, collections.FakeCollection{})
			// insert multiple times if needed
			for _, args := range tt.argsQueue {
				if err := maw.Queue(args.ctx, args.model); (err != nil) != tt.wantErr {
					t.Errorf("MongoAsyncWriter.Queue() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			// non blocking function
			waiting, err := maw.Flush(tt.argsFlush.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("MongoAsyncWriter.Flush() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// For our test case we want to wait that the flush as done everything it was required to do
			<-waiting

			// Should now be reset to 0
			gotSize := len(maw.ops)
			if gotSize != tt.queueSize {
				t.Errorf("MongoAsyncWriter.Flush() didn't flushed all items, got size: %d, wanted: %d", gotSize, tt.queueSize)
			}
		})
	}
}
