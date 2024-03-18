package pipeline

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockwriter "github.com/DataDog/KubeHound/pkg/dump/writer/mockwriter"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestDumpIngestor_IngestRoleBinding(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// no ingestion
	noIngest := func(t *testing.T, _ []types.RoleBindingType) *RoleBindingIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewRoleBindingIngestor(ctx, mDumpWriter)

		return ingestor
	}

	// ingesting n entries
	nIngest := func(t *testing.T, roleBindings []types.RoleBindingType) *RoleBindingIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewRoleBindingIngestor(ctx, mDumpWriter)

		buffer := make(map[string]*rbacv1.RoleBindingList)
		for _, roleBinding := range roleBindings {
			err := bufferObject[rbacv1.RoleBindingList, types.RoleBindingType](ctx, ingestRoleBindingPath(roleBinding), buffer, roleBinding)
			if err != nil {
				t.Fatal(err)
			}
		}

		for path, endpoint := range buffer {
			mDumpWriter.EXPECT().Write(ctx, endpoint, path).Return(nil).Once()
		}

		return ingestor
	}

	type args struct {
		roleBindings []types.RoleBindingType
	}
	tests := []struct {
		name     string
		ingestor *RoleBindingIngestor
		testfct  func(t *testing.T, clusterRole []types.RoleBindingType) *RoleBindingIngestor
		args     args
		wantErr  bool
	}{
		{
			name:    "no entry",
			testfct: noIngest,
			args: args{
				roleBindings: []types.RoleBindingType{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:    "entries found",
			testfct: nIngest,
			args: args{
				roleBindings: []types.RoleBindingType{
					collector.FakeRoleBinding("namespace1", "name1"),
					collector.FakeRoleBinding("namespace2", "name2"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ingestor := tt.testfct(t, tt.args.roleBindings)
			for _, roleBinding := range tt.args.roleBindings {
				if err := ingestor.IngestRoleBinding(ctx, roleBinding); (err != nil) != tt.wantErr {
					t.Errorf("ingestor.IngestRoleBinding() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err := ingestor.Complete(ctx); err != nil {
				t.Errorf("ingestor.Complete() error = %v", err)
			}
		})
	}
}
