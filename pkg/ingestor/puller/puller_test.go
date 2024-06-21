package puller

import (
	"os"
	"testing"
)

func Test_sanitizeExtractPath(t *testing.T) {
	t.Parallel()

	type args struct {
		filePath    string
		destination string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Sanitize path",
			args:    args{filePath: "test", destination: "/tmp"},
			wantErr: false,
		},
		{
			name:    "Error on illegal path",
			args:    args{filePath: "../../test", destination: "/tmp"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := sanitizeExtractPath(tt.args.filePath, tt.args.destination); (err != nil) != tt.wantErr {
				t.Errorf("sanitizeExtractPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckSanePath(t *testing.T) {
	t.Parallel()
	type args struct {
		path       string
		baseFolder string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Path is sane",
			args:    args{path: "/tmp", baseFolder: "/tmp"},
			wantErr: false,
		},
		{
			name:    "Path is sane",
			args:    args{path: "/tmp/kubehound/kh-1234/test-cluster/id/archive.tar.gz", baseFolder: "/tmp"},
			wantErr: false,
		},
		{
			name:    "Path is sane",
			args:    args{path: "/tmp/kubehound/kh-1234/test-cluster/id/archive.tar.gz", baseFolder: "/tmp/kubehound"},
			wantErr: false,
		},
		{
			name:    "Path is NOT sane, relative path",
			args:    args{path: "../../tmp", baseFolder: "/tmp"},
			wantErr: true,
		},
		{
			name:    "Path is NOT sane, root dir",
			args:    args{path: "/", baseFolder: "/tmp"},
			wantErr: true,
		},
		{
			name:    "Path is NOT sane, invalid dir",
			args:    args{path: "/etc", baseFolder: "/tmp"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := CheckSanePath(tt.args.path, tt.args.baseFolder); (err != nil) != tt.wantErr {
				t.Errorf("CheckSanePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractTarGz(t *testing.T) {
	t.Parallel()
	type args struct {
		maxArchiveSize int64
	}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		expectedFiles []string
	}{
		{
			name:    "Extract archive",
			args:    args{maxArchiveSize: 1073741824},
			wantErr: false,
			expectedFiles: []string{
				"pod.json",
				"node.json",
				"role.json",
				"rolebinding.json",
			},
		},
		{
			name:          "Extract archive too large",
			args:          args{maxArchiveSize: 100},
			wantErr:       true,
			expectedFiles: []string{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpPath, err := os.MkdirTemp("/tmp", "kubehound-test")
			defer os.RemoveAll(tmpPath)
			if err != nil {
				t.Error(err)
			}
			dryRun := false
			if err := ExtractTarGz(dryRun, "./testdata/archive.tar.gz", tmpPath, tt.args.maxArchiveSize); (err != nil) != tt.wantErr {
				t.Errorf("ExtractTarGz() error = %v, wantErr %v", err, tt.wantErr)
			}
			for _, file := range tt.expectedFiles {
				if _, err := os.Stat(tmpPath + "/test-cluster/" + file); os.IsNotExist(err) {
					t.Errorf("ExtractTarGz() file = %v, wantErr %v", file, err)
				}
			}
		})
	}
}
