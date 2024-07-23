package dump

import (
	"context"
	"reflect"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/dump/pipeline"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
	mockwriter "github.com/DataDog/KubeHound/pkg/dump/writer/mockwriter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	mockDirectoryOutput = "/tmp"
)

func TestNewDumpIngestor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	clientset := fake.NewSimpleClientset()
	collectorClient := collector.NewTestK8sAPICollector(ctx, clientset)

	type args struct {
		collectorClient collector.CollectorClient
		compression     bool
		directoryOutput string
		runID           *config.RunID
	}
	tests := []struct {
		name    string
		args    args
		want    *DumpIngestor
		wantErr bool
	}{

		{
			name: "no compression",
			args: args{
				collectorClient: collectorClient,
				compression:     false,
				directoryOutput: mockDirectoryOutput,
				runID:           config.NewRunID(),
			},
			want: &DumpIngestor{
				writer: &writer.FileWriter{},
			},
			wantErr: false,
		},
		{
			name: "compression activated",
			args: args{
				collectorClient: collectorClient,
				compression:     true,
				directoryOutput: mockDirectoryOutput,
				runID:           config.NewRunID(),
			},
			want: &DumpIngestor{
				writer: &writer.TarWriter{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewDumpIngestor(ctx, tt.args.collectorClient, tt.args.compression, tt.args.directoryOutput, tt.args.runID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDumpIngestorsss() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !assert.Equal(t, reflect.TypeOf(got.writer), reflect.TypeOf(tt.want.writer)) {
				t.Errorf("NewDumpIngestor() = %v, want %v", reflect.TypeOf(got.writer), reflect.TypeOf(tt.want.writer))
			}
		})
	}
}

func TestDumpIngestor_DumpK8sObjects(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	writerOutputPath := "/tmp/cluster_name"
	singleThreadedPipelineLive := func(t *testing.T) (*mockwriter.DumperWriter, collector.CollectorClient) {
		t.Helper()
		mDumpWriter, mCollectorClient := pipeline.PipelineLiveTest(ctx, t, 1)
		mDumpWriter.EXPECT().Flush(mock.Anything).Return(nil).Once()
		mDumpWriter.EXPECT().Close(mock.Anything).Return(nil).Once()
		mDumpWriter.EXPECT().OutputPath().Return(writerOutputPath).Once()

		return mDumpWriter, mCollectorClient
	}

	multiThreadedPipelineLive := func(t *testing.T) (*mockwriter.DumperWriter, collector.CollectorClient) {
		t.Helper()
		mDumpWriter, mCollectorClient := pipeline.PipelineLiveTest(ctx, t, 0)
		mDumpWriter.EXPECT().Flush(mock.Anything).Return(nil).Once()
		mDumpWriter.EXPECT().Close(mock.Anything).Return(nil).Once()
		mDumpWriter.EXPECT().OutputPath().Return(writerOutputPath).Once()

		return mDumpWriter, mCollectorClient
	}

	tests := []struct {
		name    string
		testfct func(t *testing.T) (*mockwriter.DumperWriter, collector.CollectorClient)
		wantErr bool
	}{
		{
			name:    "single thread dump",
			testfct: singleThreadedPipelineLive,
			wantErr: false,
		},
		{
			name:    "multi-threaded dump entry",
			testfct: multiThreadedPipelineLive,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mDumpWriter, mCollectorClient := tt.testfct(t)
			d := &DumpIngestor{
				collector: mCollectorClient,
				writer:    mDumpWriter,
			}
			if err := d.DumpK8sObjects(ctx); (err != nil) != tt.wantErr {
				t.Errorf("DumpIngestor.DumpK8sObjects() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := d.Close(ctx); err != nil {
				t.Errorf("DumpIngestor.DumpK8sObjects() error = %v", err)
			}
			if d.OutputPath() != writerOutputPath {
				t.Errorf("DumpIngestor.OutputPath() error = %v", d.OutputPath())
			}
		})
	}
}
