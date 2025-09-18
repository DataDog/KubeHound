package pipeline

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockwriter "github.com/DataDog/KubeHound/pkg/dump/writer/mockwriter"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	discoveryv1 "k8s.io/api/discovery/v1"
)

func TestDumpIngestor_IngestEndpoint(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	// no ingestion
	noIngest := func(t *testing.T, _ []*discoveryv1.EndpointSlice) *EndpointIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewEndpointIngestor(ctx, mDumpWriter)

		return ingestor
	}

	// ingesting n entries
	nIngest := func(t *testing.T, endpoints []*discoveryv1.EndpointSlice) *EndpointIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewEndpointIngestor(ctx, mDumpWriter)

		buffer := make(map[string]*discoveryv1.EndpointSliceList)
		for _, endpoint := range endpoints {
			err := bufferObject[discoveryv1.EndpointSliceList, types.EndpointType](ctx, ingestEndpointPath(endpoint), buffer, endpoint)
			if err != nil {
				t.Fatal(err)
			}
		}

		for path, endpointListNamespaced := range buffer {
			rawBuffer, err := json.Marshal(endpointListNamespaced)
			if err != nil {
				t.Fatalf("failed to marshal Kubernetes object: %v", err)
			}
			mDumpWriter.EXPECT().Write(ctx, rawBuffer, path).Return(nil).Once()
		}

		return ingestor
	}

	type args struct {
		endpoints []*discoveryv1.EndpointSlice
	}
	tests := []struct {
		name     string
		ingestor *EndpointIngestor
		testfct  func(t *testing.T, clusterRole []*discoveryv1.EndpointSlice) *EndpointIngestor
		args     args
		wantErr  bool
	}{
		{
			name:    "no entry",
			testfct: noIngest,
			args: args{
				endpoints: []*discoveryv1.EndpointSlice{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:    "entries found",
			testfct: nIngest,
			args: args{
				endpoints: []*discoveryv1.EndpointSlice{
					collector.FakeEndpoint("name1", "namespace1", []int32{int32(80)}),
					collector.FakeEndpoint("name2", "namespace2", []int32{int32(443)}),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ingestor := tt.testfct(t, tt.args.endpoints)
			for _, endpoint := range tt.args.endpoints {
				if err := ingestor.IngestEndpoint(ctx, endpoint); (err != nil) != tt.wantErr {
					t.Errorf("Dumper.IngestNode() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err := ingestor.Complete(ctx); err != nil {
				t.Errorf("Dumper.IngestNode() error = %v", err)
			}
		})
	}
}
