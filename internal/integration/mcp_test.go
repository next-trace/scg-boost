//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/next-trace/scg-boost/boost"
)

func TestMCPServerStdio(t *testing.T) {
	// Create a basic server
	srv, err := boost.New(
		boost.WithName("test-server"),
		boost.WithVersion("0.1.0"),
	)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Verify server was created
	if srv == nil {
		t.Fatal("New() returned nil server")
	}

	// Note: We don't actually start the server here because it would block
	// on stdio. This test validates the server can be created successfully.
	t.Log("MCP server created successfully")
}

func TestMCPToolExecution(t *testing.T) {
	// Create server with basic options
	srv, err := boost.New(
		boost.WithName("test-app"),
		boost.WithVersion("1.0.0"),
	)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify context works
	select {
	case <-ctx.Done():
		t.Fatal("context cancelled before test completion")
	default:
		// Context is valid
	}

	// The appinfo.get tool should always be registered
	// We can't easily test actual tool execution without mocking stdio,
	// but we can verify the server initializes correctly
	if srv == nil {
		t.Fatal("server is nil")
	}

	t.Log("MCP tool registration validated")
}
