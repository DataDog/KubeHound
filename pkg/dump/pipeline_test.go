package dump

import (
	"context"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockcollector "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	mockwriter "github.com/DataDog/KubeHound/pkg/dump/writer/mockwriter"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func genK8sObjects() []runtime.Object {
	k8sOjb := []runtime.Object{
		collector.FakeNode("node1", "provider1"),
		collector.FakePod("pod1", "namespace1", "Running"),
		collector.FakeRole("namespace1", "name1"),
		collector.FakeClusterRole("name1"),
		collector.FakeRoleBinding("namespace1", "name1"),
		collector.FakeClusterRoleBinding("name1"),
		collector.FakeEndpoint("namespace1", "name1", []int32{80}),
	}

	return k8sOjb
}

func newFakeDumpIngestorPipeline(t *testing.T, mockCollector bool) (*DumpIngestor, *mockwriter.DumperWriter, *mockcollector.CollectorClient) {
	t.Helper()

	mDumpWriter := mockwriter.NewDumperWriter(t)

	mCollectorClient := mockcollector.NewCollectorClient(t)
	clientset := fake.NewSimpleClientset(genK8sObjects()...)
	collectorClient := collector.NewTestK8sAPICollector(context.TODO(), clientset)

	// Generate path for the dump
	clusterName := "test-cluster"
	// ./<clusterName>/kubehound_<clusterName>_<date>
	resName := path.Join(clusterName, fmt.Sprintf("%s%s_%s", OfflineDumpPrefix, clusterName, time.Now().Format(OfflineDumpDateFormat)))

	if mockCollector {
		return &DumpIngestor{
			directoryOutput: mockDirectoryOutput,
			collector:       mCollectorClient,
			ClusterName:     clusterName,
			ResName:         resName,
			writer:          mDumpWriter,
		}, mDumpWriter, mCollectorClient
	}

	return &DumpIngestor{
		directoryOutput: mockDirectoryOutput,
		collector:       collectorClient,
		ClusterName:     clusterName,
		ResName:         resName,
		writer:          mDumpWriter,
	}, mDumpWriter, mCollectorClient

}

func liveTest(t *testing.T, workerNum int) *DumpIngestor {
	t.Helper()
	dumperIngestor, mDumpWriter, _ := newFakeDumpIngestorPipeline(t, false)

	mDumpWriter.EXPECT().WorkerNumber().Return(workerNum)

	for range genK8sObjects() {
		mDumpWriter.EXPECT().Write(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		mDumpWriter.EXPECT().Flush(mock.Anything).Return(nil).Once()
	}

	mFlush := mDumpWriter.EXPECT().Flush(mock.Anything).Return(nil).Once()
	mDumpWriter.EXPECT().Close(mock.Anything).Return(nil).Once().NotBefore(mFlush)

	return dumperIngestor
}

func TestPipelineDumpIngestor_Run(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	singleThreadedPipeline := func(t *testing.T) *DumpIngestor {
		t.Helper()
		dumperIngestor, mDumpWriter, mCollectorClient := newFakeDumpIngestorPipeline(t, true)
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

		mFlush := mDumpWriter.EXPECT().Flush(mock.Anything).Return(nil).Once()
		mDumpWriter.EXPECT().Close(mock.Anything).Return(nil).Once().NotBefore(mFlush)

		return dumperIngestor
	}

	multiThreadedPipeline := func(t *testing.T) *DumpIngestor {
		t.Helper()
		dumperIngestor, mDumpWriter, mCollectorClient := newFakeDumpIngestorPipeline(t, true)
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

		mFlush := mDumpWriter.EXPECT().Flush(mock.Anything).Return(nil).Once()
		mDumpWriter.EXPECT().Close(mock.Anything).Return(nil).Once().NotBefore(mFlush)

		return dumperIngestor
	}

	singleThreadedPipelineLive := func(t *testing.T) *DumpIngestor {
		t.Helper()

		return liveTest(t, 1)
	}

	multiThreadedPipelineLive := func(t *testing.T) *DumpIngestor {
		t.Helper()

		return liveTest(t, 0)
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
		{
			name:    "single threaded pipeline live",
			testfct: singleThreadedPipelineLive,
			wantErr: false,
		},
		{
			name:    "multi threaded pipeline live",
			testfct: multiThreadedPipelineLive,
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

			if err := dumperIngestor.Close(ctx); (err != nil) != tt.wantErr {
				t.Errorf("dumperIngestor.Close(ctx) error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}
