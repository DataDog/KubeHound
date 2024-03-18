package pipeline

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockwriter "github.com/DataDog/KubeHound/pkg/dump/writer/mockwriter"
	"github.com/stretchr/testify/mock"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestDumpIngestor_IngestClusterRole(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// no ingestion
	noIngest := func(t *testing.T, _ []*rbacv1.ClusterRole) *ClusterRoleIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewClusterRoleIngestor(ctx, mDumpWriter)

		return ingestor
	}

	// ingesting n entries
	nIngest := func(t *testing.T, clusterRole []*rbacv1.ClusterRole) *ClusterRoleIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewClusterRoleIngestor(ctx, mDumpWriter)

		buffer := &rbacv1.ClusterRoleList{}

		for _, clusterRole := range clusterRole {
			buffer.Items = append(buffer.Items, *clusterRole)
		}
		mDumpWriter.EXPECT().Write(mock.Anything, buffer, collector.ClusterRolesPath).Return(nil).Once()

		return ingestor
	}

	type args struct {
		clusterRoles []*rbacv1.ClusterRole
	}
	tests := []struct {
		name     string
		ingestor *ClusterRoleIngestor
		testfct  func(t *testing.T, clusterRole []*rbacv1.ClusterRole) *ClusterRoleIngestor
		args     args
		wantErr  bool
	}{
		{
			name:    "no entry",
			testfct: noIngest,
			args: args{
				clusterRoles: []*rbacv1.ClusterRole{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:    "entries found",
			testfct: nIngest,
			args: args{
				clusterRoles: []*rbacv1.ClusterRole{
					collector.FakeClusterRole("name1"),
					collector.FakeClusterRole("name2"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ingestor := tt.testfct(t, tt.args.clusterRoles)
			for _, clusterRole := range tt.args.clusterRoles {
				if err := ingestor.IngestClusterRole(ctx, clusterRole); (err != nil) != tt.wantErr {
					t.Errorf("Dumper.IngestNode() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err := ingestor.Complete(ctx); err != nil {
				t.Errorf("Dumper.IngestNode() error = %v", err)
			}
		})
	}
}
