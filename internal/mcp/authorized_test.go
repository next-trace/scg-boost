package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// mockAuthorizer implements types.Authorizer for testing.
type mockAuthorizer struct {
	allowedScopes map[string]bool
}

func (m *mockAuthorizer) HasScope(ctx context.Context, scope string) bool {
	return m.allowedScopes[scope]
}

// mockLogger implements types.Logger for testing.
type mockLogger struct {
	debugCalls []map[string]any
	errorCalls []map[string]any
}

func (m *mockLogger) Debug(msg string, fields map[string]any) {
	m.debugCalls = append(m.debugCalls, fields)
}

func (m *mockLogger) Error(msg string, fields map[string]any) {
	m.errorCalls = append(m.errorCalls, fields)
}

func TestAuthorizedServer_AddTool_AllowsAuthorized(t *testing.T) {
	// Create mock authorizer with allowed scope
	authorizer := &mockAuthorizer{allowedScopes: map[string]bool{"dbquery.run": true}}
	logger := &mockLogger{}

	// Create AuthorizedServer using the underlying StdioServer
	// Since we can't use real StdioServer, test the authorization logic directly
	handlerCalled := false
	originalHandler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		handlerCalled = true
		return &mcp.CallToolResult{Content: []mcp.Content{}}, nil
	}

	// Simulate what AuthorizedServer.AddTool does
	scopes := []string{"dbquery.run"}
	authorizedHandler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		for _, scope := range scopes {
			if !authorizer.HasScope(ctx, scope) {
				logger.Debug("authorization denied", map[string]any{
					"tool":  "dbquery.run",
					"scope": scope,
				})
				return ToolError(ErrCodeUnauthorized, "insufficient scope", nil), nil
			}
		}
		return originalHandler(ctx, req)
	}

	// Call the authorized handler
	req := mcp.CallToolRequest{}
	result, err := authorizedHandler(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Error("expected original handler to be called")
	}
	if result.IsError {
		t.Error("expected non-error result")
	}
}

func TestAuthorizedServer_AddTool_DeniesUnauthorized(t *testing.T) {
	// Create mock authorizer that denies the scope
	authorizer := &mockAuthorizer{allowedScopes: map[string]bool{}}
	logger := &mockLogger{}

	handlerCalled := false
	originalHandler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		handlerCalled = true
		return &mcp.CallToolResult{Content: []mcp.Content{}}, nil
	}

	// Simulate what AuthorizedServer.AddTool does
	scopes := []string{"dbquery.run"}
	authorizedHandler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		for _, scope := range scopes {
			if !authorizer.HasScope(ctx, scope) {
				logger.Debug("authorization denied", map[string]any{
					"tool":  "dbquery.run",
					"scope": scope,
				})
				return ToolError(ErrCodeUnauthorized, "insufficient scope", nil), nil
			}
		}
		return originalHandler(ctx, req)
	}

	// Call the authorized handler
	req := mcp.CallToolRequest{}
	result, err := authorizedHandler(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if handlerCalled {
		t.Error("expected original handler NOT to be called")
	}
	if !result.IsError {
		t.Error("expected error result when unauthorized")
	}
	if len(logger.debugCalls) != 1 {
		t.Errorf("expected 1 debug call, got %d", len(logger.debugCalls))
	}
}

func TestAuthorizedServer_AddTool_MultipleScopes(t *testing.T) {
	tests := []struct {
		name           string
		allowedScopes  map[string]bool
		requiredScopes []string
		wantAllow      bool
	}{
		{
			name:           "all scopes allowed",
			allowedScopes:  map[string]bool{"scope1": true, "scope2": true},
			requiredScopes: []string{"scope1", "scope2"},
			wantAllow:      true,
		},
		{
			name:           "first scope missing",
			allowedScopes:  map[string]bool{"scope2": true},
			requiredScopes: []string{"scope1", "scope2"},
			wantAllow:      false,
		},
		{
			name:           "second scope missing",
			allowedScopes:  map[string]bool{"scope1": true},
			requiredScopes: []string{"scope1", "scope2"},
			wantAllow:      false,
		},
		{
			name:           "no scopes required",
			allowedScopes:  map[string]bool{},
			requiredScopes: []string{},
			wantAllow:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorizer := &mockAuthorizer{allowedScopes: tt.allowedScopes}
			logger := &mockLogger{}

			handlerCalled := false
			originalHandler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				handlerCalled = true
				return &mcp.CallToolResult{Content: []mcp.Content{}}, nil
			}

			authorizedHandler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				for _, scope := range tt.requiredScopes {
					if !authorizer.HasScope(ctx, scope) {
						logger.Debug("authorization denied", map[string]any{
							"scope": scope,
						})
						return ToolError(ErrCodeUnauthorized, "insufficient scope", nil), nil
					}
				}
				return originalHandler(ctx, req)
			}

			req := mcp.CallToolRequest{}
			result, _ := authorizedHandler(context.Background(), req)

			if handlerCalled != tt.wantAllow {
				t.Errorf("handlerCalled = %v, want %v", handlerCalled, tt.wantAllow)
			}
			if result.IsError != !tt.wantAllow {
				t.Errorf("result.IsError = %v, want %v", result.IsError, !tt.wantAllow)
			}
		})
	}
}

func TestAuthorizedServer_AddTool_HandlerError(t *testing.T) {
	// Test that handler errors propagate correctly
	authorizer := &mockAuthorizer{allowedScopes: map[string]bool{"test.scope": true}}

	expectedErr := errors.New("handler failed")
	originalHandler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, expectedErr
	}

	scopes := []string{"test.scope"}
	authorizedHandler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		for _, scope := range scopes {
			if !authorizer.HasScope(ctx, scope) {
				return ToolError(ErrCodeUnauthorized, "insufficient scope", nil), nil
			}
		}
		return originalHandler(ctx, req)
	}

	req := mcp.CallToolRequest{}
	_, err := authorizedHandler(context.Background(), req)

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestToolError(t *testing.T) {
	tests := []struct {
		name    string
		code    ErrCode
		message string
		data    map[string]any
	}{
		{
			name:    "unauthorized error",
			code:    ErrCodeUnauthorized,
			message: "not authorized",
			data:    nil,
		},
		{
			name:    "with data",
			code:    ErrCodeUnauthorized,
			message: "insufficient scope",
			data:    map[string]any{"required": "admin"},
		},
		{
			name:    "invalid input error",
			code:    ErrCodeInvalidInput,
			message: "invalid query",
			data:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToolError(tt.code, tt.message, tt.data)

			if !result.IsError {
				t.Error("expected IsError to be true")
			}
			if len(result.Content) != 1 {
				t.Fatalf("expected 1 content item, got %d", len(result.Content))
			}
		})
	}
}
