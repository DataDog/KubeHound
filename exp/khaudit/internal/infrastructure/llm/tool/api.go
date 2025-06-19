package tool

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// Tool represents MCP Tool builder contract.
type Tool interface {
	// Manifest returns the tool registration manifest.
	Manifest() mcp.Tool
	// Hanlder is the tool request handler.
	Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}
