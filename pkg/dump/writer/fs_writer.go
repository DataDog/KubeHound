package writer

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/spf13/afero"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	FSWriterChmod = 0600
)

// The FileWriter uses a map of buffers to write data to files
// Each file has its own buffer to optimize IO calls
type FSWriter struct {
	mu  sync.Mutex
	vfs afero.Fs
}

func NewFSWriter(ctx context.Context) (*FSWriter, error) {
	return &FSWriter{
		vfs: afero.NewMemMapFs(),
	}, nil
}

// Write function writes the Kubernetes object to a buffer
// All buffer are stored in a map which is flushed at the end of every type processed
func (f *FSWriter) WriteFile(ctx context.Context, pathObj string, k8sObj []byte) error {
	l := log.Logger(ctx)
	l.Debug("Writing to file", log.String(log.FieldPathKey, pathObj))
	f.mu.Lock()
	defer f.mu.Unlock()

	// Create directories if they do not exist
	err := f.vfs.MkdirAll(filepath.Dir(pathObj), WriterDirMod)
	if err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	err = afero.WriteFile(f.vfs, pathObj, k8sObj, FSWriterChmod)
	if err != nil {
		return fmt.Errorf("failed to write JSON data to file: %w", err)
	}

	return nil
}

// No flush needed for the file writer as we are flushing the buffer at every write
func (f *FSWriter) Flush(ctx context.Context) error {
	span, _ := span.SpanRunFromContext(ctx, span.DumperWriterFlush)
	span.SetTag(tag.DumperWriterTypeTag, TarTypeTag)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	return nil
}

func (f *FSWriter) Close(ctx context.Context) error {
	return nil
}
