package pipeline

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockwriter "github.com/DataDog/KubeHound/pkg/dump/writer/mockwriter"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestDumpIngestor_IngestRole(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// no ingestion
	noIngest := func(t *testing.T, _ []types.RoleType) *RoleIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewRoleIngestor(ctx, mDumpWriter)

		return ingestor
	}

	// ingesting n entries
	nIngest := func(t *testing.T, roles []types.RoleType) *RoleIngestor {
		t.Helper()
		mDumpWriter := mockwriter.NewDumperWriter(t)
		ingestor := NewRoleIngestor(ctx, mDumpWriter)

		buffer := make(map[string]*rbacv1.RoleList)
		for _, role := range roles {
			err := bufferObject[rbacv1.RoleList, types.RoleType](ctx, ingestRolePath(role), buffer, role)
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
		roles []types.RoleType
	}
	tests := []struct {
		name     string
		ingestor *RoleIngestor
		testfct  func(t *testing.T, clusterRole []types.RoleType) *RoleIngestor
		args     args
		wantErr  bool
	}{
		{
			name:    "no entry",
			testfct: noIngest,
			args: args{
				roles: []types.RoleType{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:    "entries found",
			testfct: nIngest,
			args: args{
				roles: []types.RoleType{
					collector.FakeRole("namespace1", "name1"),
					collector.FakeRole("namespace2", "name2"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ingestor := tt.testfct(t, tt.args.roles)
			for _, role := range tt.args.roles {
				if err := ingestor.IngestRole(ctx, role); (err != nil) != tt.wantErr {
					t.Errorf("ingestor.IngestRole() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err := ingestor.Complete(ctx); err != nil {
				t.Errorf("ingestor.Complete() error = %v", err)
			}
		})
	}
}
