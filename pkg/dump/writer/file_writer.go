package writer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	FileWriterChmod = 0600
	FileTypeTag     = "file"

	// Multi-threading the dump with one worker for each types
	// The number of workers is set to the number of differents entities (roles, pods, ...)
	// 1 thread per k8s object type to pull from the Kubernetes API
	// 0 means as many thread as k8s entity types (calculated by the dumper_pipeline)
	FileWriterWorkerNumber = 0
)

// The FileWriter uses a map of buffers to write data to files
// Each file has its own buffer to optimize IO calls
type FileWriter struct {
	files           map[string]*os.File
	directoryOutput string
	mu              sync.Mutex
}

func NewFileWriter(ctx context.Context, directoryOutput string, resName string) (*FileWriter, error) {
	return &FileWriter{
		files:           make(map[string]*os.File),
		directoryOutput: path.Join(directoryOutput, resName),
		mu:              sync.Mutex{},
	}, nil
}

func (f *FileWriter) OutputPath() string {
	return f.directoryOutput
}

func (f *FileWriter) WorkerNumber() int {
	return FileWriterWorkerNumber
}

// Write function writes the Kubernetes object to a buffer
// All buffer are stored in a map which is flushed at the end of every type processed
func (f *FileWriter) Write(ctx context.Context, k8sObj any, pathObj string) error {
	log.I.Debugf("Writing to file %s", pathObj)
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := marshalK8sObj(k8sObj)
	if err != nil {
		return err
	}

	// Constructing output complete path to write the file
	filePath := path.Join(f.directoryOutput, pathObj)

	// Create directories if they do not exist
	err = os.MkdirAll(filepath.Dir(filePath), WriterDirChmod)
	if err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, FileWriterChmod)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	f.files[pathObj] = file
	buffer := bufio.NewWriter(file)

	_, err = buffer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write JSON data to buffer: %w", err)
	}

	err = buffer.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

// No flush needed for the file writer as we are flushing the buffer at every write
func (f *FileWriter) Flush(ctx context.Context) error {
	return nil
}

func (f *FileWriter) Close(ctx context.Context) error {
	log.I.Debug("Closing writers")
	span, _ := tracer.StartSpanFromContext(ctx, span.DumperWriterClose, tracer.Measured())
	span.SetTag(tag.DumperWriterTypeTag, FileTypeTag)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()
	for path, file := range f.files {
		err = file.Close()
		if err != nil {
			return fmt.Errorf("failed to close writer: %w", err)
		}
		delete(f.files, path)
	}

	return nil
}
