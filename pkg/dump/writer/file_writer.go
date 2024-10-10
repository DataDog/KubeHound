package writer

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/spf13/afero"
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
	directoryOutput string
	mu              sync.Mutex
	fsWriter        *FSWriter
}

func NewFileWriter(ctx context.Context, directoryOutput string) (*FileWriter, error) {
	fsWriter, err := NewFSWriter(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating fs writer: %w", err)
	}

	return &FileWriter{
		directoryOutput: directoryOutput,
		mu:              sync.Mutex{},
		fsWriter:        fsWriter,
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
func (f *FileWriter) Write(ctx context.Context, k8sObj []byte, pathObj string) error {
	l := log.Logger(ctx)
	l.Debug("Writing to file", log.String("path", pathObj))
	f.mu.Lock()
	defer f.mu.Unlock()

	// Writing file to Afero memfs
	err := f.fsWriter.WriteFile(ctx, pathObj, k8sObj)
	if err != nil {
		return fmt.Errorf("write file %s: %w", pathObj, err)
	}

	// Constructing output complete path to write the file on disk
	fileDiskPath := path.Join(f.directoryOutput, pathObj)

	fs := afero.NewIOFS(f.fsWriter.vfs)
	fsAferoSource, err := f.fsWriter.vfs.OpenFile(pathObj, os.O_RDONLY, FileWriterChmod)
	if err != nil {
		return fmt.Errorf("open file from afero %s: %w", pathObj, err)
	}

	// Create directories if they do not exist
	err = os.MkdirAll(filepath.Dir(fileDiskPath), WriterDirMod)
	if err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	fdFileDisk, err := os.Create(fileDiskPath)
	if err != nil {
		return fmt.Errorf("create file %s: %w", pathObj, err)
	}

	_, err = io.Copy(fdFileDisk, fsAferoSource)
	if err != nil {
		return fmt.Errorf("copy file %s: %w", pathObj, err)
	}

	err = fdFileDisk.Close()
	if err != nil {
		return fmt.Errorf("close file %s: %w", pathObj, err)
	}

	err = fs.Remove(pathObj)
	if err != nil {
		return fmt.Errorf("remove file from afero %s: %w", pathObj, err)
	}

	return nil
}

// No flush needed for the file writer as we are flushing the buffer at every write
func (f *FileWriter) Flush(ctx context.Context) error {
	span, _ := tracer.StartSpanFromContext(ctx, span.DumperWriterFlush, tracer.Measured())
	span.SetTag(tag.DumperWriterTypeTag, FileTypeTag)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	return nil
}

func (f *FileWriter) Close(ctx context.Context) error {
	l := log.Logger(ctx)
	l.Debug("Closing writers")
	span, _ := tracer.StartSpanFromContext(ctx, span.DumperWriterClose, tracer.Measured())
	span.SetTag(tag.DumperWriterTypeTag, FileTypeTag)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()
	err = f.fsWriter.Close(ctx)
	if err != nil {
		return fmt.Errorf("close afero: %w", err)
	}

	return nil
}
