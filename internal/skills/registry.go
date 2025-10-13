package skills

import (
	"fmt"
	"io/fs"
	"sort"
)

// Registry holds all available skills indexed by name.
type Registry struct {
	skills map[string]*Metadata
}

// NewRegistry creates an empty skill registry.
func NewRegistry() *Registry {
	return &Registry{
		skills: make(map[string]*Metadata),
	}
}

// Load scans the given filesystem for skill.json files and populates the registry.
// Expects directory structure: <fsys>/<skill-name>/skill.json
// Silently skips directories without skill.json (backward compatibility).
func Load(fsys fs.FS) (*Registry, error) {
	reg := NewRegistry()

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read root: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillName := entry.Name()

		// Skip internal/metadata files
		if skillName == "_PACK_PLAN.md" || skillName == "_PACK_README.md" {
			continue
		}

		// Try to load metadata; skip if not present (backward compat)
		meta, err := LoadMetadata(fsys, skillName)
		if err != nil {
			// Silently skip directories without skill.json
			continue
		}

		reg.skills[meta.Name] = meta
	}

	return reg, nil
}

// Get retrieves a skill by name. Returns nil if not found.
func (r *Registry) Get(name string) *Metadata {
	return r.skills[name]
}

// List returns all skills sorted by name.
func (r *Registry) List() []*Metadata {
	names := make([]string, 0, len(r.skills))
	for name := range r.skills {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]*Metadata, 0, len(names))
	for _, name := range names {
		result = append(result, r.skills[name])
	}
	return result
}

// MatchRepoType returns skills applicable to the given repo type.
// Returns all skills if repoType is empty.
func (r *Registry) MatchRepoType(repoType string) []*Metadata {
	if repoType == "" {
		return r.List()
	}

	var matches []*Metadata
	for _, skill := range r.skills {
		for _, rt := range skill.RepoTypes {
			if rt == repoType || rt == "generic" {
				matches = append(matches, skill)
				break
			}
		}
	}

	// Sort by name
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Name < matches[j].Name
	})

	return matches
}

// HasSkill returns true if the registry contains the named skill.
func (r *Registry) HasSkill(name string) bool {
	_, exists := r.skills[name]
	return exists
}

// Count returns the number of registered skills.
func (r *Registry) Count() int {
	return len(r.skills)
}
