package writer

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	discoveryv1 "k8s.io/api/discovery/v1"
)

func TestTarWriter_Write(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tmpTarFileDir, err := os.MkdirTemp("/tmp/", "kh-unit-tests-*")
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	tmpTarExtractDir, err := os.MkdirTemp("/tmp/", "kh-unit-tests-*")
	if err != nil {
		log.I.Fatalf(err.Error())
	}

	// Constructing a buffer of Endpoints objects in different namespaces/files
	tarBundle := make(map[string]any)

	fileNameK8sObject := collector.EndpointPath
	dummyNamespace1 := "namespace1"
	dummyK8sObject1 := []*discoveryv1.EndpointSlice{
		collector.FakeEndpoint("name1", "namespace1", []int32{int32(80)}),
		collector.FakeEndpoint("name2", "namespace1", []int32{int32(443)}),
	}
	vfsResourcePath1 := path.Join(dummyNamespace1, fileNameK8sObject)
	tarBundle[vfsResourcePath1] = dummyK8sObject1

	dummyNamespace2 := "namespace2"
	dummyK8sObject2 := []*discoveryv1.EndpointSlice{
		collector.FakeEndpoint("name1", "namespace2", []int32{int32(80)}),
		collector.FakeEndpoint("name2", "namespace2", []int32{int32(443)}),
	}
	vfsResourcePath2 := path.Join(dummyNamespace2, fileNameK8sObject)
	tarBundle[vfsResourcePath2] = dummyK8sObject2

	writer, err := NewTarWriter(ctx, tmpTarFileDir, fileNameK8sObject)
	if err != nil {
		t.Fatalf("failed to create file writer: %v", err)
	}

	for vfsResourcePath, dummyK8sObject := range tarBundle {
		jsonData, err := json.Marshal(dummyK8sObject)
		if err != nil {
			t.Fatalf("failed to marshal Kubernetes object: %v", err)
		}
		err = writer.Write(ctx, jsonData, vfsResourcePath)
		if err != nil {
			t.Fatalf("write %s: %v", reflect.TypeOf(dummyK8sObject), err)
		}
	}

	writer.Flush(ctx)
	writer.Close(ctx)

	dryRun := false
	err = puller.ExtractTarGz(dryRun, writer.OutputPath(), tmpTarExtractDir, config.DefaultMaxArchiveSize)
	if err != nil {
		t.Fatalf("failed to extract tar.gz: %v", err)
	}

	countFileSuccess := 0
	err = filepath.Walk(tmpTarExtractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Fatalf("failed to walk path: %v", err)
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			t.Fatalf("failed reading file: %v", err)
		}
		defer file.Close()
		readK8sObject, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("failed reading file: %v", err)
		}

		var processedDummyK8sObject []*discoveryv1.EndpointSlice
		err = json.Unmarshal(readK8sObject, &processedDummyK8sObject)
		if err != nil {
			t.Fatalf("failed to unmarshal Kubernetes object: %v", err)
		}

		// reformating the relative path from the virtual filesystem
		vfsResourcePath := strings.ReplaceAll(tmpTarExtractDir, path, "")
		if reflect.DeepEqual(processedDummyK8sObject, tarBundle[vfsResourcePath]) {
			t.Fatalf("expected %v, got %v", processedDummyK8sObject, readK8sObject)
		}
		countFileSuccess++

		err = os.Remove(file.Name())
		if err != nil {
			t.Fatalf("failed to remove file: %v", err)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk path: %v", err)
	}

	if countFileSuccess != len(tarBundle) {
		t.Fatalf("expected %d, got %d", len(tarBundle), countFileSuccess)
	}

}
