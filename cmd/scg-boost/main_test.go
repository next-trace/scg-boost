package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdInstallWritesMCPConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	code := cmdInstall([]string{"--root", root, "--repo", "_generic", "--force", "--name", "demo-server", "--check-mcp-up=false"})
	if code != 0 {
		t.Fatalf("cmdInstall() = %d, want 0", code)
	}

	if _, err := os.Stat(filepath.Join(root, ".claude", "CLAUDE.md")); err != nil {
		t.Fatalf("missing .claude/CLAUDE.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".codex", "CODEX.md")); err != nil {
		t.Fatalf("missing .codex/CODEX.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".gemini", "GEMINI.md")); err != nil {
		t.Fatalf("missing .gemini/GEMINI.md: %v", err)
	}

	body, err := os.ReadFile(filepath.Join(root, ".mcp.json"))
	if err != nil {
		t.Fatalf("missing .mcp.json: %v", err)
	}

	var cfg struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(body, &cfg); err != nil {
		t.Fatalf("invalid .mcp.json: %v", err)
	}

	server, ok := cfg.MCPServers["scg-boost"]
	if !ok {
		t.Fatalf("missing scg-boost server in .mcp.json")
	}
	if server.Command != "scg-boost" {
		t.Fatalf("command = %q, want scg-boost", server.Command)
	}
	if len(server.Args) < 5 {
		t.Fatalf("args too short: %#v", server.Args)
	}

	gitignore, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("missing .gitignore: %v", err)
	}
	content := string(gitignore)
	for _, line := range []string{".claude/", ".codex/", ".gemini/"} {
		if !containsLine(content, line) {
			t.Fatalf(".gitignore missing %q", line)
		}
	}
}

func TestCmdUpdateWritesMCPConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	code := cmdUpdate([]string{"--root", root, "--repo", "_generic", "--name", "updated-server", "--check-mcp-up=false"})
	if code != 0 {
		t.Fatalf("cmdUpdate() = %d, want 0", code)
	}

	if _, err := os.Stat(filepath.Join(root, ".mcp.json")); err != nil {
		t.Fatalf("missing .mcp.json after update: %v", err)
	}
}

func TestCmdInstall_WithPresetBoost(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	code := cmdInstall([]string{"--root", root, "--preset", "boost", "--force", "--check-mcp-up=false"})
	if code != 0 {
		t.Fatalf("cmdInstall() = %d, want 0", code)
	}

	if _, err := os.Stat(filepath.Join(root, ".claude", "commands")); err != nil {
		t.Fatalf("missing .claude/commands: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".codex", "skills")); err != nil {
		t.Fatalf("missing .codex/skills: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".gemini", "skills")); err != nil {
		t.Fatalf("missing .gemini/skills: %v", err)
	}
}

func TestCmdInstall_InvalidPreset(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	code := cmdInstall([]string{"--root", root, "--preset", "unknown", "--check-mcp-up=false"})
	if code != 2 {
		t.Fatalf("cmdInstall() = %d, want 2", code)
	}
}

func TestCmdValidate_SuccessAfterInstall(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if code := cmdInstall([]string{"--root", root, "--repo", "_generic", "--force", "--name", "demo-server", "--check-mcp-up=false"}); code != 0 {
		t.Fatalf("cmdInstall() = %d, want 0", code)
	}
	if code := cmdValidate([]string{"--root", root}); code != 0 {
		t.Fatalf("cmdValidate() = %d, want 0", code)
	}
}

func containsLine(content, line string) bool {
	for _, l := range strings.Split(content, "\n") {
		if l == line {
			return true
		}
	}
	return false
}
