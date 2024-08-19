package writer

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/spf13/afero"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	TarWriterChmod = 0600
	TarTypeTag     = "tar"

	// Multi-threading the dump with one worker for each types
	// The number of workers is set to the number of differents entities (roles, pods, ...)
	// 1 thread per k8s object type to pull from the Kubernetes API
	// 0 means as many thread as k8s entity types (calculated by the dumper_pipeline)
	TarWorkerNumber = 0
)

// TarWriter keeps track of all handlers used to create the tar file
// The write occurs in memory and is flushed to the file at the end of the process
type TarWriter struct {
	tarFile    *os.File
	gzipWriter *gzip.Writer
	tarWriter  *tar.Writer
	tarPath    string
	mu         sync.Mutex
	fsWriter   *FSWriter
}

func NewTarWriter(ctx context.Context, tarPath string) (*TarWriter, error) {
	tarFile, err := createTarFile(tarPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create tar file: %w", err)
	}
	gzipWriter := gzip.NewWriter(tarFile)

	fsWriter, err := NewFSWriter(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating fs writer: %w", err)
	}

	return &TarWriter{
		tarPath:    tarPath,
		gzipWriter: gzipWriter,
		tarWriter:  tar.NewWriter(gzipWriter),
		tarFile:    tarFile,
		fsWriter:   fsWriter,
		mu:         sync.Mutex{},
	}, nil
}

func createTarFile(tarPath string) (*os.File, error) {
	log.I.Debugf("Creating tar file %s", tarPath)
	err := os.MkdirAll(filepath.Dir(tarPath), WriterDirMod)
	if err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	return os.Create(tarPath)
}

func (f *TarWriter) OutputPath() string {
	return f.tarPath
}

func (f *TarWriter) WorkerNumber() int {
	return TarWorkerNumber
}

func (f *TarWriter) WriteMetadata(ctx context.Context) error {
	return nil
}

// Write function writes the Kubernetes object to a buffer
// All buffer are stored in a map which is flushed at the end of every type processed
func (t *TarWriter) Write(ctx context.Context, k8sObj []byte, filePath string) error {
	log.I.Debugf("Writing to file %s", filePath)
	t.mu.Lock()
	defer t.mu.Unlock()

	// Writing file to Afero memfs
	err := t.fsWriter.WriteFile(ctx, filePath, k8sObj)
	if err != nil {
		return fmt.Errorf("write file %s: %w", filePath, err)
	}

	return nil
}

// Flush function flushes all kubernetes object from the buffers to the tar file
func (t *TarWriter) Flush(ctx context.Context) error {
	log.I.Debug("Flushing writers")
	span, _ := tracer.StartSpanFromContext(ctx, span.DumperWriterFlush, tracer.Measured())
	span.SetTag(tag.DumperWriterTypeTag, TarTypeTag)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	t.mu.Lock()
	defer t.mu.Unlock()

	fs := afero.NewIOFS(t.fsWriter.vfs)

	err = t.tarWriter.AddFS(fs)
	if err != nil {
		return fmt.Errorf("add fs to tar: %w", err)
	}

	err = t.tarWriter.Flush()
	if err != nil {
		return fmt.Errorf("flush tar: %w", err)
	}

	return nil
}

// Close all the handler used to write the tar file
// Need to be closed only when all assets are dumped
func (t *TarWriter) Close(ctx context.Context) error {
	log.I.Debug("Closing handlers for tar")
	span, _ := tracer.StartSpanFromContext(ctx, span.DumperWriterClose, tracer.Measured())
	span.SetTag(tag.DumperWriterTypeTag, TarTypeTag)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()
	err = t.tarWriter.Close()
	if err != nil {
		return err
	}

	err = t.gzipWriter.Close()
	if err != nil {
		return err
	}

	err = t.tarFile.Close()
	if err != nil {
		return err
	}

	return nil
}
