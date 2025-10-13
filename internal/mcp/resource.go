package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// RegisterBaseResources exposes static resources (guidelines) to the MCP server.
func RegisterBaseResources(s ToolAdder, guidelinesContent string) error {
	resource := mcp.Resource{Name: "guidelines/scg"}
	handler := func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      req.Params.URI,
				MIMEType: "text/markdown",
				Text:     guidelinesContent,
			},
		}, nil
	}
	if err := s.AddResource(resource, handler); err != nil {
		return err
	}
	return nil
}
