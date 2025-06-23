package gremlin

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/container"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/kubehound"
	"github.com/DataDog/KubeHound/pkg/config"
)

func asRef[T any](v T) *T {
	return &v
}

func TestContainer_CountByNamespaces(t *testing.T) {
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

	repo := Containers(conn)

	resultChan := make(chan container.NamespaceAggregation, 2000)
	err = repo.CountByNamespaces(context.Background(), "test.cluster.local", "01jnh21qtrt41ddmgyfpfm29qj", container.NamespaceAggregationFilter{}, resultChan)
	if err != nil {
		t.Fatal(err)
	}
	close(resultChan)

	for res := range resultChan {
		t.Logf("Namespace: %s, Count: %d", res.Namespace, res.Count)
	}
}

func TestContainer_GetAttackPathProfiles(t *testing.T) {
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

	repo := Containers(conn)

	paths, err := repo.GetAttackPathProfiles(context.Background(), "test.cluster.local", "01jnh21qtrt41ddmgyfpfm29qj", container.AttackPathFilter{})
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range paths {
		t.Logf("Attack path: %v", path)
	}
}

func TestContainer_GetVulnerables(t *testing.T) {
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

	repo := Containers(conn)

	resultChan := make(chan container.Container, 2000)
	err = repo.GetVulnerables(context.Background(), "test.cluster.local", "01jnh21qtrt41ddmgyfpfm29qj", container.AttackPathFilter{
		Namespace: asRef("default"),
	}, resultChan)
	if err != nil {
		t.Fatal(err)
	}

	close(resultChan)

	for res := range resultChan {
		t.Logf("Container: %v", res)
	}
}

func TestContainer_GetAttackPaths(t *testing.T) {
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

	repo := Containers(conn)

	resultChan := make(chan kubehound.AttackPath, 2000)
	err = repo.GetAttackPaths(context.Background(), "test.cluster.local", "01jp28sxeagj7mbsdtkfaf4c9t", container.AttackPathFilter{
		Namespace: asRef("default"),
		App:       asRef("toolbox"),
	}, resultChan)
	if err != nil {
		t.Fatal(err)
	}

	close(resultChan)

	for res := range resultChan {
		t.Logf("Container: %v", res)
	}
}
