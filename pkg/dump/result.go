package dump

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/DataDog/KubeHound/pkg/collector"
)

type DumpResult struct {
	isDir     bool
	extension string
	Metadata  collector.Metadata
}

const (
	DumpResultClusterNameRegex = `([A-Za-z0-9\.\-_]+)`
	DumpResultRunIDRegex       = `([a-z0-9]{26})`
	DumpResultExtensionRegex   = `\.?([a-z0-9\.]+)?`
	DumpResultPrefix           = "kubehound_"

	DumpResultTarWriterExtension = "tar.gz"
)

func NewDumpResult(clusterName, runID string, isCompressed bool) (*DumpResult, error) {
	dumpResult := &DumpResult{
		Metadata: collector.Metadata{
			ClusterName: clusterName,
			RunID:       runID,
		},
		isDir: true,
	}
	if isCompressed {
		dumpResult.Compressed()
	}

	err := dumpResult.Validate()
	if err != nil {
		return nil, err
	}

	return dumpResult, nil
}

func (i *DumpResult) Validate() error {
	re := regexp.MustCompile(DumpResultClusterNameRegex)
	if !re.MatchString(i.Metadata.ClusterName) {
		return fmt.Errorf("Invalid clustername: %q", i.Metadata.ClusterName)
	}

	matches := re.FindStringSubmatch(i.Metadata.ClusterName)
	if len(matches) == 2 && matches[1] != i.Metadata.ClusterName {
		return fmt.Errorf("Invalid clustername: %q", i.Metadata.ClusterName)
	}

	re = regexp.MustCompile(DumpResultRunIDRegex)
	if !re.MatchString(i.Metadata.RunID) {
		return fmt.Errorf("Invalid runID: %q", i.Metadata.RunID)
	}

	return nil
}

func (i *DumpResult) Compressed() {
	i.isDir = false
	i.extension = DumpResultTarWriterExtension
}

// ./<clusterName>/kubehound_<clusterName>_<run_id>
func (i *DumpResult) GetFullPath() string {
	filename := i.GetFilename()

	return path.Join(i.Metadata.ClusterName, filename)
}

func (i *DumpResult) GetFilename() string {
	filename := fmt.Sprintf("%s%s_%s", DumpResultPrefix, i.Metadata.ClusterName, i.Metadata.RunID)
	if i.isDir {
		return filename
	}

	return fmt.Sprintf("%s.%s", filename, i.extension)
}

func GetDumpMetadata(ctx context.Context, metadataFilePath string) (collector.Metadata, error) {
	var metadata collector.Metadata

	bytes, err := os.ReadFile(metadataFilePath)
	if err != nil {
		return metadata, fmt.Errorf("read file %s: %w", metadataFilePath, err)
	}

	if len(bytes) == 0 {
		return metadata, nil
	}

	err = json.Unmarshal(bytes, &metadata)
	if err != nil {
		return metadata, fmt.Errorf("unmarshalling %T in %s: %w", metadata, metadataFilePath, err)
	}

	return metadata, nil
}
