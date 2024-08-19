package dump

import (
	"fmt"
	"path"
	"reflect"
	"testing"
)

const (
	validClusterName = "cluster1.k8s.local"
	validRunID       = "01j2qs8th6yarr5hkafysekn0j"
	// cluster name with invalid characters (for instance /)
	nonValidClusterName = "cluster1.k8s.local/"
	// RunID with capital letters
	nonValidRunID = "01j2qs8TH6yarr5hkafysekn0j"
)

func TestDumpResult_GetFilename(t *testing.T) {
	t.Parallel()

	type fields struct {
		ClusterName string
		RunID       string
		IsDir       bool
		Extension   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "valide dump result object no compression",
			fields: fields{
				ClusterName: validClusterName,
				RunID:       validRunID,
				IsDir:       true,
				Extension:   "",
			},
			want: fmt.Sprintf("%s%s", "kubehound_", "cluster1.k8s.local_01j2qs8th6yarr5hkafysekn0j"),
		},
		{
			name: "valide dump result object compressed",
			fields: fields{
				ClusterName: validClusterName,
				RunID:       validRunID,
				IsDir:       false,
				Extension:   "tar.gz",
			},
			want: fmt.Sprintf("%s%s", "kubehound_", "cluster1.k8s.local_01j2qs8th6yarr5hkafysekn0j.tar.gz"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			i := &DumpResult{
				Metadata: Metadata{
					ClusterName: tt.fields.ClusterName,
					RunID:       tt.fields.RunID,
				},
				isDir:     tt.fields.IsDir,
				extension: tt.fields.Extension,
			}
			if got := i.GetFilename(); got != tt.want {
				t.Errorf("DumpResult.GetFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDumpResult_GetFullPath(t *testing.T) {
	t.Parallel()

	type fields struct {
		ClusterName string
		RunID       string
		IsDir       bool
		Extension   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "valide dump result object no compression",
			fields: fields{
				ClusterName: validClusterName,
				RunID:       validRunID,
				IsDir:       true,
				Extension:   "",
			},
			want: path.Join(validClusterName, fmt.Sprintf("%s%s", "kubehound_", "cluster1.k8s.local_01j2qs8th6yarr5hkafysekn0j")),
		},
		{
			name: "valide dump result object compressed",
			fields: fields{
				ClusterName: validClusterName,
				RunID:       validRunID,
				IsDir:       false,
				Extension:   "tar.gz",
			},
			want: path.Join(validClusterName, fmt.Sprintf("%s%s", "kubehound_", "cluster1.k8s.local_01j2qs8th6yarr5hkafysekn0j.tar.gz")),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			i := &DumpResult{
				Metadata: Metadata{
					ClusterName: tt.fields.ClusterName,
					RunID:       tt.fields.RunID,
				},
				isDir:     tt.fields.IsDir,
				extension: tt.fields.Extension,
			}
			if got := i.GetFullPath(); got != tt.want {
				t.Errorf("DumpResult.GetFullPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDumpResult(t *testing.T) {
	t.Parallel()

	type args struct {
		clusterName  string
		runID        string
		isCompressed bool
	}
	tests := []struct {
		name    string
		args    args
		want    *DumpResult
		wantErr bool
	}{
		{
			name: "valid entry",
			args: args{
				clusterName:  validClusterName,
				runID:        validRunID,
				isCompressed: false,
			},
			want: &DumpResult{
				Metadata: Metadata{
					ClusterName: validClusterName,
					RunID:       validRunID,
				},
				isDir: true,
			},
			wantErr: false,
		},
		{
			name: "invalid clustername",
			args: args{
				clusterName:  nonValidClusterName,
				runID:        validRunID,
				isCompressed: false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty clustername",
			args: args{
				clusterName:  "",
				runID:        validRunID,
				isCompressed: false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid runID",
			args: args{
				clusterName:  validClusterName,
				runID:        nonValidRunID,
				isCompressed: false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid runID",
			args: args{
				clusterName:  validClusterName,
				runID:        "",
				isCompressed: false,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewDumpResult(tt.args.clusterName, tt.args.runID, tt.args.isCompressed)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDumpResult() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDumpResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
