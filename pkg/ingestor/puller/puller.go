package puller

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

type DataPuller interface {
	Pull(ctx context.Context, clusterName string, runID string) (string, error)
	Extract(ctx context.Context, archivePath string) error
	Close(ctx context.Context, basePath string) error
}

func FormatArchiveKey(clusterName string, runID string, archiveName string) string {
	return strings.Join([]string{clusterName, runID, archiveName}, "/")
}

// checkSanePath just to make sure we don't delete or overwrite somewhere where we are not supposed to
func CheckSanePath(path string, baseFolder string) error {
	if path == "/" || path == "" || !strings.HasPrefix(path, baseFolder) {
		return fmt.Errorf("Invalid path provided: %q", path)
	}

	return nil
}

// https://security.snyk.io/research/zip-slip-vulnerability
func sanitizeExtractPath(filePath string, destination string) error {
	destpath := filepath.Join(destination, filePath)
	if !strings.HasPrefix(destpath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("%s: illegal file path", filePath)
	}

	return nil
}

func ExtractTarGz(gzipFileReader io.Reader, basePath string, maxArchiveSize int64) error {
	uncompressedStream, err := gzip.NewReader(gzipFileReader)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(io.LimitReader(uncompressedStream, maxArchiveSize))
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		err = sanitizeExtractPath(basePath, header.Name)
		if err != nil {
			return err
		}
		cleanPath := filepath.Join(basePath, header.Name) //nolint:gosec // We check the path just above
		log.I.Infof("Extracting %s", cleanPath)

		switch header.Typeflag {
		// Handle sub folder containing namespaces
		case tar.TypeDir:
			err := os.Mkdir(cleanPath, os.ModePerm)
			if err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(cleanPath)
			if err != nil {
				return err
			}
			defer outFile.Close()
			// We don't really have an upper limit of archive size and adding a limited writer is not trivial without importing
			// a third party library (like our internal secure lib)
			_, err = io.Copy(outFile, tarReader) //nolint:gosec
			if err != nil {
				return err
			}
		default:
			log.I.Info("unsupported archive item (not a folder, not a regular file): ", header.Typeflag)
		}
	}

	return nil
}
