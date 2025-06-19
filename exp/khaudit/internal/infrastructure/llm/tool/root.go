package tool

import (
	"errors"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/container"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/ingestion"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterKubeHoundTools registers all MCP tools.
func RegisterKubeHoundTools(srv *server.MCPServer, ingestions ingestion.Reader, containers container.Reader) error {
	if srv == nil {
		return errors.New("server is nil")
	}

	// Register tools.
	srv.AddTools(
		listRunIDs(ingestions),
		cePathProfiles(containers),
		vulnerableImages(containers),
		ceAttackPaths(containers),
	)

	return nil
}
