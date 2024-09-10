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
	// TODO: migrate to Metadata.json file instead of relying on the file path
	// See PR https://github.com/DataDog/KubeHound/pull/247
	dumpMetadata, err := dump.ParsePath(resultPath)
	if err != nil {
		log.I.Warn("parsing path failed", err)
	}
	khCfg.Dynamic.ClusterName = dumpMetadata.ClusterName
	khCfg.Dynamic.RunID, err = config.LoadRunID(dumpMetadata.RunID)
	if err != nil {
		log.I.Warn("parsing run id from file path", err)
	}
	khCfg.Ingestor.ClusterName = dumpMetadata.ClusterName
	khCfg.Ingestor.RunID = dumpMetadata.RunID
	// Using the collector config to ingest the data
	khCfg.Collector.File.ClusterName = dumpMetadata.ClusterName
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
		return err
	}
	khCfg.Collector.File.ClusterName = md.ClusterName

	return CoreLive(ctx, khCfg)
}
