package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectRepoType(t *testing.T) {
	// Create temp test directories
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func(string) string
		wantType string
	}{
		{
			name: "go service with cmd",
			setup: func(base string) string {
				dir := filepath.Join(base, "go-service")
				if err := os.MkdirAll(filepath.Join(dir, "cmd"), 0o755); err != nil {
					t.Fatalf("mkdir cmd: %v", err)
				}
				if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644); err != nil {
					t.Fatalf("write go.mod: %v", err)
				}
				return dir
			},
			wantType: "go-service",
		},
		{
			name: "go library without cmd",
			setup: func(base string) string {
				dir := filepath.Join(base, "go-library")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("mkdir library: %v", err)
				}
				if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644); err != nil {
					t.Fatalf("write go.mod: %v", err)
				}
				return dir
			},
			wantType: "go-library",
		},
		{
			name: "terraform repo",
			setup: func(base string) string {
				dir := filepath.Join(base, "terraform")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("mkdir terraform: %v", err)
				}
				if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte("# terraform"), 0o644); err != nil {
					t.Fatalf("write main.tf: %v", err)
				}
				return dir
			},
			wantType: "terraform",
		},
		{
			name: "generic repo",
			setup: func(base string) string {
				dir := filepath.Join(base, "generic")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("mkdir generic: %v", err)
				}
				if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# readme"), 0o644); err != nil {
					t.Fatalf("write README.md: %v", err)
				}
				return dir
			},
			wantType: "generic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(tmpDir)
			got := DetectRepoType(dir)
			if got != tt.wantType {
				t.Errorf("DetectRepoType() = %v, want %v", got, tt.wantType)
			}
		})
	}
}
