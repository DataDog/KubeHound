package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type S3Store struct {
	client *s3.Client
	bucket string
}

const maxRetry = 3
const maxBackoffDelay = 5

func NewS3Store(ctx context.Context, region string, bucket string) (*S3Store, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxAttempts(retry.NewStandard(), maxRetry)
		}),
		config.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxBackoffDelay(retry.NewStandard(), time.Second*maxBackoffDelay)
		}),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	return &S3Store{
		client: client,
		bucket: bucket,
	}, nil
}

func (s *S3Store) formatS3URI(objectKey string) string {
	return fmt.Sprintf("s3://%s/%s", s.bucket, objectKey)
}

func (s *S3Store) Upload(ctx context.Context, objectKey string, data []byte) error {
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorS3Push, tracer.Measured())
	span.SetTag(tag.CollectorS3Bucket, s.bucket)
	span.SetTag(tag.CollectorS3key, objectKey)
	defer span.Finish()
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(data),
	})

	if err != nil {
		return fmt.Errorf("upload file to %s: %w", s.formatS3URI(objectKey), err)
	}

	return nil
}

func (s *S3Store) Download(ctx context.Context, objectKey string, filePath string) ([]byte, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorS3Push, tracer.Measured())
	span.SetTag(tag.CollectorS3Bucket, s.bucket)
	span.SetTag(tag.CollectorS3key, objectKey)
	defer span.Finish()
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		return nil, fmt.Errorf("download file from %s: %w", s.formatS3URI(objectKey), err)
	}

	defer result.Body.Close()
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("read object body from %s: %w", s.formatS3URI(objectKey), err)
	}
	return data, nil
}

func SaveToFile(data []byte, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("write content to file %s: %w", filePath, err)
	}

	return nil
}
