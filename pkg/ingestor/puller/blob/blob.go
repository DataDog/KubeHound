package blob

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/dump"
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
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"

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
	if blobConfig.BucketUrl == "" {
		return nil, ErrInvalidBucketName
	}

	return &BlobStore{
		bucketName: blobConfig.BucketUrl,
		cfg:        cfg,
		region:     blobConfig.Region,
	}, nil
}

func (bs *BlobStore) openBucket(ctx context.Context) (*blob.Bucket, error) {
	l := log.Logger(ctx)
	l.Info("Opening bucket", log.String("bucket_name", bs.bucketName))

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

func (bs *BlobStore) listFiles(ctx context.Context, b *blob.Bucket, prefix string, recursive bool, listObjects []*puller.ListObject) ([]*puller.ListObject, error) {
	iter := b.List(&blob.ListOptions{
		Delimiter: "/",
		Prefix:    prefix,
	})
	for {
		obj, err := iter.Next(ctx)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("listing objects: %w", err)
		}

		if obj.IsDir && recursive {
			listObjects, _ = bs.listFiles(ctx, b, obj.Key, true, listObjects)
		}
		listObjects = append(listObjects, &puller.ListObject{
			Key:     obj.Key,
			ModTime: obj.ModTime,
		})
	}

	return listObjects, nil
}

func (bs *BlobStore) ListFiles(ctx context.Context, prefix string, recursive bool) ([]*puller.ListObject, error) {
	b, err := bs.openBucket(ctx)
	if err != nil {
		return nil, err
	}
	listObjects := []*puller.ListObject{}

	return bs.listFiles(ctx, b, prefix, recursive, listObjects)
}

// Pull pulls the data from the blob store (e.g: s3) and returns the path of the folder containing the archive
func (bs *BlobStore) Put(outer context.Context, archivePath string, clusterName string, runID string) error {
	l := log.Logger(outer)
	var err error

	// Triggering a span only when it is an actual run and not the rehydration process (download the kubehound dump to get the metadata)
	if log.GetRunIDFromContext(outer) != "" {
		var spanPut ddtrace.Span
		spanPut, outer = span.SpanRunFromContext(outer, span.IngestorBlobPull)
		defer func() { spanPut.Finish(tracer.WithError(err)) }()
	}
	l.Info("Putting data on blob store bucket", log.String("bucket_name", bs.bucketName), log.String(log.FieldClusterKey, clusterName), log.String(log.FieldRunIDKey, runID))

	dumpResult, err := dump.NewDumpResult(clusterName, runID, true)
	if err != nil {
		return err
	}
	key := dumpResult.GetFullPath()
	l.Info("Opening bucket", log.String("bucket_name", bs.bucketName))
	b, err := bs.openBucket(outer)
	if err != nil {
		return err
	}
	defer b.Close()
	l.Info("Opening archive file", log.String(log.FieldPathKey, archivePath))
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	l.Info("Uploading archive from blob store", log.String("key", key))
	w := bufio.NewReader(f)
	err = b.Upload(outer, key, w, &blob.WriterOptions{
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
func (bs *BlobStore) Pull(outer context.Context, key string) (string, error) {
	l := log.Logger(outer)
	var err error
	if log.GetRunIDFromContext(outer) != "" {
		var spanPull ddtrace.Span
		spanPull, outer = span.SpanRunFromContext(outer, span.IngestorBlobPull)
		defer func() { spanPull.Finish(tracer.WithError(err)) }()
	}
	l.Info("Pulling data from blob store bucket", log.String("bucket_name", bs.bucketName), log.String("key", key))

	b, err := bs.openBucket(outer)
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

	l.Info("Created temporary directory", log.String(log.FieldPathKey, dirname))
	archivePath := filepath.Join(dirname, config.DefaultArchiveName)
	f, err := os.Create(archivePath)
	if err != nil {
		return archivePath, err
	}
	defer f.Close()

	l.Info("Downloading archive from blob store", log.String("key", key))
	w := bufio.NewWriter(f)
	err = b.Download(outer, key, w, nil)
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
	var err error
	if log.GetRunIDFromContext(ctx) != "" {
		var spanPull ddtrace.Span
		spanPull, ctx = span.SpanRunFromContext(ctx, span.IngestorBlobExtract)
		defer func() { spanPull.Finish(tracer.WithError(err)) }()
	}

	basePath := filepath.Dir(archivePath)
	err = puller.CheckSanePath(archivePath, basePath)
	if err != nil {
		return fmt.Errorf("dangerous file path used during extraction, aborting: %w", err)
	}

	dryRun := false
	err = puller.ExtractTarGz(ctx, dryRun, archivePath, basePath, bs.cfg.Ingestor.MaxArchiveSize)
	if err != nil {
		return err
	}

	return nil
}

// Once downloaded and processed, we should cleanup the disk so we can reduce the disk usage
// required for large infrastructure
func (bs *BlobStore) Close(ctx context.Context, archivePath string) error {
	var err error
	if log.GetRunIDFromContext(ctx) != "" {
		var spanClose ddtrace.Span
		spanClose, _ = span.SpanRunFromContext(ctx, span.IngestorBlobClose)
		defer func() { spanClose.Finish(tracer.WithError(err)) }()
	}

	path := filepath.Dir(archivePath)
	err = puller.CheckSanePath(archivePath, bs.cfg.Ingestor.TempDir)
	if err != nil {
		return fmt.Errorf("dangerous file path used while closing, aborting: %w", err)
	}

	err = os.RemoveAll(path)
	if err != nil {
		return err
	}

	return nil
}
