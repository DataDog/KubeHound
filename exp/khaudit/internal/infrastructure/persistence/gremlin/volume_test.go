package gremlin

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/volume"
	"github.com/DataDog/KubeHound/pkg/config"
)

func TestVolumeRepository_GetVolumes(t *testing.T) {
	if config.IsCI() {
		t.Skip("skipping in CI")
		return
	}

	conn, err := NewConnection(Config{
		Endpoint: "ws://localhost:8182/gremlin",
		AuthMode: "plain",
	})
	if err != nil {
		t.Fatal(err)
	}

	repo := Volumes(conn)

	volumes, err := repo.GetVolumes(context.Background(), "01jq92qsx1f9ndvk48qm9hp7k4", volume.Filter{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Volumes: %v", volumes)
}

func TestVolumeRepository_GetMountedHostPaths(t *testing.T) {
	if config.IsCI() {
		t.Skip("skipping in CI")
		return
	}

	conn, err := NewConnection(Config{
		Endpoint: "ws://localhost:8182/gremlin",
		AuthMode: "plain",
	})
	if err != nil {
		t.Fatal(err)
	}

	repo := Volumes(conn)

	mountedHostPaths, err := repo.GetMountedHostPaths(context.Background(), "01jqnvmdt2rdd19fjm2skzkc3j", volume.Filter{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Mounted host paths: %v", mountedHostPaths)
}
