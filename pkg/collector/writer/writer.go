package writer

import "context"

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
