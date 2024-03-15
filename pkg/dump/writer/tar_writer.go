package writer

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
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
	TarWriterExtension = ".tar.gz"
	TarWriterChmod     = 0600
	TarTypeTag         = "tar"

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
	fs         *afero.MemMapFs
	mu         sync.Mutex
}

func NewTarWriter(ctx context.Context, directoryPath string, resName string) (*TarWriter, error) {
	tarPath := path.Join(directoryPath, fmt.Sprintf("%s%s", resName, TarWriterExtension))
	tarFile, err := createTarFile(tarPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create tar file: %w", err)
	}
	gzipWriter := gzip.NewWriter(tarFile)

	return &TarWriter{
		tarPath:    tarPath,
		gzipWriter: gzipWriter,
		tarWriter:  tar.NewWriter(gzipWriter),
		tarFile:    tarFile,
		fs:         &afero.MemMapFs{},
		mu:         sync.Mutex{},
	}, nil
}

func createTarFile(tarPath string) (*os.File, error) {
	log.I.Debugf("Creating tar file %s", tarPath)
	err := os.MkdirAll(filepath.Dir(tarPath), WriterDirChmod)
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

// Write function writes the Kubernetes object to a buffer
// All buffer are stored in a map which is flushed at the end of every type processed
func (t *TarWriter) Write(ctx context.Context, k8sObj any, filePath string) error {
	log.I.Debugf("Writing to file %s", filePath)
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := marshalK8sObj(k8sObj)
	if err != nil {
		return err
	}

	// Create directories if they do not exist
	err = t.fs.MkdirAll(filepath.Dir(filePath), WriterDirChmod)
	if err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	file, err := t.fs.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, FileWriterChmod)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	// t.files[pathObj] = &file
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

// Flush function flushes all kubernetes object from the buffers to the tar file
func (t *TarWriter) Flush(ctx context.Context) error {
	log.I.Debug("Flushing writers")
	t.mu.Lock()
	defer t.mu.Unlock()

	// Walking the memfs and copying all the files to the tar
	err := afero.Walk(t.fs, "", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		file, err := t.fs.Open(path)
		if err != nil {
			return fmt.Errorf("open file %s from fs: %w", path, err)
		}
		fileInfo, err := t.fs.Stat(path)
		if err != nil {
			return fmt.Errorf("stat file %s from fs: %w", path, err)
		}
		tarHader, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return fmt.Errorf("create tar header for file %s: %w", path, err)
		}
		tarHader.Name = path
		if err := t.tarWriter.WriteHeader(tarHader); err != nil {
			return fmt.Errorf("write tar header for file %s: %w", path, err)
		}
		if _, err := io.Copy(t.tarWriter, file); err != nil {
			return fmt.Errorf("copy file %s to tar: %w", path, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk fs: %w", err)
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
