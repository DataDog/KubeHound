package tool

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/container"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func vulnerableImages(containers container.Reader) server.ServerTool {
	type args struct {
		Cluster   string  `json:"cluster"`
		RunID     string  `json:"runID"`
		Namespace string  `json:"namespace"`
		Team      *string `json:"team"`
		App       *string `json:"app"`
	}

	return server.ServerTool{
		Tool: mcp.NewTool(
			"kh_vulnerable_images",
			mcp.WithDescription("List container images that are affected by a container escape attack path."),
			mcp.WithString(
				"cluster",
				mcp.Required(),
				mcp.Description("Kubernetes cluster name"),
			),
			mcp.WithString(
				"runID",
				mcp.Required(),
				mcp.Description("KubeHound import identifier"),
			),
			mcp.WithString(
				"namespace",
				mcp.Required(),
				mcp.Description("Kubernetes namespace to start from"),
			),
			mcp.WithString(
				"team",
				mcp.Description("Team name"),
			),
			mcp.WithString(
				"app",
				mcp.Description("Application name"),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Parse the arguments.
			var args args
			if err := request.BindArguments(&args); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			filter := container.AttackPathFilter{
				Namespace: &args.Namespace,
			}
			filter.Team = args.Team
			filter.App = args.App

			// Run the query.
			resultChan := make(chan container.Container, 100)
			err := containers.GetVulnerables(ctx, args.Cluster, args.RunID, filter, resultChan)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Close the channel.
			close(resultChan)

			// Prepare a response.
			var response string
			for r := range resultChan {
				response += fmt.Sprintf("Cluster: %s\nRunID: %s\nNamespace: %s\nImage: %s\nApp: %s\nTeam: %s\n\n", r.Cluster, r.RunID, r.Namespace, r.Image, r.App, r.Team)
			}

			return mcp.NewToolResultText(response), nil
		},
	}
}
