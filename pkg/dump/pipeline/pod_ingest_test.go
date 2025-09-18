package pipeline

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockwriter "github.com/DataDog/KubeHound/pkg/dump/writer/mockwriter"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	corev1 "k8s.io/api/core/v1"
)

func TestDumpIngestor_IngestPod(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	// no ingestion
	noIngest := func(t *testing.T, _ []types.PodType) *PodIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewPodIngestor(ctx, mDumpWriter)

		return ingestor
	}

	// ingesting n entries
	nIngest := func(t *testing.T, pods []types.PodType) *PodIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewPodIngestor(ctx, mDumpWriter)

		buffer := make(map[string]*corev1.PodList)
		for _, pod := range pods {
			err := bufferObject[corev1.PodList, types.PodType](ctx, ingestPodPath(pod), buffer, pod)
			if err != nil {
				t.Fatal(err)
			}
		}

		for path, podListNamespaced := range buffer {
			rawBuffer, err := json.Marshal(podListNamespaced)
			if err != nil {
				t.Fatalf("failed to marshal Kubernetes object: %v", err)
			}
			mDumpWriter.EXPECT().Write(ctx, rawBuffer, path).Return(nil).Once()
		}

		return ingestor
	}

	type args struct {
		pods []types.PodType
	}
	tests := []struct {
		name     string
		ingestor *PodIngestor
		testfct  func(t *testing.T, clusterRole []types.PodType) *PodIngestor
		args     args
		wantErr  bool
	}{
		{
			name:    "no entry",
			testfct: noIngest,
			args: args{
				pods: []types.PodType{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:    "entries found",
			testfct: nIngest,
			args: args{
				pods: []types.PodType{
					collector.FakePod("name1", "image1", "Running"),
					collector.FakePod("name2", "image2", "Running"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ingestor := tt.testfct(t, tt.args.pods)
			for _, pod := range tt.args.pods {
				if err := ingestor.IngestPod(ctx, pod); (err != nil) != tt.wantErr {
					t.Errorf("ingestor.IngestPod() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err := ingestor.Complete(ctx); err != nil {
				t.Errorf("ingestor.Complete() error = %v", err)
			}
		})
	}
}
