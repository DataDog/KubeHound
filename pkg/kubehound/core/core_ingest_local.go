package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/dump"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

func CoreLocalIngest(ctx context.Context, khCfg *config.KubehoundConfig, resultPath string) error {
	l := log.Logger(ctx)
	// Using the collector config to ingest the data
	khCfg.Collector.Type = config.CollectorTypeFile

	// Treating by default as data not compressed (directory of the results)
	khCfg.Collector.File.Directory = resultPath

	// Checking dynamically if the data is being compressed
	compress, err := puller.IsTarGz(resultPath, khCfg.Ingestor.MaxArchiveSize)
	if err != nil {
		return err
	}
	metadataFilePath := filepath.Join(resultPath, collector.MetadataPath)
	if compress {
		tmpDir, err := os.MkdirTemp("/tmp/", "kh-local-ingest-*")
		if err != nil {
			return fmt.Errorf("creating temp dir: %w", err)
		}
		// Resetting the directory to the temp directory used to extract the data
		khCfg.Collector.File.Directory = tmpDir
		dryRun := false
		err = puller.ExtractTarGz(dryRun, resultPath, tmpDir, config.DefaultMaxArchiveSize)
		if err != nil {
			return err
		}
		metadataFilePath = filepath.Join(tmpDir, collector.MetadataPath)
	}
	// Getting the metadata from the metadata file
	md, err := dump.ParseMetadata(ctx, metadataFilePath)
	if err != nil {
		// Backward Compatibility: not returning error for now as the metadata feature is new
		l.Warn("no metadata has been parsed (old dump format from v1.4.0 or below do not embed metadata information)", log.ErrorField(err))
	} else {
		khCfg.Dynamic.ClusterName = md.ClusterName
	}

	// Backward Compatibility: Extracting the metadata from the path or input args
	// If the cluster name is not provided by the command args (deprecated flag), we try to get it from the path
	if khCfg.Dynamic.ClusterName == "" {
		dumpMetadata, err := dump.ParsePath(ctx, resultPath)
		if err != nil {
			l.Warnf("parsing path for metadata", log.ErrorField(err))
		}
		khCfg.Dynamic.ClusterName = dumpMetadata.Metadata.ClusterName
	}

	return CoreLive(ctx, khCfg)
}
