package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestApplyOverrides_FullFile(t *testing.T) {
	root := t.TempDir()

	basePath := filepath.Join(root, ".claude", "commands", "custom.md")
	if err := os.MkdirAll(filepath.Dir(basePath), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(basePath, []byte("base"), 0o600); err != nil {
		t.Fatalf("write base: %v", err)
	}

	overridePath := filepath.Join(root, ".scg", "overrides", "test-skill", "commands", "custom.md")
	if err := os.MkdirAll(filepath.Dir(overridePath), 0o750); err != nil {
		t.Fatalf("mkdir override: %v", err)
	}
	if err := os.WriteFile(overridePath, []byte("override"), 0o600); err != nil {
		t.Fatalf("write override: %v", err)
	}

	warnings, err := ApplyOverrides(root, "test-skill", []string{".claude/commands/custom.md"})
	if err != nil {
		t.Fatalf("ApplyOverrides error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	got, err := os.ReadFile(basePath)
	if err != nil {
		t.Fatalf("read base: %v", err)
	}
	if string(got) != "override" {
		t.Fatalf("override not applied, got %q", string(got))
	}
}

func TestApplyOverrides_Section(t *testing.T) {
	root := t.TempDir()

	base := strings.Join([]string{
		"# Title",
		"",
		"## Repo-Specific Rules",
		"base rules",
		"",
		"## Other",
		"keep this",
		"",
	}, "\n")
	basePath := filepath.Join(root, ".claude", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(basePath), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(basePath, []byte(base), 0o600); err != nil {
		t.Fatalf("write base: %v", err)
	}

	overridePath := filepath.Join(root, ".scg", "overrides", "test-skill", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(overridePath), 0o750); err != nil {
		t.Fatalf("mkdir override: %v", err)
	}
	if err := os.WriteFile(overridePath, []byte("override rules"), 0o600); err != nil {
		t.Fatalf("write override: %v", err)
	}

	_, err := ApplyOverrides(root, "test-skill", []string{".claude/CLAUDE.md#repo-specific-rules"})
	if err != nil {
		t.Fatalf("ApplyOverrides error: %v", err)
	}

	got, err := os.ReadFile(basePath)
	if err != nil {
		t.Fatalf("read base: %v", err)
	}

	if !strings.Contains(string(got), "override rules") {
		t.Fatalf("section override not applied")
	}
	if !strings.Contains(string(got), "keep this") {
		t.Fatalf("unexpected change to other sections")
	}
}

func TestApplyOverrides_DisallowedPath(t *testing.T) {
	root := t.TempDir()
	overridePath := filepath.Join(root, ".scg", "overrides", "test-skill", "commands", "custom.md")
	if err := os.MkdirAll(filepath.Dir(overridePath), 0o750); err != nil {
		t.Fatalf("mkdir override: %v", err)
	}
	if err := os.WriteFile(overridePath, []byte("override"), 0o600); err != nil {
		t.Fatalf("write override: %v", err)
	}

	warnings, err := ApplyOverrides(root, "test-skill", []string{})
	if err != nil {
		t.Fatalf("ApplyOverrides error: %v", err)
	}
	if len(warnings) == 0 {
		t.Fatalf("expected warning for disallowed path")
	}
}

func TestDetectConflicts(t *testing.T) {
	templates := fstest.MapFS{
		"skill-a/.claude/CLAUDE.md":        &fstest.MapFile{Data: []byte("a")},
		"skill-b/.claude/CLAUDE.md":        &fstest.MapFile{Data: []byte("b")},
		"skill-b/.claude/commands/test.md": &fstest.MapFile{Data: []byte("b")},
	}

	conflicts, err := DetectConflicts(templates, []string{"skill-a"}, "skill-b")
	if err != nil {
		t.Fatalf("DetectConflicts error: %v", err)
	}
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].Path != ".claude/CLAUDE.md" {
		t.Fatalf("unexpected conflict path: %s", conflicts[0].Path)
	}
}
