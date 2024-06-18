package blob

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/dump"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	awsv2cfg "github.com/aws/aws-sdk-go-v2/config"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/memblob"
	"gocloud.dev/blob/s3blob"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var (
	ErrInvalidBucketName = errors.New("empty bucket name")
)

type BlobStore struct {
	bucketName string
	cfg        *config.KubehoundConfig
	region     string
}

var _ puller.DataPuller = (*BlobStore)(nil)

func NewBlobStorage(cfg *config.KubehoundConfig, blobConfig *config.BlobConfig) (*BlobStore, error) {
	if blobConfig.Bucket == "" {
		return nil, ErrInvalidBucketName
	}

	return &BlobStore{
		bucketName: blobConfig.Bucket,
		cfg:        cfg,
		region:     blobConfig.Region,
	}, nil
}

func getKeyPath(clusterName, runID string) string {
	return fmt.Sprintf("%s%s", dump.DumpIngestorResultName(clusterName, runID), writer.TarWriterExtension)
}

func (bs *BlobStore) openBucket(ctx context.Context) (*blob.Bucket, error) {
	urlStruct, err := url.Parse(bs.bucketName)
	if err != nil {
		return nil, err
	}
	cloudPrefix, bucketName := urlStruct.Scheme, urlStruct.Host
	var bucket *blob.Bucket
	switch cloudPrefix {
	case "file":
		// url Parse not working for local files, using raw name file:///path/to/dir
		bucket, err = blob.OpenBucket(ctx, bs.bucketName)
	case "wasbs":
		// AZURE_STORAGE_ACCOUNT env is set in conf/k8s
		bucketName = urlStruct.User.Username()
		bucket, err = blob.OpenBucket(ctx, "azblob://"+bucketName)
	case "s3":
		// Establish a AWS V2 Config.
		// See https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/ for more info.
		cfg, err := awsv2cfg.LoadDefaultConfig(
			ctx,
			awsv2cfg.WithRegion(bs.region),
		)
		if err != nil {
			return nil, err
		}

		// Create a *blob.Bucket.
		clientV2 := s3v2.NewFromConfig(cfg)
		bucket, err = s3blob.OpenBucketV2(ctx, clientV2, bucketName, nil)
		if err != nil {
			return nil, err
		}
	default:
		bucket, err = blob.OpenBucket(ctx, cloudPrefix+"://"+bucketName)
	}

	if err != nil {
		return nil, err
	}

	return bucket, nil
}

// Pull pulls the data from the blob store (e.g: s3) and returns the path of the folder containing the archive
func (bs *BlobStore) Put(outer context.Context, archivePath string, clusterName string, runID string) error {
	log.I.Infof("Pulling data from blob store bucket %s, %s, %s", bs.bucketName, clusterName, runID)
	spanPut, ctx := span.SpanIngestRunFromContext(outer, span.IngestorBlobPull)
	var err error
	defer func() { spanPut.Finish(tracer.WithError(err)) }()

	key := getKeyPath(clusterName, runID)
	log.I.Infof("Downloading archive (%s) from blob store", key)
	b, err := bs.openBucket(ctx)
	if err != nil {
		return err
	}
	defer b.Close()
	log.I.Infof("Opening archive file %s", archivePath)
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	log.I.Infof("Downloading archive (%q) from blob store", key)
	w := bufio.NewReader(f)
	err = b.Upload(ctx, key, w, &blob.WriterOptions{
		ContentType: "application/gzip",
	})
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}

// Pull pulls the data from the blob store (e.g: s3) and returns the path of the folder containing the archive
func (bs *BlobStore) Pull(outer context.Context, clusterName string, runID string) (string, error) {
	log.I.Infof("Pulling data from blob store bucket %s, %s, %s", bs.bucketName, clusterName, runID)
	spanPull, ctx := span.SpanIngestRunFromContext(outer, span.IngestorBlobPull)
	var err error
	defer func() { spanPull.Finish(tracer.WithError(err)) }()

	key := getKeyPath(clusterName, runID)
	log.I.Infof("Downloading archive (%s) from blob store", key)
	b, err := bs.openBucket(ctx)
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
	archivePath := filepath.Join(dirname, config.DefaultArchiveName)
	f, err := os.Create(archivePath)
	if err != nil {
		return archivePath, err
	}
	defer f.Close()

	log.I.Infof("Downloading archive (%q) from blob store", key)
	w := bufio.NewWriter(f)
	err = b.Download(ctx, key, w, nil)
	if err != nil {
		return archivePath, err
	}

	err = f.Sync()
	if err != nil {
		return archivePath, err
	}

	return archivePath, nil
}

func (bs *BlobStore) Extract(ctx context.Context, archivePath string) error {
	spanExtract, _ := span.SpanIngestRunFromContext(ctx, span.IngestorBlobExtract)
	var err error
	defer func() { spanExtract.Finish(tracer.WithError(err)) }()

	basePath := filepath.Dir(archivePath)
	err = puller.CheckSanePath(archivePath, bs.cfg.Ingestor.TempDir)
	if err != nil {
		return fmt.Errorf("Dangerous file path used during extraction, aborting: %w", err)
	}

	dryRun := false
	err = puller.ExtractTarGz(dryRun, archivePath, basePath, bs.cfg.Ingestor.MaxArchiveSize)
	if err != nil {
		return err
	}

	return nil
}

// Once downloaded and processed, we should cleanup the disk so we can reduce the disk usage
// required for large infrastructure
func (bs *BlobStore) Close(ctx context.Context, archivePath string) error {
	spanClose, _ := span.SpanIngestRunFromContext(ctx, span.IngestorBlobClose)
	var err error
	defer func() { spanClose.Finish(tracer.WithError(err)) }()

	path := filepath.Base(archivePath)
	err = puller.CheckSanePath(archivePath, bs.cfg.Ingestor.TempDir)
	if err != nil {
		return fmt.Errorf("Dangerous file path used while closing, aborting: %w", err)
	}

	err = os.RemoveAll(path)
	if err != nil {
		return err
	}

	return nil
}
