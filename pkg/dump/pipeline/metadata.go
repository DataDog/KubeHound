package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
)

type MetadataIngestor struct {
	buffer map[string]collector.Metadata
	writer writer.DumperWriter
}

func NewMetadataIngestor(ctx context.Context, dumpWriter writer.DumperWriter) *MetadataIngestor {
	return &MetadataIngestor{
		buffer: make(map[string]collector.Metadata),
		writer: dumpWriter,
	}
}

func (d *MetadataIngestor) DumpMetadata(ctx context.Context, metadata collector.Metadata) error {
	data := make(map[string]*collector.Metadata)
	data[collector.MetadataPath] = &metadata

	return dumpObj[*collector.Metadata](ctx, data, d.writer)

}
