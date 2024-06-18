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

//go:generate mockery --name DataPuller --output mocks --case underscore --filename mock_puller.go --with-expecter
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

func IsTarGz(filePath string, maxArchiveSize int64) (bool, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false, fmt.Errorf("stat %s: %w", filePath, err)
	}

	switch mod := fileInfo.Mode(); {
	case mod.IsDir():
		return false, nil
	case mod.IsRegular():
		dryRun := true
		err = ExtractTarGz(dryRun, filePath, "/tmp", maxArchiveSize)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	return false, fmt.Errorf("file type not supported")
}

func ExtractTarGz(checkOnly bool, archivePath string, basePath string, maxArchiveSize int64) error { //nolint:gocognit
	gzipFileReader, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer gzipFileReader.Close()

	uncompressedStream, err := gzip.NewReader(gzipFileReader)
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()

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

		switch header.Typeflag {
		// Handle sub folder containing namespaces
		case tar.TypeDir:
			if checkOnly {
				continue
			}
			err := mkdirIfNotExists(cleanPath)
			if err != nil {
				return err
			}
		case tar.TypeReg:
			if checkOnly {
				continue
			}
			err := mkdirIfNotExists(filepath.Dir(cleanPath))
			if err != nil {
				return err
			}
			outFile, err := os.Create(cleanPath)
			if err != nil {
				return fmt.Errorf("creating file %s: %w", cleanPath, err)
			}
			defer outFile.Close()
			// We don't really have an upper limit of archive size and adding a limited writer is not trivial without importing
			// a third party library (like our internal secure lib)
			_, err = io.Copy(outFile, tarReader) //nolint:gosec
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return fmt.Errorf("archive size exceeds the limit: %d: %w", maxArchiveSize, err)
			}
			if err != nil {
				return fmt.Errorf("copying file %s: %w", cleanPath, err)
			}
		default:
			log.I.Info("unsupported archive item (not a folder, not a regular file): ", header.Typeflag)
		}
	}

	return nil
}

func mkdirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("mkdir %s: %w", path, err)
		}
	}

	return nil
}
