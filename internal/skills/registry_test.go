package skills

import (
	"testing"
	"testing/fstest"
)

func TestRegistry_Load(t *testing.T) {
	fs := fstest.MapFS{
		"gateway-service/skill.json": &fstest.MapFile{
			Data: []byte(`{
				"id": "gateway-service",
				"name": "gateway-service",
				"type": "concrete",
				"version": "1.0.0",
				"description": "Gateway service",
				"tags": ["service"],
				"repo_types": ["go-service"],
				"author": "SupplyChainGuard"
			}`),
		},
		"scg-config/skill.json": &fstest.MapFile{
			Data: []byte(`{
				"id": "scg-config",
				"name": "scg-config",
				"type": "library",
				"version": "1.0.0",
				"description": "Config library",
				"tags": ["library"],
				"repo_types": ["go-library"],
				"author": "SupplyChainGuard"
			}`),
		},
		"no-metadata/.claude/CLAUDE.md": &fstest.MapFile{
			Data: []byte("some content"),
		},
		"_PACK_PLAN.md": &fstest.MapFile{
			Data: []byte("plan"),
		},
	}

	reg, err := Load(fs)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got := reg.Count(); got != 2 {
		t.Errorf("Load() loaded %d skills, want 2", got)
	}

	if !reg.HasSkill("gateway-service") {
		t.Error("Load() missing gateway-service")
	}

	if !reg.HasSkill("scg-config") {
		t.Error("Load() missing scg-config")
	}

	if reg.HasSkill("no-metadata") {
		t.Error("Load() should skip directories without skill.json")
	}
}

func TestRegistry_Get(t *testing.T) {
	reg := NewRegistry()
	reg.skills["test"] = &Metadata{Name: "test"}

	if got := reg.Get("test"); got == nil {
		t.Error("Get(test) returned nil")
	}

	if got := reg.Get("missing"); got != nil {
		t.Error("Get(missing) should return nil")
	}
}

func TestRegistry_List(t *testing.T) {
	reg := NewRegistry()
	reg.skills["zebra"] = &Metadata{Name: "zebra"}
	reg.skills["alpha"] = &Metadata{Name: "alpha"}
	reg.skills["middle"] = &Metadata{Name: "middle"}

	list := reg.List()
	if len(list) != 3 {
		t.Fatalf("List() length = %d, want 3", len(list))
	}

	// Should be sorted
	if list[0].Name != "alpha" || list[1].Name != "middle" || list[2].Name != "zebra" {
		t.Errorf("List() not sorted: %v, %v, %v", list[0].Name, list[1].Name, list[2].Name)
	}
}

func TestRegistry_MatchRepoType(t *testing.T) {
	reg := NewRegistry()
	reg.skills["service1"] = &Metadata{
		Name:      "service1",
		RepoTypes: []string{"go-service"},
	}
	reg.skills["service2"] = &Metadata{
		Name:      "service2",
		RepoTypes: []string{"go-service"},
	}
	reg.skills["library"] = &Metadata{
		Name:      "library",
		RepoTypes: []string{"go-library"},
	}
	reg.skills["generic"] = &Metadata{
		Name:      "generic",
		RepoTypes: []string{"generic"},
	}

	tests := []struct {
		name     string
		repoType string
		wantLen  int
	}{
		{
			name:     "match go-service",
			repoType: "go-service",
			wantLen:  3, // service1, service2, generic
		},
		{
			name:     "match go-library",
			repoType: "go-library",
			wantLen:  2, // library, generic
		},
		{
			name:     "empty returns all",
			repoType: "",
			wantLen:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := reg.MatchRepoType(tt.repoType)
			if len(matches) != tt.wantLen {
				t.Errorf("MatchRepoType(%q) matched %d, want %d", tt.repoType, len(matches), tt.wantLen)
			}
		})
	}
}
