package blob

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/s3blob"
)

var (
	ErrInvalidBucketName = errors.New("Invalid bucket name")
)

type BlobStore struct {
	bucketName string
	cfg        *config.KubehoundConfig
}

var _ puller.DataPuller = (*BlobStore)(nil)

func NewBlobStoragePuller(cfg *config.KubehoundConfig) (*BlobStore, error) {
	if cfg.Ingestor.BucketName == "" {
		return nil, ErrInvalidBucketName
	}

	return &BlobStore{
		bucketName: cfg.Ingestor.BucketName,
	}, nil
}

// Pull pulls the data from the blob store (e.g: s3) and returns the path of the folder containing the archive
func (bs *BlobStore) Pull(ctx context.Context, clusterName string, runID string) (string, error) {
	log.I.Infof("Pulling data from blob store bucket %s, %s, %s", bs.bucketName, clusterName, runID)
	key := puller.FormatArchiveKey(clusterName, runID, bs.cfg.Ingestor.ArchiveName)
	b, err := blob.OpenBucket(ctx, bs.bucketName)
	if err != nil {
		return "", err
	}
	defer b.Close()
	// MkdirTemp needs the base path to exists.
	// We thus enforce its creation here.
	err = os.MkdirAll(bs.cfg.Ingestor.TempDir, os.ModePerm)
	if err != nil {
		return "", err
	}
	dirname, err := os.MkdirTemp(bs.cfg.Ingestor.TempDir, "kh-*")
	if err != nil {
		return dirname, err
	}

	log.I.Infof("Created temporary directory %s", dirname)
	archivePath := filepath.Join(dirname, "archive.tar.gz")
	f, err := os.Create(archivePath)
	if err != nil {
		return dirname, err
	}
	defer f.Close()

	log.I.Infof("Downloading archive (%q) from blob store", key)
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
	err := puller.CheckSanePath(archivePath, bs.cfg.Ingestor.TempDir)
	if err != nil {
		return fmt.Errorf("Dangerous file path used during extraction, aborting: %w", err)
	}
	r, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	err = puller.ExtractTarGz(r, basePath, bs.cfg.Ingestor.MaxArchiveSize)
	if err != nil {
		return err
	}

	return nil
}

// Once downloaded and processed, we should cleanup the disk so we can reduce the disk usage
// required for large infrastructure
func (bs *BlobStore) Close(ctx context.Context, archivePath string) error {
	path := filepath.Base(archivePath)
	err := puller.CheckSanePath(archivePath, bs.cfg.Ingestor.TempDir)
	if err != nil {
		return fmt.Errorf("Dangerous file path used while closing, aborting: %w", err)
	}

	err = os.RemoveAll(path)
	if err != nil {
		return err
	}

	return nil
}
