package dump

import (
	"context"
	"fmt"
	"path"
	"testing"
	"time"

	mockcollector "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	mockwriter "github.com/DataDog/KubeHound/pkg/dump/writer/mockwriter"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/stretchr/testify/mock"
)

func newFakeDumpIngestorPipeline(t *testing.T) (*DumpIngestor, *mockwriter.DumperWriter, *mockcollector.CollectorClient) {
	t.Helper()

	directoryOutput := "/tmp"
	mDumpWriter := mockwriter.NewDumperWriter(t)
	mCollectorClient := mockcollector.NewCollectorClient(t)

	// Generate path for the dump
	clusterName := "test-cluster"
	// ./<clusterName>/kubehound_<clusterName>_<date>
	resName := path.Join(clusterName, fmt.Sprintf("%s%s_%s", OfflineDumpPrefix, clusterName, time.Now().Format(OfflineDumpDateFormat)))

	return &DumpIngestor{
		directoryOutput: directoryOutput,
		collector:       mCollectorClient,
		ClusterName:     clusterName,
		ResName:         resName,
		writer:          mDumpWriter,
	}, mDumpWriter, mCollectorClient
}

func TestPipelineDumpIngestor_Run(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	singleThreadedPipeline := func(t *testing.T) *DumpIngestor {
		t.Helper()
		dumperIngestor, mDumpWriter, mCollectorClient := newFakeDumpIngestorPipeline(t)
		sequence := dumpIngestorSequence(dumperIngestor)

		mDumpWriter.EXPECT().WorkerNumber().Return(1)
		var mStreamNodes, mStreamPods, mStreamRoles, mStreamClusterRoles, mStreamRoleBindings, mStreamClusteRoleBindings *mock.Call

		for _, step := range sequence {
			switch step.entity {
			case tag.EntityNodes:
				mStreamNodes = mCollectorClient.EXPECT().StreamNodes(mock.Anything, dumperIngestor).Return(nil).Once()
			case tag.EntityPods:
				mStreamPods = mCollectorClient.EXPECT().StreamPods(mock.Anything, dumperIngestor).Return(nil).Once().NotBefore(mStreamNodes)
			case tag.EntityRoles:
				mStreamRoles = mCollectorClient.EXPECT().StreamRoles(mock.Anything, dumperIngestor).Return(nil).Once().NotBefore(mStreamPods)
			case tag.EntityClusterRoles:
				mStreamClusterRoles = mCollectorClient.EXPECT().StreamClusterRoles(mock.Anything, dumperIngestor).Return(nil).Once().NotBefore(mStreamRoles)
			case tag.EntityRolebindings:
				mStreamRoleBindings = mCollectorClient.EXPECT().StreamRoleBindings(mock.Anything, dumperIngestor).Return(nil).Once().NotBefore(mStreamClusterRoles)
			case tag.EntityClusterRolebindings:
				mStreamClusteRoleBindings = mCollectorClient.EXPECT().StreamClusterRoleBindings(mock.Anything, dumperIngestor).Return(nil).Once().NotBefore(mStreamRoleBindings)
			case tag.EntityEndpoints:
				mCollectorClient.EXPECT().StreamEndpoints(mock.Anything, dumperIngestor).Return(nil).Once().NotBefore(mStreamClusteRoleBindings)
			}
		}

		return dumperIngestor
	}

	multiThreadedPipeline := func(t *testing.T) *DumpIngestor {
		t.Helper()
		dumperIngestor, mDumpWriter, mCollectorClient := newFakeDumpIngestorPipeline(t)
		sequence := dumpIngestorSequence(dumperIngestor)

		mDumpWriter.EXPECT().WorkerNumber().Return(0)

		for _, step := range sequence {
			switch step.entity {
			case tag.EntityNodes:
				mCollectorClient.EXPECT().StreamNodes(mock.Anything, dumperIngestor).Return(nil).Once()
			case tag.EntityPods:
				mCollectorClient.EXPECT().StreamPods(mock.Anything, dumperIngestor).Return(nil).Once()
			case tag.EntityRoles:
				mCollectorClient.EXPECT().StreamRoles(mock.Anything, dumperIngestor).Return(nil).Once()
			case tag.EntityClusterRoles:
				mCollectorClient.EXPECT().StreamClusterRoles(mock.Anything, dumperIngestor).Return(nil).Once()
			case tag.EntityRolebindings:
				mCollectorClient.EXPECT().StreamRoleBindings(mock.Anything, dumperIngestor).Return(nil).Once()
			case tag.EntityClusterRolebindings:
				mCollectorClient.EXPECT().StreamClusterRoleBindings(mock.Anything, dumperIngestor).Return(nil).Once()
			case tag.EntityEndpoints:
				mCollectorClient.EXPECT().StreamEndpoints(mock.Anything, dumperIngestor).Return(nil).Once()
			}

		}

		return dumperIngestor
	}

	tests := []struct {
		name    string
		testfct func(t *testing.T) *DumpIngestor
		wantErr bool
	}{
		{
			name:    "single threaded pipeline",
			testfct: singleThreadedPipeline,
			wantErr: false,
		},
		{
			name:    "multi threaded pipeline",
			testfct: multiThreadedPipeline,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dumperIngestor := tt.testfct(t)
			ctx, pipeline, _ := newPipelineDumpIngestor(ctx, dumperIngestor)

			if err := pipeline.Run(ctx); (err != nil) != tt.wantErr {
				t.Errorf("PipelineDumpIngestor.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := pipeline.Wait(ctx); (err != nil) != tt.wantErr {
				t.Errorf("PipelineDumpIngestor.Wait() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}
