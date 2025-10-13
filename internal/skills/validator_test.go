package skills

import (
	"testing"
	"testing/fstest"
)

func TestValidateSkill(t *testing.T) {
	tests := []struct {
		name      string
		fsys      fstest.MapFS
		skillPath string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid skill",
			fsys: fstest.MapFS{
				"test-skill/skill.json": &fstest.MapFile{
					Data: []byte(`{
						"id": "test-skill",
						"name": "Test Skill",
						"type": "concrete",
						"version": "1.0.0",
						"description": "A test skill",
						"repo_types": ["go-service"]
					}`),
				},
				"test-skill/.claude/CLAUDE.md": &fstest.MapFile{
					Data: []byte("# Test"),
				},
			},
			skillPath: "test-skill",
			wantErr:   false,
		},
		{
			name: "missing id field",
			fsys: fstest.MapFS{
				"test-skill/skill.json": &fstest.MapFile{
					Data: []byte(`{
						"name": "Test Skill",
						"type": "concrete",
						"version": "1.0.0",
						"description": "A test skill",
						"repo_types": ["go-service"]
					}`),
				},
				"test-skill/.claude/CLAUDE.md": &fstest.MapFile{
					Data: []byte("# Test"),
				},
			},
			skillPath: "test-skill",
			wantErr:   true,
			errMsg:    "id is required",
		},
		{
			name: "missing type field",
			fsys: fstest.MapFS{
				"test-skill/skill.json": &fstest.MapFile{
					Data: []byte(`{
						"id": "test-skill",
						"name": "Test Skill",
						"version": "1.0.0",
						"description": "A test skill",
						"repo_types": ["go-service"]
					}`),
				},
				"test-skill/.claude/CLAUDE.md": &fstest.MapFile{
					Data: []byte("# Test"),
				},
			},
			skillPath: "test-skill",
			wantErr:   true,
			errMsg:    "type is required",
		},
		{
			name: "id mismatch with directory",
			fsys: fstest.MapFS{
				"test-skill/skill.json": &fstest.MapFile{
					Data: []byte(`{
						"id": "wrong-id",
						"name": "Test Skill",
						"type": "concrete",
						"version": "1.0.0",
						"description": "A test skill",
						"repo_types": ["go-service"]
					}`),
				},
				"test-skill/.claude/CLAUDE.md": &fstest.MapFile{
					Data: []byte("# Test"),
				},
			},
			skillPath: "test-skill",
			wantErr:   true,
			errMsg:    "does not match directory name",
		},
		{
			name: "invalid semver",
			fsys: fstest.MapFS{
				"test-skill/skill.json": &fstest.MapFile{
					Data: []byte(`{
						"id": "test-skill",
						"name": "Test Skill",
						"type": "concrete",
						"version": "1.0",
						"description": "A test skill",
						"repo_types": ["go-service"]
					}`),
				},
				"test-skill/.claude/CLAUDE.md": &fstest.MapFile{
					Data: []byte("# Test"),
				},
			},
			skillPath: "test-skill",
			wantErr:   true,
			errMsg:    "invalid semver",
		},
		{
			name: "missing required file",
			fsys: fstest.MapFS{
				"test-skill/skill.json": &fstest.MapFile{
					Data: []byte(`{
						"id": "test-skill",
						"name": "Test Skill",
						"type": "concrete",
						"version": "1.0.0",
						"description": "A test skill",
						"repo_types": ["go-service"]
					}`),
				},
			},
			skillPath: "test-skill",
			wantErr:   true,
			errMsg:    "required file",
		},
		{
			name: "invalid id pattern",
			fsys: fstest.MapFS{
				"test-skill/skill.json": &fstest.MapFile{
					Data: []byte(`{
						"id": "Test_Skill",
						"name": "Test Skill",
						"type": "concrete",
						"version": "1.0.0",
						"description": "A test skill",
						"repo_types": ["go-service"]
					}`),
				},
				"test-skill/.claude/CLAUDE.md": &fstest.MapFile{
					Data: []byte("# Test"),
				},
			},
			skillPath: "test-skill",
			wantErr:   true,
			errMsg:    "must match pattern",
		},
		{
			name: "invalid repo_type",
			fsys: fstest.MapFS{
				"test-skill/skill.json": &fstest.MapFile{
					Data: []byte(`{
						"id": "test-skill",
						"name": "Test Skill",
						"type": "concrete",
						"version": "1.0.0",
						"description": "A test skill",
						"repo_types": ["invalid-type"]
					}`),
				},
				"test-skill/.claude/CLAUDE.md": &fstest.MapFile{
					Data: []byte("# Test"),
				},
			},
			skillPath: "test-skill",
			wantErr:   true,
			errMsg:    "invalid repo_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSkill(tt.fsys, tt.skillPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSkill() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateSkill() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateRegistry(t *testing.T) {
	tests := []struct {
		name    string
		reg     *Registry
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid registry",
			reg: &Registry{
				skills: map[string]*Metadata{
					"skill-a": {
						ID:        "skill-a",
						Name:      "Skill A",
						Version:   "1.0.0",
						DependsOn: []string{"skill-b"},
					},
					"skill-b": {
						ID:      "skill-b",
						Name:    "Skill B",
						Version: "1.0.0",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "non-existent dependency",
			reg: &Registry{
				skills: map[string]*Metadata{
					"skill-a": {
						ID:        "skill-a",
						Name:      "Skill A",
						Version:   "1.0.0",
						DependsOn: []string{"skill-missing"},
					},
				},
			},
			wantErr: true,
			errMsg:  "non-existent skill",
		},
		{
			name: "circular dependency",
			reg: &Registry{
				skills: map[string]*Metadata{
					"skill-a": {
						ID:        "skill-a",
						Name:      "Skill A",
						Version:   "1.0.0",
						DependsOn: []string{"skill-b"},
					},
					"skill-b": {
						ID:        "skill-b",
						Name:      "Skill B",
						Version:   "1.0.0",
						DependsOn: []string{"skill-a"},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
		{
			name: "non-existent conflict",
			reg: &Registry{
				skills: map[string]*Metadata{
					"skill-a": {
						ID:            "skill-a",
						Name:          "Skill A",
						Version:       "1.0.0",
						ConflictsWith: []string{"skill-missing"},
					},
				},
			},
			wantErr: true,
			errMsg:  "non-existent skill",
		},
		{
			name: "three-way circular dependency",
			reg: &Registry{
				skills: map[string]*Metadata{
					"skill-a": {
						ID:        "skill-a",
						Name:      "Skill A",
						Version:   "1.0.0",
						DependsOn: []string{"skill-b"},
					},
					"skill-b": {
						ID:        "skill-b",
						Name:      "Skill B",
						Version:   "1.0.0",
						DependsOn: []string{"skill-c"},
					},
					"skill-c": {
						ID:        "skill-c",
						Name:      "Skill C",
						Version:   "1.0.0",
						DependsOn: []string{"skill-a"},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistry(tt.reg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRegistry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateRegistry() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
