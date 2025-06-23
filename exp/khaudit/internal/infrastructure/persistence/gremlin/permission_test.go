package gremlin

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/permission"
	"github.com/DataDog/KubeHound/pkg/config"
)

func TestPermissionRepository_GetReachablePodCountPerNamespace(t *testing.T) {
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

	repo := Permissions(conn)

	namespaceCounts, err := repo.GetReachablePodCountPerNamespace(context.Background(), "01jq6drwavcfzbpaaab4v21f5s")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Namespace counts: %v", namespaceCounts)
}

func TestPermissionRepository_GetKubectlExecutablePodCount(t *testing.T) {
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

	repo := Permissions(conn)

	podCount, err := repo.GetKubectlExecutablePodCount(context.Background(), "01jq6drwavcfzbpaaab4v21f5s", "employees")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Kubectl executable pod count: %d", podCount)
}

func TestPermissionRepository_GetExposedPodCountPerNamespace(t *testing.T) {
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

	repo := Permissions(conn)

	exposedPodCounts, err := repo.GetExposedPodCountPerNamespace(context.Background(), "01jq6drwavcfzbpaaab4v21f5s", "employees")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Exposed pod counts: %v", exposedPodCounts)
}

func TestPermissionRepository_GetKubectlExecutableGroupsForNamespace(t *testing.T) {
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

	repo := Permissions(conn)

	groups, err := repo.GetKubectlExecutableGroupsForNamespace(context.Background(), "01jq6drwavcfzbpaaab4v21f5s", "apm-driveline")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Kubectl executable groups: %v", groups)
}

func TestPermissionRepository_GetExposedNamespacePods(t *testing.T) {
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

	repo := Permissions(conn)

	groups, err := repo.GetExposedNamespacePods(context.Background(), "01jq6drwavcfzbpaaab4v21f5s", "apm-driveline", "employees", permission.ExposedPodFilter{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Kubectl executable groups: %v", groups)
}
