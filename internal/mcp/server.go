package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// StdioServer is an MCP server running over stdin/stdout.
type StdioServer struct {
	s *server.MCPServer
}

// ToolHandler is the function signature for a tool handler.
type ToolHandler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)

// ResourceHandler is the function signature for a resource handler.
type ResourceHandler func(context.Context, mcp.ReadResourceRequest) ([]mcp.ResourceContents, error)

// ToolAdder abstracts the AddTool and AddResource methods of an MCP server.
type ToolAdder interface {
	AddTool(tool mcp.Tool, handler ToolHandler) error
	AddResource(resource mcp.Resource, handler ResourceHandler) error
}

// NewStdioServer creates a new MCP server configured for stdio.
func NewStdioServer(name, version string) *StdioServer {
	s := server.NewMCPServer(name, version)
	return &StdioServer{s: s}
}

// AddTool adds a tool to the MCP server.
func (s *StdioServer) AddTool(tool mcp.Tool, handler ToolHandler) error {
	s.s.AddTool(tool, server.ToolHandlerFunc(handler))
	return nil
}

// AddResource adds a resource to the MCP server.
func (s *StdioServer) AddResource(resource mcp.Resource, handler ResourceHandler) error {
	s.s.AddResource(resource, server.ResourceHandlerFunc(handler))
	return nil
}

// Start starts the stdio listener and blocks until the context is canceled.
func (s *StdioServer) Start(ctx context.Context) error {
	errC := make(chan error, 1)
	go func() {
		// server.ServeStdio is a blocking call
		errC <- server.ServeStdio(s.s)
	}()

	select {
	case err := <-errC:
		return fmt.Errorf("mcp server failed: %w", err)
	case <-ctx.Done():
		// There is no graceful shutdown for stdio server in the library.
		// We just exit.
		return ctx.Err()
	}
}
