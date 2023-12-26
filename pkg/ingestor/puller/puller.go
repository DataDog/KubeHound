package puller

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

var (
	ArchiveName = "archive.tar.gz"
	BasePath    = "/tmp/kubehound"
)

type DataPuller interface {
	Pull(ctx context.Context, clusterName string, runID string) (string, error)
	Extract(ctx context.Context, archivePath string) error
	Close(ctx context.Context, basePath string) error
}

func FormatArchiveKey(clusterName string, runID string) string {
	return strings.Join([]string{clusterName, runID, ArchiveName}, "/")
}

// checkSanePath just to make sure we don't delete or overwrite somewhere where we are not supposed to
func CheckSanePath(path string) error {
	if path == "/" || path == "" || !strings.HasPrefix(path, BasePath) {
		return fmt.Errorf("Invalid path provided: %q", path)
	}
	return nil
}

func ExtractTarGz(gzipFileReader io.Reader, basePath string) error {
	uncompressedStream, err := gzip.NewReader(gzipFileReader)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(uncompressedStream)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch header.Typeflag {
		// Handle sub folder containing namespaces
		case tar.TypeDir:
			err := os.Mkdir(filepath.Join(basePath, header.Name), 0755)
			if err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(filepath.Join(basePath, header.Name))
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				return err
			}
		default:
			log.I.Info("unsupported archive item (not a folder, not a regular file): ", header.Typeflag)
		}
	}
	return nil
}
