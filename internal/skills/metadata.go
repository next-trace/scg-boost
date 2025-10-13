package skills

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
)

// Metadata describes a skill's properties and applicability.
type Metadata struct {
	ID            string   `json:"id"`                       // Unique skill identifier (must match directory name)
	Name          string   `json:"name"`                     // Human-readable skill name
	Type          string   `json:"type"`                     // Skill category (concrete, library, generic)
	Version       string   `json:"version"`                  // Semantic version (e.g., 1.0.0)
	Description   string   `json:"description"`              // Brief description of the skill
	Tags          []string `json:"tags"`                     // Tags for categorization and search
	RepoTypes     []string `json:"repo_types"`               // Repository types this skill applies to
	Author        string   `json:"author"`                   // Author or organization name
	DependsOn     []string `json:"depends_on,omitempty"`     // Skill IDs this depends on
	ConflictsWith []string `json:"conflicts_with,omitempty"` // Skill IDs that conflict with this
	Provides      []string `json:"provides,omitempty"`       // Capabilities this skill provides
	OverridePaths []string `json:"override_paths,omitempty"` // Paths safe to override
}

// LoadMetadata reads skill.json from the given skill path within the FS.
// Returns error if file not found or JSON is malformed.
func LoadMetadata(fsys fs.FS, skillPath string) (*Metadata, error) {
	metaPath := filepath.Join(skillPath, "skill.json")
	data, err := fs.ReadFile(fsys, metaPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", metaPath, err)
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", metaPath, err)
	}

	// Validate required fields
	if meta.ID == "" {
		return nil, fmt.Errorf("%s: id is required", metaPath)
	}
	if meta.Name == "" {
		return nil, fmt.Errorf("%s: name is required", metaPath)
	}
	if meta.Type == "" {
		return nil, fmt.Errorf("%s: type is required", metaPath)
	}
	if meta.Version == "" {
		return nil, fmt.Errorf("%s: version is required", metaPath)
	}
	if meta.Description == "" {
		return nil, fmt.Errorf("%s: description is required", metaPath)
	}
	if len(meta.RepoTypes) == 0 {
		return nil, fmt.Errorf("%s: repo_types is required", metaPath)
	}

	return &meta, nil
}
