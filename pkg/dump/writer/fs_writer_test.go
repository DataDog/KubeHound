package writer

import (
	"encoding/json"
	"io"
	"path"
	"reflect"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	discoveryv1 "k8s.io/api/discovery/v1"
)

func TestFSWriter_WriteFile(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	fileNameK8sObject := collector.EndpointPath
	dummyNamespace := "namespace1"
	dummyK8sObject := []*discoveryv1.EndpointSlice{
		collector.FakeEndpoint("name1", "namespace1", []int32{int32(80)}),
		collector.FakeEndpoint("name2", "namespace1", []int32{int32(443)}),
	}

	writer, err := NewFSWriter(ctx)
	if err != nil {
		t.Fatalf("failed to create file writer: %v", err)
	}

	vfsResourcePath := path.Join(dummyNamespace, fileNameK8sObject)

	jsonData, err := json.Marshal(dummyK8sObject)
	if err != nil {
		t.Fatalf("failed to marshal Kubernetes object: %v", err)
	}

	err = writer.WriteFile(ctx, vfsResourcePath, jsonData)
	if err != nil {
		t.Fatalf("write %s: %v", reflect.TypeOf(dummyK8sObject), err)
	}

	file, err := writer.vfs.Open(vfsResourcePath)
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

	if !reflect.DeepEqual(processedDummyK8sObject, dummyK8sObject) {
		t.Fatalf("expected %v, got %v", processedDummyK8sObject, readK8sObject)
	}

	err = writer.Flush(ctx)
	if err != nil {
		t.Fatalf("failed to flush: %v", err)
	}
	err = writer.Close(ctx)
	if err != nil {
		t.Fatalf("failed to close: %v", err)
	}
}
