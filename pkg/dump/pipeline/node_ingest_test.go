package pipeline

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockwriter "github.com/DataDog/KubeHound/pkg/dump/writer/mockwriter"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
)

func TestDumpIngestor_IngestNode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// no ingestion
	noIngest := func(t *testing.T, _ []types.NodeType) *NodeIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewNodeIngestor(ctx, mDumpWriter)

		return ingestor
	}

	// ingesting n entries
	nIngest := func(t *testing.T, nodes []types.NodeType) *NodeIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewNodeIngestor(ctx, mDumpWriter)

		buffer := &corev1.NodeList{}

		for _, node := range nodes {
			buffer.Items = append(buffer.Items, *node)
		}
		mDumpWriter.EXPECT().Write(mock.Anything, buffer, collector.NodePath).Return(nil).Once()

		return ingestor
	}

	type args struct {
		nodes []types.NodeType
	}
	tests := []struct {
		name           string
		dumperIngestor *ClusterRoleIngestor
		testfct        func(t *testing.T, clusterRole []types.NodeType) *NodeIngestor
		args           args
		wantErr        bool
	}{
		{
			name:    "no entry",
			testfct: noIngest,
			args: args{
				nodes: []types.NodeType{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:    "entries found",
			testfct: nIngest,
			args: args{
				nodes: []types.NodeType{
					collector.FakeNode("name1", "provider1"),
					collector.FakeNode("name2", "provider2"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ingestor := tt.testfct(t, tt.args.nodes)
			for _, node := range tt.args.nodes {
				if err := ingestor.IngestNode(ctx, node); (err != nil) != tt.wantErr {
					t.Errorf("Dumper.IngestNode() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err := ingestor.Complete(ctx); err != nil {
				t.Errorf("Dumper.IngestNode() error = %v", err)
			}
		})
	}
}
