package pipeline

import (
	"context"
	"testing"

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
		collector.FakePod("namespace1", "name11", "Running"),
		collector.FakePod("namespace1", "name12", "Running"),
		collector.FakePod("namespace2", "name21", "Running"),
		collector.FakePod("namespace2", "name22", "Running"),
		collector.FakeRole("namespace1", "name11"),
		collector.FakeRole("namespace1", "name12"),
		collector.FakeRole("namespace2", "name21"),
		collector.FakeRole("namespace2", "name22"),
		collector.FakeClusterRole("name1"),
		collector.FakeRoleBinding("namespace1", "name11"),
		collector.FakeRoleBinding("namespace1", "name12"),
		collector.FakeRoleBinding("namespace2", "name21"),
		collector.FakeRoleBinding("namespace2", "name22"),
		collector.FakeClusterRoleBinding("name1"),
		collector.FakeEndpoint("namespace1", "name11", []int32{80}),
		collector.FakeEndpoint("namespace1", "name12", []int32{80}),
		collector.FakeEndpoint("namespace2", "name21", []int32{80}),
		collector.FakeEndpoint("namespace2", "name22", []int32{80}),
	}

	return k8sOjb
}

func newFakeDumpIngestorPipeline(ctx context.Context, t *testing.T, mockCollector bool) (*mockwriter.DumperWriter, collector.CollectorClient) {
	t.Helper()

	mDumpWriter := mockwriter.NewDumperWriter(t)

	mCollectorClient := mockcollector.NewCollectorClient(t)
	clientset := fake.NewSimpleClientset(genK8sObjects()...)
	collectorClient := collector.NewTestK8sAPICollector(ctx, clientset)

	if mockCollector {
		return mDumpWriter, mCollectorClient
	}

	return mDumpWriter, collectorClient

}

func TestPipelineDumpIngestor_Run(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	singleThreadedPipeline := func(t *testing.T) (*mockwriter.DumperWriter, collector.CollectorClient) {
		t.Helper()
		mDumpWriter, collectorClient := newFakeDumpIngestorPipeline(ctx, t, true)
		mCollectorClient, ok := collectorClient.(*mockcollector.CollectorClient)
		if !ok {
			t.Fatalf("failed to cast collector client to mock collector client")
		}

		sequence := dumpIngestorSequence(mCollectorClient, mDumpWriter)

		mDumpWriter.EXPECT().WorkerNumber().Return(1)
		var mStreamNodes, mStreamPods, mStreamRoles, mStreamClusterRoles, mStreamRoleBindings, mStreamClusteRoleBindings *mock.Call

		for _, step := range sequence {
			switch step.entity {
			case tag.EntityNodes:
				mStreamNodes = mCollectorClient.EXPECT().StreamNodes(mock.Anything, NewNodeIngestor(ctx, mDumpWriter)).Return(nil).Once()
			case tag.EntityPods:
				mStreamPods = mCollectorClient.EXPECT().StreamPods(mock.Anything, NewPodIngestor(ctx, mDumpWriter)).Return(nil).Once().NotBefore(mStreamNodes)
			case tag.EntityRoles:
				mStreamRoles = mCollectorClient.EXPECT().StreamRoles(mock.Anything, NewRoleIngestor(ctx, mDumpWriter)).Return(nil).Once().NotBefore(mStreamPods)
			case tag.EntityClusterRoles:
				mStreamClusterRoles = mCollectorClient.EXPECT().StreamClusterRoles(mock.Anything, NewClusterRoleIngestor(ctx, mDumpWriter)).Return(nil).Once().NotBefore(mStreamRoles)
			case tag.EntityRolebindings:
				mStreamRoleBindings = mCollectorClient.EXPECT().StreamRoleBindings(mock.Anything, NewRoleBindingIngestor(ctx, mDumpWriter)).Return(nil).Once().NotBefore(mStreamClusterRoles)
			case tag.EntityClusterRolebindings:
				mStreamClusteRoleBindings = mCollectorClient.EXPECT().StreamClusterRoleBindings(mock.Anything, NewClusterRoleBindingIngestor(ctx, mDumpWriter)).Return(nil).Once().NotBefore(mStreamRoleBindings)
			case tag.EntityEndpoints:
				mCollectorClient.EXPECT().StreamEndpoints(mock.Anything, NewEndpointIngestor(ctx, mDumpWriter)).Return(nil).Once().NotBefore(mStreamClusteRoleBindings)
			}
		}

		return mDumpWriter, mCollectorClient
	}

	multiThreadedPipeline := func(t *testing.T) (*mockwriter.DumperWriter, collector.CollectorClient) {
		t.Helper()
		mDumpWriter, collectorClient := newFakeDumpIngestorPipeline(ctx, t, true)
		mCollectorClient, ok := collectorClient.(*mockcollector.CollectorClient)
		if !ok {
			t.Fatalf("failed to cast collector client to mock collector client")
		}

		sequence := dumpIngestorSequence(mCollectorClient, mDumpWriter)

		mDumpWriter.EXPECT().WorkerNumber().Return(0)

		for _, step := range sequence {
			switch step.entity {
			case tag.EntityNodes:
				mCollectorClient.EXPECT().StreamNodes(mock.Anything, NewNodeIngestor(ctx, mDumpWriter)).Return(nil).Once()
			case tag.EntityPods:
				mCollectorClient.EXPECT().StreamPods(mock.Anything, NewPodIngestor(ctx, mDumpWriter)).Return(nil).Once()
			case tag.EntityRoles:
				mCollectorClient.EXPECT().StreamRoles(mock.Anything, NewRoleIngestor(ctx, mDumpWriter)).Return(nil).Once()
			case tag.EntityClusterRoles:
				mCollectorClient.EXPECT().StreamClusterRoles(mock.Anything, NewClusterRoleIngestor(ctx, mDumpWriter)).Return(nil).Once()
			case tag.EntityRolebindings:
				mCollectorClient.EXPECT().StreamRoleBindings(mock.Anything, NewRoleBindingIngestor(ctx, mDumpWriter)).Return(nil).Once()
			case tag.EntityClusterRolebindings:
				mCollectorClient.EXPECT().StreamClusterRoleBindings(mock.Anything, NewClusterRoleBindingIngestor(ctx, mDumpWriter)).Return(nil).Once()
			case tag.EntityEndpoints:
				mCollectorClient.EXPECT().StreamEndpoints(mock.Anything, NewEndpointIngestor(ctx, mDumpWriter)).Return(nil).Once()
			}
		}

		return mDumpWriter, mCollectorClient
	}

	singleThreadedPipelineLive := func(t *testing.T) (*mockwriter.DumperWriter, collector.CollectorClient) {
		t.Helper()

		return PipelineLiveTest(ctx, t, 1)
	}

	multiThreadedPipelineLive := func(t *testing.T) (*mockwriter.DumperWriter, collector.CollectorClient) {
		t.Helper()

		return PipelineLiveTest(ctx, t, 0)
	}

	tests := []struct {
		name    string
		testfct func(t *testing.T) (*mockwriter.DumperWriter, collector.CollectorClient)
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
			mDumpWriter, mCollectorClient := tt.testfct(t)
			ctx, pipeline, _ := NewPipelineDumpIngestor(ctx, mCollectorClient, mDumpWriter)

			if err := pipeline.Run(ctx); (err != nil) != tt.wantErr {
				t.Errorf("PipelineDumpIngestor.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := pipeline.Wait(ctx); (err != nil) != tt.wantErr {
				t.Errorf("PipelineDumpIngestor.Wait() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
