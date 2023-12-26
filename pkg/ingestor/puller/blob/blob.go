package blob

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/s3blob"
)

var (
	DefaultBucketName = "s3://test"
)

var (
	ErrInvalidBucketName = errors.New("Invalid bucket name")
)

type BlobStore struct {
	bucketName string
}

var _ puller.DataPuller = (*BlobStore)(nil)

func NewBlobStoragePuller() (*BlobStore, error) {
	// TODO: change me with config
	bucketName := DefaultBucketName
	if bucketName == "" {
		return nil, ErrInvalidBucketName
	}

	return &BlobStore{
		bucketName: bucketName,
	}, nil
}

// Pull pulls the data from the blob store (e.g: s3) and returns the path of the folder containing the archive
func (bs *BlobStore) Pull(ctx context.Context, clusterName string, runID string) (string, error) {
	key := puller.FormatArchiveKey(clusterName, runID)
	b, err := blob.OpenBucket(ctx, bs.bucketName)
	if err != nil {
		return "", err
	}
	defer b.Close()

	dirname, err := os.MkdirTemp(puller.BasePath, "kh-*")
	if err != nil {
		return dirname, err
	}
	archivePath := filepath.Join(dirname, "archive.tar.gz")
	f, err := os.Create(archivePath)
	if err != nil {
		return dirname, err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	err = b.Download(ctx, key, w, nil)
	if err != nil {
		return dirname, err
	}

	err = f.Sync()
	if err != nil {
		return dirname, err
	}

	return dirname, nil
}

func (bs *BlobStore) Extract(ctx context.Context, archivePath string) error {
	basePath := filepath.Base(archivePath)
	err := puller.CheckSanePath(archivePath)
	if err != nil {
		return fmt.Errorf("Dangerous file path used during extraction, aborting: %w", err)
	}
	r, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	err = puller.ExtractTarGz(r, basePath)
	if err != nil {
		return err
	}
	return nil
}

// Once downloaded and processed, we should cleanup the disk so we can reduce the disk usage
// required for large infrastructure
func (bs *BlobStore) Close(ctx context.Context, archivePath string) error {
	path := filepath.Base(archivePath)
	err := puller.CheckSanePath(archivePath)
	if err != nil {
		return fmt.Errorf("Dangerous file path used while closing, aborting: %w", err)
	}

	err = os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}
