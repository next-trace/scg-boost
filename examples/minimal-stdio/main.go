package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/next-trace/scg-boost/boost"
)

// mapConfig is a simple in-memory implementation of types.SafeConfig.
type mapConfig struct{ m map[string]string }

func (c mapConfig) Get(key string) (string, bool) { v, ok := c.m[key]; return v, ok }
func (c mapConfig) List(prefix string) map[string]string {
	out := make(map[string]string)
	for k, v := range c.m {
		if prefix == "" || strings.HasPrefix(k, prefix) {
			out[k] = v
		}
	}
	return out
}

// loggerAdapter adapts the standard log.Logger to the types.Logger interface.
type loggerAdapter struct{ l *log.Logger }

func (a loggerAdapter) Debug(msg string, fields map[string]any) {
	a.l.Printf("DEBUG: %s %v", msg, fields)
}
func (a loggerAdapter) Error(msg string, fields map[string]any) {
	a.l.Printf("ERROR: %s %v", msg, fields)
}

const (
	appName    = "scg-boost-host"
	appVersion = "0.1.0"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	stdLogger := log.New(os.Stderr, fmt.Sprintf("[%s] ", appName), log.LstdFlags|log.Lmsgprefix)
	logger := loggerAdapter{l: stdLogger}

	cfg := mapConfig{
		m: map[string]string{
			"DB_HOST":     "localhost",
			"DB_PORT":     "5432",
			"DB_USER":     "admin",
			"DB_PASS":     "***redacted***",
			"API_KEY":     "***redacted***",
			"ENVIRONMENT": "development",
		},
	}

	opts := []boost.Option{
		boost.WithName(appName),
		boost.WithVersion(appVersion),
		boost.WithLogger(logger),
		boost.WithConfig(cfg),
	}

	server, err := boost.New(opts...)
	if err != nil {
		stdLogger.Fatalf("FATAL: failed to create server: %v", err)
	}

	stdLogger.Println("Starting MCP server over stdio...")
	if _, err := server.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		stdLogger.Printf("ERROR: server start failed: %v", err)
		return
	}

	<-ctx.Done()
	stdLogger.Println("Shutting down...")
}
