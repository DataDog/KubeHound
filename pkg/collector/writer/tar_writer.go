package writer

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	TarWriterExtension = ".tar.gz"
	TarWriterChmod     = 0600
)

type TarWriter struct {
	tarFile    *os.File
	gzipWriter *gzip.Writer
	tarWriter  *tar.Writer
	buffers    map[string]*[]byte
	tarPath    string
}

func (t *TarWriter) initializedTarFile(ctx context.Context, directoryOutput string, resName string) error {
	t.tarPath = path.Join(directoryOutput, fmt.Sprintf("%s%s", resName, TarWriterExtension))

	log.I.Debugf("Creating tar file %s", t.tarPath)
	err := os.MkdirAll(filepath.Dir(t.tarPath), WriterDirChmod)
	if err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}
	t.tarFile, err = os.Create(t.tarPath)
	if err != nil {
		return err
	}

	return nil
}

func (f *TarWriter) GetOutputPath() string {
	return f.tarPath
}

func (t *TarWriter) Initialize(ctx context.Context, path string, resName string) error {
	err := t.initializedTarFile(ctx, path, resName)
	if err != nil {
		return err
	}

	t.gzipWriter = gzip.NewWriter(t.tarFile)
	t.tarWriter = tar.NewWriter(t.gzipWriter)
	t.buffers = make(map[string]*[]byte)

	return nil
}

// Write function writes the Kubernetes object to a buffer
// All buffer are stored in a map which is flushed at the end of every type processed
func (t *TarWriter) Write(ctx context.Context, object []byte, filePath string) error {
	span, _ := tracer.StartSpanFromContext(ctx, span.CollectorTarWriterWrite, tracer.Measured())
	span.SetTag(tag.CollectorFilePath, filePath)
	defer span.Finish()
	buf, ok := t.buffers[filePath]
	if ok {
		*buf = append(*buf, object...)
	} else {
		buf = &object
		t.buffers[filePath] = buf
	}
	return nil
}

// Flush function flushes all kubernetes object from the buffers to the tar file
func (t *TarWriter) Flush(ctx context.Context) error {
	log.I.Debug("Flushing writers")
	span, _ := tracer.StartSpanFromContext(ctx, span.CollectorTarWriterFlush, tracer.Measured())
	defer span.Finish()
	for path, data := range t.buffers {
		header := &tar.Header{
			Name: path,
			Mode: TarWriterChmod,
			Size: int64(len(*data)),
		}

		if err := t.tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if _, err := t.tarWriter.Write(*data); err != nil {
			return err
		}
	}

	return nil
}

// Close all the handler used to write the tar file
// Need to be closed only when all assets are dumped
func (t *TarWriter) Close(ctx context.Context) error {
	var err error
	log.I.Debug("Closing handlers for tar")
	span, _ := tracer.StartSpanFromContext(ctx, span.CollectorTarWriterClose, tracer.Measured())
	defer span.Finish()
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