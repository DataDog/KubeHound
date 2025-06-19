package llm

import (
	"fmt"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/container"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/ingestion"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/llm/prompt"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/llm/tool"
	"github.com/mark3labs/mcp-go/server"
)

// Server holds implementation details for a MCP server.
type Server struct {
	apiServer *server.MCPServer
}

// NewMCPServer rturns a MCP-compatible server instance.
func NewMCPServer(ingestions ingestion.Reader, containers container.Reader) (*Server, error) {
	// Prepare an MCP protocol compliant server.
	apiServer := server.NewMCPServer(
		"KubeHound MCP",
		"v0.0.1",
		server.WithLogging(),
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
	)

	// Register tools.
	if err := tool.RegisterKubeHoundTools(apiServer, ingestions, containers); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	// Register prompts.
	prompt.RegisterKubeHoundPrompts(apiServer)

	// Wrap the MCP server instance to allow user to select which exposition
	// protocol they want to use.
	return &Server{
		apiServer: apiServer,
	}, nil
}

// ServeStdio runs the MCP service over Stdio.
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.apiServer)
}

// ServeSse runs the MCP service over HTTP/SSE.
func (s *Server) ServeSse(publicHost string, port int) *server.SSEServer {
	return server.NewSSEServer(s.apiServer, server.WithBaseURL(fmt.Sprintf("http://%s:%d", publicHost, port)))
}
