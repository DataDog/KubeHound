package gremlin

import (
	"testing"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/ingestion"
	"github.com/DataDog/KubeHound/pkg/config"
)

func TestIngestion_List(t *testing.T) {
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

	repo := Ingestions(conn)

	ingestions, err := repo.List(t.Context(), ingestion.ListFilter{})
	if err != nil {
		t.Fatal(err)
	}

	for _, ingestion := range ingestions {
		t.Logf("Ingestion: %v", ingestion)
	}
}

func TestIngestion_GetEdgeCountPerClasses(t *testing.T) {
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

	repo := Ingestions(conn)

	edgeCounts, err := repo.GetEdgeCountPerClasses(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	for label, count := range edgeCounts {
		t.Logf("Edge count for %s: %d", label, count)
	}
}

func TestIngestion_GetVertexCountPerClasses(t *testing.T) {
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

	repo := Ingestions(conn)

	vertexCounts, err := repo.GetVertexCountPerClasses(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Vertex counts: %v", vertexCounts)
}
