package blob

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/dump"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/memblob"
)

const (
	bucketName       = "file://./testdata/fakeBlobStorage"
	tempFileRegexp   = `.*/testdata/tmpdir/kh-[0-9]+/archive.tar.gz`
	validRunID       = "01j2qs8th6yarr5hkafysekn0j"
	validClusterName = "cluster1.k8s.local"
)

func getTempDir(t *testing.T) string {
	t.Helper()

	return getAbsPath(t, "testdata/tmpdir")
}

func getAbsPath(t *testing.T, filepath string) string {
	t.Helper()
	// Get current working directory to pass SaneCheckPath()
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Error getting current working directory: %v", err)

		return ""
	}

	return path.Join(pwd, filepath)
}

func dummyKubehoundConfig(t *testing.T) *config.KubehoundConfig {
	t.Helper()

	return &config.KubehoundConfig{
		Ingestor: config.IngestorConfig{
			TempDir: getTempDir(t),
		},
	}
}

func TestBlobStore_ListFiles(t *testing.T) {
	t.Parallel()
	type fields struct {
		bucketName string
		cfg        *config.KubehoundConfig
		region     string
	}
	type args struct {
		prefix    string
		recursive bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*puller.ListObject
		wantErr bool
	}{
		{
			name: "Sanitize path",
			fields: fields{
				bucketName: bucketName,
			},
			args: args{
				recursive: true,
				prefix:    "",
			},
			want: []*puller.ListObject{
				{
					Key: "cluster1.k8s.local/kubehound_cluster1.k8s.local_01j2qs8th6yarr5hkafysekn0j.tar.gz",
				},
				{
					Key: "cluster1.k8s.local/kubehound_cluster1.k8s.local_11j2qs8th6yarr5hkafysekn0j.tar.gz",
				},
				{
					Key: "cluster1.k8s.local/kubehound_cluster1.k8s.local_21j2qs8th6yarr5hkafysekn0j.tar.gz",
				},
				{
					Key: "cluster1.k8s.local/",
				},
				{
					Key: "cluster2.k8s.local/kubehound_cluster2.k8s.local_01j2qs8th6yarr5hkafysekn0j.tar.gz",
				},
				{
					Key: "cluster2.k8s.local/kubehound_cluster2.k8s.local_11j2qs8th6yarr5hkafysekn0j.tar.gz",
				},
				{
					Key: "cluster2.k8s.local/",
				},
			},
			wantErr: false,
		},
		{
			name: "Sanitize path",
			fields: fields{
				bucketName: bucketName,
			},
			args: args{
				recursive: true,
				prefix:    "cluster1.k8s.local",
			},
			want: []*puller.ListObject{
				{
					Key: "cluster1.k8s.local/kubehound_cluster1.k8s.local_01j2qs8th6yarr5hkafysekn0j.tar.gz",
				},
				{
					Key: "cluster1.k8s.local/kubehound_cluster1.k8s.local_11j2qs8th6yarr5hkafysekn0j.tar.gz",
				},
				{
					Key: "cluster1.k8s.local/kubehound_cluster1.k8s.local_21j2qs8th6yarr5hkafysekn0j.tar.gz",
				},
				{
					Key: "cluster1.k8s.local/",
				},
			},
			wantErr: false,
		},
		{
			name: "Sanitize path",
			fields: fields{
				bucketName: bucketName,
			},
			args: args{
				recursive: true,
				prefix:    "cluster1.k8s.local/",
			},
			want: []*puller.ListObject{
				{
					Key: "cluster1.k8s.local/kubehound_cluster1.k8s.local_01j2qs8th6yarr5hkafysekn0j.tar.gz",
				},
				{
					Key: "cluster1.k8s.local/kubehound_cluster1.k8s.local_11j2qs8th6yarr5hkafysekn0j.tar.gz",
				},
				{
					Key: "cluster1.k8s.local/kubehound_cluster1.k8s.local_21j2qs8th6yarr5hkafysekn0j.tar.gz",
				},
			},
			wantErr: false,
		},
		{
			name: "Sanitize path",
			fields: fields{
				bucketName: bucketName,
			},
			args: args{
				recursive: false,
				prefix:    "",
			},
			want: []*puller.ListObject{
				{
					Key: "cluster1.k8s.local/",
				},
				{
					Key: "cluster2.k8s.local/",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			bs := &BlobStore{
				bucketName: tt.fields.bucketName,
				cfg:        tt.fields.cfg,
				region:     tt.fields.region,
			}
			got, err := bs.ListFiles(ctx, tt.args.prefix, tt.args.recursive)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlobStore.ListFiles() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			// Reset modtime to avoid comparison issues
			for _, v := range got {
				v.ModTime = time.Time{}
			}

			if !reflect.DeepEqual(got, tt.want) {
				for i, v := range got {
					t.Logf("Got: %d: %s", i, v.Key)
				}
				t.Errorf("BlobStore.ListFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlobStore_Pull(t *testing.T) {
	t.Parallel()

	// Get current working directory to pass SaneCheckPath()
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Error getting current working directory: %v", err)

		return
	}
	tmpDir := path.Join(pwd, "testdata/tmpdir")
	type fields struct {
		bucketName string
		cfg        *config.KubehoundConfig
		region     string
	}
	type args struct {
		clusterName string
		runID       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Pulling file successfully",
			fields: fields{
				bucketName: bucketName,
				cfg: &config.KubehoundConfig{
					Ingestor: config.IngestorConfig{
						TempDir: tmpDir,
					},
				},
			},
			args: args{
				clusterName: validClusterName,
				runID:       validRunID,
			},
			wantErr: false,
		},
		{
			name: "Empty tmp dir",
			fields: fields{
				bucketName: bucketName,
				cfg: &config.KubehoundConfig{
					Ingestor: config.IngestorConfig{},
				},
			},
			args: args{
				clusterName: validClusterName,
				runID:       validRunID,
			},
			wantErr: true,
		},
		{
			name: "Wrong cluster name",
			fields: fields{
				bucketName: bucketName,
				cfg: &config.KubehoundConfig{
					Ingestor: config.IngestorConfig{
						TempDir: tmpDir,
					},
				},
			},
			args: args{
				clusterName: "cluster4.k8s.local",
				runID:       validRunID,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			bs := &BlobStore{
				bucketName: tt.fields.bucketName,
				cfg:        tt.fields.cfg,
				region:     tt.fields.region,
			}
			dumpResult, err := dump.NewDumpResult(tt.args.clusterName, tt.args.runID, true)
			if err != nil {
				t.Errorf("dump.NewDumpResult() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			key := dumpResult.GetFullPath()
			got, err := bs.Pull(ctx, key)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlobStore.Pull() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			// No path was returned so no need to go further
			if got == "" {
				return
			}

			re := regexp.MustCompile(tempFileRegexp)
			if !re.MatchString(got) {
				t.Errorf("Path is not valid() = %q, should respect %v", got, tempFileRegexp)

				return
			}

			err = bs.Close(ctx, got)
			if err != nil {
				t.Errorf("bs.Close() error = %v", err)
			}
		})
	}
}

func TestNewBlobStorage(t *testing.T) {
	t.Parallel()
	type args struct {
		cfg        *config.KubehoundConfig
		blobConfig *config.BlobConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *BlobStore
		wantErr bool
	}{
		{
			name: "empty bucket name",
			args: args{
				blobConfig: &config.BlobConfig{
					BucketUrl: "",
				},
				cfg: &config.KubehoundConfig{
					Ingestor: config.IngestorConfig{
						TempDir: getTempDir(t),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid blob storage",
			args: args{
				blobConfig: &config.BlobConfig{
					BucketUrl: "fakeBlobStorage",
				},
				cfg: &config.KubehoundConfig{
					Ingestor: config.IngestorConfig{
						TempDir: getTempDir(t),
					},
				},
			},
			want: &BlobStore{
				bucketName: "fakeBlobStorage",
				cfg: &config.KubehoundConfig{
					Ingestor: config.IngestorConfig{
						TempDir: getTempDir(t),
					},
				},
				region: "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewBlobStorage(tt.args.cfg, tt.args.blobConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBlobStorage() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlobStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlobStore_Put(t *testing.T) {
	t.Parallel()
	fakeBlobStoragePut := "./testdata/fakeBlobStoragePut"
	bucketNamePut := fmt.Sprintf("file://%s", fakeBlobStoragePut)
	type fields struct {
		bucketName string
		cfg        *config.KubehoundConfig
		region     string
	}
	type args struct {
		archivePath string
		clusterName string
		runID       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Push new dump",
			fields: fields{
				bucketName: bucketNamePut,
				cfg:        dummyKubehoundConfig(t),
			},
			args: args{
				archivePath: "./testdata/archive.tar.gz",
				clusterName: validClusterName,
				runID:       "91j2qs8th6yarr5hkafysekn0j",
			},
			wantErr: false,
		},
		{
			name: "non existing filepath",
			fields: fields{
				bucketName: bucketNamePut,
				cfg:        dummyKubehoundConfig(t),
			},
			args: args{
				archivePath: "./testdata/archive2.tar.gz",
				clusterName: validClusterName,
				runID:       "91j2qs8th6yarr5hkafysekn0j",
			},
			wantErr: true,
		},
		{
			name: "invalid runID",
			fields: fields{
				bucketName: bucketNamePut,
				cfg:        dummyKubehoundConfig(t),
			},
			args: args{
				archivePath: "./testdata/archive2.tar.gz",
				clusterName: validClusterName,
				runID:       "91j2qs8th6yarr5hkafysekn0T",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			bs := &BlobStore{
				bucketName: tt.fields.bucketName,
				cfg:        tt.fields.cfg,
				region:     tt.fields.region,
			}
			var err error
			if err = bs.Put(ctx, tt.args.archivePath, tt.args.clusterName, tt.args.runID); (err != nil) != tt.wantErr {
				t.Errorf("BlobStore.Put() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			// Building result path to clean up
			dumpResult, err := dump.NewDumpResult(tt.args.clusterName, tt.args.runID, true)
			if err != nil {
				t.Errorf("NewDumpResult cluster:%s runID:%s", tt.args.clusterName, tt.args.runID)
			}
			key := path.Join(bs.cfg.Ingestor.TempDir, dumpResult.GetFullPath())

			err = os.RemoveAll(path.Join(fakeBlobStoragePut, tt.args.clusterName))
			if err != nil {
				t.Errorf("Error removing file %s.attrs: %v", key, err)
			}
		})
	}
}

func TestBlobStore_Extract(t *testing.T) {
	t.Parallel()
	fakeBlobStoragePut := "./testdata/fakeBlobStorageExtract"
	bucketNameExtract := fmt.Sprintf("file://%s", fakeBlobStoragePut)
	type fields struct {
		bucketName string
		cfg        *config.KubehoundConfig
		region     string
	}
	type args struct {
		archivePath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "invalid runID",
			fields: fields{
				bucketName: bucketNameExtract,
				cfg:        dummyKubehoundConfig(t),
			},
			args: args{
				archivePath: getAbsPath(t, "testdata/archive.tar.gz"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			bs := &BlobStore{
				bucketName: tt.fields.bucketName,
				cfg:        tt.fields.cfg,
				region:     tt.fields.region,
			}
			if err := bs.Extract(ctx, tt.args.archivePath); (err != nil) != tt.wantErr {
				t.Errorf("BlobStore.Extract() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
