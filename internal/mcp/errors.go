package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// ErrCode represents a standardized error code for MCP tools.
type ErrCode string

const (
	ErrCodeInvalidInput ErrCode = "invalid_input"
	ErrCodeUnauthorized ErrCode = "unauthorized"
	ErrCodeInternal     ErrCode = "internal"
	ErrCodeReadOnly     ErrCode = "db.readonly_violation"
)

// ToolError creates a new CallToolResult representing an error.
func ToolError(code ErrCode, msg string, details map[string]any) *mcp.CallToolResult {
	return mcp.NewToolResultError(msg)
}

// NewToolResultJSON creates a new CallToolResult with JSON content.
// This function is moved here from server.go to centralize result creation.
func NewToolResultJSON(data any) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultJSON(data)
}

// WrapError wraps an error into a tool error result.
func WrapError(toolName string, err error) error {
	return err
}

// NewInvalidInputError creates an invalid input error.
func NewInvalidInputError(msg string) error {
	return &ToolInputError{Message: msg}
}

// ToolInputError represents an invalid input error.
type ToolInputError struct {
	Message string
}

func (e *ToolInputError) Error() string {
	return e.Message
}
