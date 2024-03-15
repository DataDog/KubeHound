package writer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

const (
	WriterDirChmod = 0700
)

// The DumperWriter handle multiple types of writer (file, tar, ...)
// It is used to centralized all writes and therefore handle all the files at once
// Some of the writers support multi-threading (WorkerNumber to retrieve the info)
//
//go:generate mockery --name DumperWriter --output mockwriter --case underscore --filename writer.go --with-expecter
type DumperWriter interface {
	Write(context.Context, any, string) error
	Flush(context.Context) error
	Close(context.Context) error

	// Multi-threading the dump with one worker for each types
	// The number of workers is set to 7 to have one thread per k8s object type to pull  fronm the Kubernetes API
	// Using single thread when zipping to avoid concurency issues
	WorkerNumber() int
	OutputPath() string
}

func DumperWriterFactory(ctx context.Context, compression bool, directoryPath string, resultName string) (DumperWriter, error) {
	// if compression is enabled, create the tar.gz file
	if compression {
		log.I.Infof("Compression enabled")

		return NewTarWriter(ctx, directoryPath, resultName)
	}

	return NewFileWriter(ctx, directoryPath, resultName)
}

func marshalK8sObj(obj interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Kubernetes object: %w", err)
	}

	return jsonData, nil
}
