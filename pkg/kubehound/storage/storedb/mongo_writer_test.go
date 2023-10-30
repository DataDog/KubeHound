//nolint:containedctx,contextcheck
package storedb

import (
	"context"
	"testing"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
)

// We need a "complex" object to store in MongoDB
type FakeElement struct {
	FieldA int
	FieldB string
}

func TestMongoAsyncWriter_Queue(t *testing.T) {
	t.Parallel()

	// FIXME: we should probably setup a mongodb test server in CI for the system tests
	if config.IsCI() {
		t.Skip("Skip mongo tests in CI")
	}

	fakeElem := FakeElement{
		FieldA: 123,
		FieldB: "lol",
	}

	ctx := context.Background()
	mongoProvider, err := NewMongoProvider(ctx, MongoLocalDatabaseURL, 1*time.Second)

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
			writer := NewMongoAsyncWriter(ctx, mongoProvider, collections.FakeCollection{}, WithTags([]string{tag.Storage("mongotest")}))
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

	// FIXME: we should probably setup a mongodb test server in CI for the system tests
	if config.IsCI() {
		t.Skip("Skip mongo tests in CI")
	}
	fakeElem := FakeElement{
		FieldA: 123,
		FieldB: "lol",
	}

	type argsQueue struct {
		ctx   context.Context
		model any
	}
	type argsFlush struct {
		ctx context.Context
	}

	tests := []struct {
		name      string
		argsQueue []argsQueue
		argsFlush argsFlush
		want      chan struct{}
		queueSize int
		wantErr   bool
	}{
		{
			name: "test flushing 2 items from mongo db queue",
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
			name:      "test flushing 0 items from mongo db queue",
			argsQueue: []argsQueue{},
			argsFlush: argsFlush{
				ctx: context.Background(),
			},
			queueSize: 0,
			wantErr:   false,
		},
		{
			name: "test flushing 6 items from mongo db queue (with TestBatchSize = 5)",
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
			mongoProvider, err := NewMongoProvider(ctx, MongoLocalDatabaseURL, 1*time.Second)
			// TODO: add another check (env var maybe?)
			// "integration test checks"
			if err != nil {
				t.Error("FAILED TO CONNECT TO LOCAL MONGO DB DURING TESTS, SKIPPING")

				return
			}
			defer mongoProvider.Close(ctx)

			maw := NewMongoAsyncWriter(ctx, mongoProvider, collections.FakeCollection{})
			// insert multiple times if needed
			for _, args := range tt.argsQueue {
				if err := maw.Queue(args.ctx, args.model); (err != nil) != tt.wantErr {
					t.Errorf("MongoAsyncWriter.Queue() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			// blocking function
			err = maw.Flush(tt.argsFlush.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("MongoAsyncWriter.Flush() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			// Should now be reset to 0
			gotSize := len(maw.ops)
			if gotSize != tt.queueSize {
				t.Errorf("MongoAsyncWriter.Flush() didn't flushed all items, got size: %d, wanted: %d", gotSize, tt.queueSize)
			}
		})
	}
}
