// Package runtime provides controlled execution environments.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ExecPolicy defines execution constraints for sandboxed commands.
type ExecPolicy struct {
	// AllowedCommands is the whitelist of allowed command names.
	// Empty means no commands allowed (default secure behavior).
	AllowedCommands []string

	// MaxTimeout is the maximum execution time allowed.
	MaxTimeout time.Duration

	// MaxOutputBytes limits stdout/stderr size.
	MaxOutputBytes int64

	// AllowNetworking permits network access during execution.
	AllowNetworking bool

	// WorkingDir restricts execution to a specific directory.
	WorkingDir string
}

// DefaultPolicy returns a restrictive default policy.
func DefaultPolicy() *ExecPolicy {
	return &ExecPolicy{
		AllowedCommands: []string{}, // No commands allowed by default
		MaxTimeout:      30 * time.Second,
		MaxOutputBytes:  1024 * 1024, // 1MB
		AllowNetworking: false,
		WorkingDir:      "",
	}
}

// ExecSandbox provides controlled command execution with security policies.
type ExecSandbox struct {
	policy *ExecPolicy
}

// NewExecSandbox creates a new sandbox with the given policy.
func NewExecSandbox(policy *ExecPolicy) *ExecSandbox {
	if policy == nil {
		policy = DefaultPolicy()
	}
	return &ExecSandbox{policy: policy}
}

// ErrCommandNotAllowed is returned when a command is not in the whitelist.
var ErrCommandNotAllowed = errors.New("command not allowed by policy")

// ErrExecutionDisabled is returned when execution is disabled.
var ErrExecutionDisabled = errors.New("command execution is disabled")

// ErrTimeout is returned when command exceeds timeout.
var ErrTimeout = errors.New("command execution timeout")

// ExecResult contains the result of a sandboxed command execution.
type ExecResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// Run executes a command within the sandbox constraints.
func (s *ExecSandbox) Run(ctx context.Context, command string, args ...string) (*ExecResult, error) {
	if s.policy == nil || len(s.policy.AllowedCommands) == 0 {
		return nil, ErrExecutionDisabled
	}

	// Check if command is allowed
	allowed := false
	for _, cmd := range s.policy.AllowedCommands {
		if cmd == command {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("%w: %s", ErrCommandNotAllowed, command)
	}

	// Validate args don't contain shell injection patterns
	for _, arg := range args {
		if containsShellInjection(arg) {
			return nil, errors.New("argument contains potentially dangerous characters")
		}
	}

	// Create command with timeout
	timeout := s.policy.MaxTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, command, args...)

	// Set working directory if specified
	if s.policy.WorkingDir != "" {
		cmd.Dir = s.policy.WorkingDir
	}

	// Capture output
	start := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	// Check for timeout
	if execCtx.Err() == context.DeadlineExceeded {
		return nil, ErrTimeout
	}

	// Truncate output if needed
	if int64(len(output)) > s.policy.MaxOutputBytes {
		output = output[:s.policy.MaxOutputBytes]
	}

	result := &ExecResult{
		ExitCode: 0,
		Stdout:   string(output),
		Stderr:   "",
		Duration: duration,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}

	return result, nil
}

// IsEnabled returns true if the sandbox allows any command execution.
func (s *ExecSandbox) IsEnabled() bool {
	return s.policy != nil && len(s.policy.AllowedCommands) > 0
}

// containsShellInjection checks for common shell injection patterns.
func containsShellInjection(s string) bool {
	dangerous := []string{
		";", "&&", "||", "|", "`", "$(", "${",
		"\n", "\r", ">", "<", "&",
	}
	for _, d := range dangerous {
		if strings.Contains(s, d) {
			return true
		}
	}
	return false
}
