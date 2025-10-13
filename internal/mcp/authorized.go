package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/next-trace/scg-boost/internal/security"
	"github.com/next-trace/scg-boost/types"
)

// AuthorizedServer wraps a StdioServer with authorization enforcement.
type AuthorizedServer struct {
	server     *StdioServer
	authorizer types.Authorizer
	logger     types.Logger
}

// NewAuthorizedServer creates a new authorized server wrapper.
func NewAuthorizedServer(server *StdioServer, authorizer types.Authorizer, logger types.Logger) *AuthorizedServer {
	return &AuthorizedServer{
		server:     server,
		authorizer: authorizer,
		logger:     logger,
	}
}

// AddTool adds a tool with authorization enforcement.
func (s *AuthorizedServer) AddTool(tool mcp.Tool, handler ToolHandler) error {
	toolName := tool.Name
	scopes := security.GetToolScopes(toolName)

	// Wrap handler with authorization check
	authorizedHandler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check each required scope
		for _, scope := range scopes {
			if !s.authorizer.HasScope(ctx, scope) {
				s.logger.Debug("authorization denied", map[string]any{
					"tool":  toolName,
					"scope": scope,
				})
				return ToolError(ErrCodeUnauthorized,
					fmt.Sprintf("insufficient scope for tool %s: requires %s", toolName, scope),
					map[string]any{"tool": toolName, "required_scope": scope},
				), nil
			}
		}

		// Authorization passed, call actual handler
		return handler(ctx, req)
	}

	return s.server.AddTool(tool, authorizedHandler)
}

// AddResource adds a resource to the underlying server.
func (s *AuthorizedServer) AddResource(resource mcp.Resource, handler ResourceHandler) error {
	return s.server.AddResource(resource, handler)
}

// Start starts the underlying server.
func (s *AuthorizedServer) Start(ctx context.Context) error {
	return s.server.Start(ctx)
}

// Unwrap returns the underlying StdioServer.
func (s *AuthorizedServer) Unwrap() *StdioServer {
	return s.server
}
