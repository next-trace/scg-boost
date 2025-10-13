package boost

import (
	"context"
	"testing"
	"time"
)

// mockLogger implements types.Logger for testing.
type mockLogger struct {
	debugCalls []logCall
	errorCalls []logCall
}

type logCall struct {
	msg    string
	fields map[string]any
}

func (m *mockLogger) Debug(msg string, fields map[string]any) {
	m.debugCalls = append(m.debugCalls, logCall{msg, fields})
}

func (m *mockLogger) Error(msg string, fields map[string]any) {
	m.errorCalls = append(m.errorCalls, logCall{msg, fields})
}

// mockAuthorizer implements types.Authorizer for testing.
type mockAuthorizer struct {
	scopes map[string]bool
}

func (m *mockAuthorizer) HasScope(ctx context.Context, scope string) bool {
	return m.scopes[scope]
}

func TestNew_DefaultOptions(t *testing.T) {
	srv, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if srv == nil {
		t.Fatal("New() returned nil server")
	}

	// Verify it's the correct type
	s, ok := srv.(*server)
	if !ok {
		t.Fatal("expected *server type")
	}

	// Check defaults
	if s.o.Name != "scg-boost" {
		t.Errorf("default Name = %q, want %q", s.o.Name, "scg-boost")
	}
	if s.o.Version != "0.1.0" {
		t.Errorf("default Version = %q, want %q", s.o.Version, "0.1.0")
	}
	if s.o.MaxRows != 500 {
		t.Errorf("default MaxRows = %d, want %d", s.o.MaxRows, 500)
	}
	if s.o.DBQueryTimeout != 3*time.Second {
		t.Errorf("default DBQueryTimeout = %v, want %v", s.o.DBQueryTimeout, 3*time.Second)
	}
}

func TestNew_WithOptions(t *testing.T) {
	logger := &mockLogger{}
	authorizer := &mockAuthorizer{scopes: map[string]bool{"test": true}}

	srv, err := New(
		WithName("my-service"),
		WithVersion("1.2.3"),
		WithLogger(logger),
		WithAuthorizer(authorizer),
		WithMaxRows(100),
		WithDBQueryTimeout(5*time.Second),
	)

	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	s := srv.(*server)

	if s.o.Name != "my-service" {
		t.Errorf("Name = %q, want %q", s.o.Name, "my-service")
	}
	if s.o.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", s.o.Version, "1.2.3")
	}
	if s.o.MaxRows != 100 {
		t.Errorf("MaxRows = %d, want %d", s.o.MaxRows, 100)
	}
	if s.o.DBQueryTimeout != 5*time.Second {
		t.Errorf("DBQueryTimeout = %v, want %v", s.o.DBQueryTimeout, 5*time.Second)
	}
	if s.o.Logger != logger {
		t.Error("Logger not set correctly")
	}
	if s.o.Authorizer != authorizer {
		t.Error("Authorizer not set correctly")
	}
}

func TestNew_NilOptionIgnored(t *testing.T) {
	// Passing nil options should not panic
	srv, err := New(nil, WithName("test"), nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	s := srv.(*server)
	if s.o.Name != "test" {
		t.Errorf("Name = %q, want %q", s.o.Name, "test")
	}
}

func TestNew_DefaultAuthorizer_DeniesAll(t *testing.T) {
	srv, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	s := srv.(*server)

	// Default authorizer should deny everything
	if s.o.Authorizer.HasScope(context.Background(), "any.scope") {
		t.Error("default authorizer should deny all scopes")
	}
}

func TestNew_DefaultLogger_NoOp(t *testing.T) {
	srv, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	s := srv.(*server)

	// Default logger should not panic when called
	s.o.Logger.Debug("test", nil)
	s.o.Logger.Error("test", nil)
}

func TestServer_Start_ContextCancellation(t *testing.T) {
	srv, err := New(WithName("test"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop, err := srv.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Give goroutine time to start
	time.Sleep(10 * time.Millisecond)

	// Cancel context to stop server
	cancel()

	// Give goroutine time to stop
	time.Sleep(10 * time.Millisecond)

	// Stop function should work without error
	if err := stop(); err != nil {
		t.Errorf("stop() error = %v", err)
	}
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version constant should not be empty")
	}
}

func TestNopLogger(t *testing.T) {
	logger := &nopLogger{}

	// Should not panic
	logger.Debug("test", map[string]any{"key": "value"})
	logger.Error("test", map[string]any{"key": "value"})
}

func TestDenyAllAuthorizer(t *testing.T) {
	auth := &denyAllAuthorizer{}

	scopes := []string{
		"admin",
		"dbquery.run",
		"config.get",
		"any.scope",
	}

	for _, scope := range scopes {
		if auth.HasScope(context.Background(), scope) {
			t.Errorf("denyAllAuthorizer should deny scope %q", scope)
		}
	}
}

func TestLogStoreAdapter(t *testing.T) {
	store := &mockLogStore{
		ts:     "2024-01-15T10:30:00Z",
		msg:    "test error",
		fields: map[string]any{"level": "error", "trace_id": "abc123"},
	}

	adapter := &logStoreAdapter{store: store}

	entry, err := adapter.LastError(context.Background())
	if err != nil {
		t.Fatalf("LastError() error = %v", err)
	}

	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Message != "test error" {
		t.Errorf("Message = %q, want %q", entry.Message, "test error")
	}
	if entry.Level != "error" {
		t.Errorf("Level = %q, want %q", entry.Level, "error")
	}
}

func TestLogStoreAdapter_EmptyEntry(t *testing.T) {
	store := &mockLogStore{
		ts:     "",
		msg:    "",
		fields: nil,
	}

	adapter := &logStoreAdapter{store: store}

	entry, err := adapter.LastError(context.Background())
	if err != nil {
		t.Fatalf("LastError() error = %v", err)
	}

	if entry != nil {
		t.Error("expected nil entry for empty timestamp/message")
	}
}

func TestLogStoreAdapter_InvalidTimestamp(t *testing.T) {
	store := &mockLogStore{
		ts:     "invalid-timestamp",
		msg:    "test error",
		fields: nil,
	}

	adapter := &logStoreAdapter{store: store}

	_, err := adapter.LastError(context.Background())
	if err == nil {
		t.Error("expected error for invalid timestamp")
	}
}

// mockLogStore implements types.LogStore for testing.
type mockLogStore struct {
	ts     string
	msg    string
	fields map[string]any
	err    error
}

func (m *mockLogStore) LastError(ctx context.Context) (ts, msg string, fields map[string]any, err error) {
	return m.ts, m.msg, m.fields, m.err
}
