package tool

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/ingestion"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func listRunIDs(ingestions ingestion.Reader) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"kh_list_runids",
			mcp.WithDescription(`List currently imported dataset.

Returns a list of all the runIDs currently imported associated with a cluster.
The runID is a unique identifier for a dataset that has been imported into KubeHound.
`),
		),
		Handler: func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// List known ingestions.
			results, err := ingestions.List(ctx, ingestion.ListFilter{})
			if err != nil {
				return mcp.NewToolResultError("unable to enumerate runID"), err
			}

			// Format a tool response.
			var response string
			for _, r := range results {
				response += fmt.Sprintf("Cluster: %s\nRunID: %s\n\n", r.Cluster, r.RunID)
			}

			return mcp.NewToolResultText(response), nil
		},
	}
}
