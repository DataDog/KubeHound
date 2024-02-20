package writer

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

const (
	WriterDirChmod = 0700
)

type CollectorWriter interface {
	Initialize(context.Context, string, string) error
	Write(context.Context, []byte, string) error
	Flush(context.Context) error
	Close(context.Context) error
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
