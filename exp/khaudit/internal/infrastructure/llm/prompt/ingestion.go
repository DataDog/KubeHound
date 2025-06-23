package prompt

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterKubeHoundPrompts registers all MCP prompts.
func RegisterKubeHoundPrompts(srv *server.MCPServer) {
	srv.AddPrompt(
		mcp.NewPrompt(
			"kh_list_ingestions",
			mcp.WithPromptDescription(`List currently imported dataset in the KubeHound database.`),
		),
		func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			var messages []mcp.PromptMessage

			messages = append(messages, mcp.NewPromptMessage(
				mcp.RoleAssistant,
				mcp.NewTextContent(
					`Returns a list of all the ingestions currently imported associated with a cluster.
The ingestion id, called RunID, is a unique identifier for a dataset that has been 
imported into KubeHound.

Enumerate all the ingestions for a given cluster and return a table with the following columns:
- RunID
- Cluster

Remember the index so that the user can use it to select an active ingestion for the following queries.`,
				),
			))

			return mcp.NewGetPromptResult("ingestions", messages), nil
		},
	)
}
