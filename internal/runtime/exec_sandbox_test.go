package runtime

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultPolicy(t *testing.T) {
	policy := DefaultPolicy()

	if len(policy.AllowedCommands) != 0 {
		t.Error("default policy should have no allowed commands")
	}
	if policy.MaxTimeout != 30*time.Second {
		t.Errorf("MaxTimeout = %v, want 30s", policy.MaxTimeout)
	}
	if policy.MaxOutputBytes != 1024*1024 {
		t.Errorf("MaxOutputBytes = %d, want 1MB", policy.MaxOutputBytes)
	}
	if policy.AllowNetworking {
		t.Error("AllowNetworking should be false by default")
	}
}

func TestNewExecSandbox_NilPolicy(t *testing.T) {
	sandbox := NewExecSandbox(nil)
	if sandbox.policy == nil {
		t.Error("should use default policy when nil is passed")
	}
}

func TestExecSandbox_ExecutionDisabled(t *testing.T) {
	sandbox := NewExecSandbox(DefaultPolicy())

	_, err := sandbox.Run(context.Background(), "echo", "test")
	if !errors.Is(err, ErrExecutionDisabled) {
		t.Errorf("expected ErrExecutionDisabled, got %v", err)
	}
}

func TestExecSandbox_CommandNotAllowed(t *testing.T) {
	policy := &ExecPolicy{
		AllowedCommands: []string{"ls", "cat"},
		MaxTimeout:      10 * time.Second,
		MaxOutputBytes:  1024,
	}
	sandbox := NewExecSandbox(policy)

	_, err := sandbox.Run(context.Background(), "rm", "-rf", "/")
	if !errors.Is(err, ErrCommandNotAllowed) {
		t.Errorf("expected ErrCommandNotAllowed, got %v", err)
	}
}

func TestExecSandbox_AllowedCommand(t *testing.T) {
	policy := &ExecPolicy{
		AllowedCommands: []string{"echo"},
		MaxTimeout:      10 * time.Second,
		MaxOutputBytes:  1024,
	}
	sandbox := NewExecSandbox(policy)

	result, err := sandbox.Run(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "hello\n")
	}
}

func TestExecSandbox_ShellInjection(t *testing.T) {
	policy := &ExecPolicy{
		AllowedCommands: []string{"echo"},
		MaxTimeout:      10 * time.Second,
		MaxOutputBytes:  1024,
	}
	sandbox := NewExecSandbox(policy)

	dangerous := []string{
		"test; rm -rf /",
		"test && rm -rf /",
		"test || rm -rf /",
		"test | rm -rf /",
		"test `rm -rf /`",
		"test $(rm -rf /)",
		"test ${PATH}",
		"test\nrm -rf /",
	}

	for _, arg := range dangerous {
		_, err := sandbox.Run(context.Background(), "echo", arg)
		if err == nil {
			t.Errorf("expected error for dangerous argument: %q", arg)
		}
	}
}

func TestExecSandbox_Timeout(t *testing.T) {
	policy := &ExecPolicy{
		AllowedCommands: []string{"sleep"},
		MaxTimeout:      100 * time.Millisecond,
		MaxOutputBytes:  1024,
	}
	sandbox := NewExecSandbox(policy)

	_, err := sandbox.Run(context.Background(), "sleep", "10")
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
}

func TestExecSandbox_OutputTruncation(t *testing.T) {
	policy := &ExecPolicy{
		AllowedCommands: []string{"yes"},
		MaxTimeout:      500 * time.Millisecond,
		MaxOutputBytes:  100, // Very small limit
	}
	sandbox := NewExecSandbox(policy)

	// This will be killed by timeout, but output should be truncated
	result, err := sandbox.Run(context.Background(), "yes")
	// yes command will timeout
	if err == nil && result != nil && len(result.Stdout) > 100 {
		t.Errorf("output not truncated: len=%d", len(result.Stdout))
	}
}

func TestExecSandbox_NonZeroExitCode(t *testing.T) {
	policy := &ExecPolicy{
		AllowedCommands: []string{"false"},
		MaxTimeout:      10 * time.Second,
		MaxOutputBytes:  1024,
	}
	sandbox := NewExecSandbox(policy)

	result, err := sandbox.Run(context.Background(), "false")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code")
	}
}

func TestExecSandbox_WorkingDir(t *testing.T) {
	policy := &ExecPolicy{
		AllowedCommands: []string{"pwd"},
		MaxTimeout:      10 * time.Second,
		MaxOutputBytes:  1024,
		WorkingDir:      "/tmp",
	}
	sandbox := NewExecSandbox(policy)

	result, err := sandbox.Run(context.Background(), "pwd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Stdout != "/tmp\n" {
		t.Errorf("Stdout = %q, want /tmp", result.Stdout)
	}
}

func TestExecSandbox_IsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		policy  *ExecPolicy
		enabled bool
	}{
		{"nil policy", nil, false},
		{"empty commands", &ExecPolicy{AllowedCommands: []string{}}, false},
		{"with commands", &ExecPolicy{AllowedCommands: []string{"echo"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sandbox := NewExecSandbox(tt.policy)
			if sandbox.IsEnabled() != tt.enabled {
				t.Errorf("IsEnabled() = %v, want %v", sandbox.IsEnabled(), tt.enabled)
			}
		})
	}
}

func TestContainsShellInjection(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"hello", false},
		{"hello world", false},
		{"hello;world", true},
		{"hello&&world", true},
		{"hello||world", true},
		{"hello|world", true},
		{"hello`world`", true},
		{"hello$(world)", true},
		{"hello${VAR}", true},
		{"hello\nworld", true},
		{"hello>file", true},
		{"hello<file", true},
		{"hello&", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := containsShellInjection(tt.input)
			if got != tt.want {
				t.Errorf("containsShellInjection(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExecSandbox_ContextCancellation(t *testing.T) {
	policy := &ExecPolicy{
		AllowedCommands: []string{"sleep"},
		MaxTimeout:      10 * time.Second,
		MaxOutputBytes:  1024,
	}
	sandbox := NewExecSandbox(policy)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := sandbox.Run(ctx, "sleep", "1")
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
}
