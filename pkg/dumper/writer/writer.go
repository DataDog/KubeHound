package writer

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

const (
	WriterDirChmod = 0700
)

// Explain file writers
type CollectorWriter interface {
	Initialize(context.Context, string, string) error
	Write(context.Context, []byte, string) error
	Flush(context.Context) error
	Close(context.Context) error

	// Multi-threading the dump with one worker for each types
	// The number of workers is set to 7 to have one thread per k8s object type to pull  fronm the Kubernetes API
	// Using single thread when zipping to avoid concurency issues
	WorkerNumber() int
	OutputPath() string
}

func CollectorWriterFactory(ctx context.Context, compression bool) (CollectorWriter, error) {
	// if compression is enabled, create the tar.gz file
	if compression {
		log.I.Infof("Compression enabled")

		return &TarWriter{}, nil
	}

	return &FileWriter{}, nil
}
