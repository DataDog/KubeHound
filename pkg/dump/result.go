package dump

import (
	"fmt"
	"path"
	"regexp"
)

type DumpResult struct {
	clusterName string
	RunID       string
	isDir       bool
	extension   string
}

const (
	DumpResultClusterNameRegex = `([A-Za-z0-9\.\-_]+)`
	DumpResultRunIDRegex       = `([a-z0-9]{26})`
	DumpResultExtensionRegex   = `\.?([a-z0-9\.]+)?`
	DumpResultPrefix           = "kubehound_"
	DumpResultFilenameRegex    = DumpResultPrefix + DumpResultClusterNameRegex + "_" + DumpResultRunIDRegex + DumpResultExtensionRegex
	DumpResultPathRegex        = DumpResultClusterNameRegex + "/" + DumpResultFilenameRegex

	DumpResultTarWriterExtension = "tar.gz"
)

func NewDumpResult(clusterName, runID string, compressed bool) (*DumpResult, error) {
	dumpResult := &DumpResult{
		clusterName: clusterName,
		RunID:       runID,
		isDir:       true,
	}
	if compressed {
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
	if !re.MatchString(i.clusterName) {
		return fmt.Errorf("Invalid clustername: %q", i.clusterName)
	}

	matches := re.FindStringSubmatch(i.clusterName)
	if len(matches) == 2 && matches[1] != i.clusterName {
		return fmt.Errorf("Invalid clustername: %q", i.clusterName)
	}

	re = regexp.MustCompile(DumpResultRunIDRegex)
	if !re.MatchString(i.RunID) {
		return fmt.Errorf("Invalid runID: %q", i.RunID)
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

	return path.Join(i.clusterName, filename)
}

func (i *DumpResult) GetFilename() string {
	filename := fmt.Sprintf("%s%s_%s", DumpResultPrefix, i.clusterName, i.RunID)
	if i.isDir {
		return filename
	}

	return fmt.Sprintf("%s.%s", filename, i.extension)
}

func ParsePath(path string) (*DumpResult, error) {
	// ./<clusterName>/kubehound_<clusterName>_<run_id>[.tar.gz]
	// re := regexp.MustCompile(`([a-z0-9\.\-_]+)/kubehound_([a-z0-9\.-_]+)_([a-z0-9]{26})\.?([a-z0-9\.]+)?`)
	re := regexp.MustCompile(DumpResultPathRegex)
	if !re.MatchString(path) {
		return nil, fmt.Errorf("Invalid path provided: %q", path)
	}

	matches := re.FindStringSubmatch(path)
	// The cluster name should match (parent dir and in the filename)
	if matches[1] != matches[2] {
		return nil, fmt.Errorf("Cluster name does not match in the path provided: %q", path)
	}

	clusterName := matches[1]
	runID := matches[3]
	extension := matches[4]

	compressed := false
	if extension != "" {
		compressed = true
	}
	result, err := NewDumpResult(clusterName, runID, compressed)
	if err != nil {
		return nil, err
	}

	return result, nil
}
