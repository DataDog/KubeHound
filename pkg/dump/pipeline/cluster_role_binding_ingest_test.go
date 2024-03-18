package pipeline

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockwriter "github.com/DataDog/KubeHound/pkg/dump/writer/mockwriter"
	"github.com/stretchr/testify/mock"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestDumpIngestor_IngestClusterRoleBinding(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// no ingestion
	noIngest := func(t *testing.T, _ []*rbacv1.ClusterRoleBinding) *ClusterRoleBindingIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewClusterRoleBindingIngestor(ctx, mDumpWriter)

		return ingestor
	}

	// ingesting n entries
	nIngest := func(t *testing.T, clusterRole []*rbacv1.ClusterRoleBinding) *ClusterRoleBindingIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewClusterRoleBindingIngestor(ctx, mDumpWriter)

		buffer := &rbacv1.ClusterRoleBindingList{}

		for _, clusterRole := range clusterRole {
			buffer.Items = append(buffer.Items, *clusterRole)
		}
		mDumpWriter.EXPECT().Write(mock.Anything, buffer, collector.ClusterRolesPath).Return(nil).Once()

		return ingestor
	}

	type args struct {
		clusterRoleBinding []*rbacv1.ClusterRoleBinding
	}
	tests := []struct {
		name           string
		dumperIngestor *ClusterRoleIngestor
		testfct        func(t *testing.T, clusterRole []*rbacv1.ClusterRoleBinding) *ClusterRoleBindingIngestor
		args           args
		wantErr        bool
	}{
		{
			name:    "no entry",
			testfct: noIngest,
			args: args{
				clusterRoleBinding: []*rbacv1.ClusterRoleBinding{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:    "entries found",
			testfct: nIngest,
			args: args{
				clusterRoleBinding: []*rbacv1.ClusterRoleBinding{
					collector.FakeClusterRoleBinding("name1"),
					collector.FakeClusterRoleBinding("name2"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ingestor := tt.testfct(t, tt.args.clusterRoleBinding)
			for _, clusterRoleBinding := range tt.args.clusterRoleBinding {
				if err := ingestor.IngestClusterRoleBinding(ctx, clusterRoleBinding); (err != nil) != tt.wantErr {
					t.Errorf("Dumper.IngestNode() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err := ingestor.Complete(ctx); err != nil {
				t.Errorf("Dumper.IngestNode() error = %v", err)
			}
		})
	}
}
