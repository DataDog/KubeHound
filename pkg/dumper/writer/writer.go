package writer

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

const (
	WriterDirChmod = 0700
)

// The DumperWriter handle multiple types of writer (file, tar, ...)
// It is used to centralized all writes and therefore handle all the files at once
// Some of the writers support multi-threading (WorkerNumber to retrieve the info)
type DumperWriter interface {
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

func DumperWriterFactory(ctx context.Context, compression bool) (DumperWriter, error) {
	// if compression is enabled, create the tar.gz file
	if compression {
		log.I.Infof("Compression enabled")

		return &TarWriter{}, nil
	}

	return &FileWriter{}, nil
}
