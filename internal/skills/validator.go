package skills

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
)

var semverRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// ValidateSkill validates a single skill against the JSON schema and business rules.
// skillPath is relative to the FS root (e.g., "gateway-service").
func ValidateSkill(fsys fs.FS, skillPath string) error {
	// Load metadata
	meta, err := LoadMetadata(fsys, skillPath)
	if err != nil {
		return fmt.Errorf("load metadata: %w", err)
	}

	// Validate against JSON schema
	if err := validateAgainstSchema(meta); err != nil {
		return fmt.Errorf("schema validation: %w", err)
	}

	// Business rule: ID must match directory name
	expectedID := filepath.Base(skillPath)
	if meta.ID != expectedID {
		return fmt.Errorf("skill ID %q does not match directory name %q", meta.ID, expectedID)
	}

	// Business rule: version must be valid semver
	if !semverRegex.MatchString(meta.Version) {
		return fmt.Errorf("invalid semver version: %s", meta.Version)
	}

	// Business rule: check required files exist
	requiredFiles := []string{
		".claude/CLAUDE.md",
	}
	for _, file := range requiredFiles {
		fullPath := filepath.Join(skillPath, file)
		if _, err := fs.Stat(fsys, fullPath); err != nil {
			return fmt.Errorf("required file %s not found", file)
		}
	}

	return nil
}

// ValidateRegistry validates the entire registry for:
// - Duplicate skill IDs
// - Circular dependencies
// - Invalid dependency references
func ValidateRegistry(reg *Registry) error {
	if reg == nil {
		return fmt.Errorf("registry is nil")
	}

	// Check for duplicate IDs (already handled by registry loading)
	idSet := make(map[string]bool)
	for _, skill := range reg.skills {
		if idSet[skill.ID] {
			return fmt.Errorf("duplicate skill ID: %s", skill.ID)
		}
		idSet[skill.ID] = true
	}

	// Validate all dependency references exist
	for _, skill := range reg.skills {
		for _, depID := range skill.DependsOn {
			if _, exists := reg.skills[depID]; !exists {
				return fmt.Errorf("skill %s depends on non-existent skill %s", skill.ID, depID)
			}
		}
		for _, conflictID := range skill.ConflictsWith {
			if _, exists := reg.skills[conflictID]; !exists {
				return fmt.Errorf("skill %s conflicts with non-existent skill %s", skill.ID, conflictID)
			}
		}
	}

	// Check for circular dependencies
	for skillID := range reg.skills {
		if err := detectCycles(reg, skillID, make(map[string]bool), make(map[string]bool)); err != nil {
			return err
		}
	}

	return nil
}

// validateAgainstSchema validates metadata against the JSON schema.
// For now, we use manual validation. In the future, we could integrate
// a proper JSON schema validator if needed.
func validateAgainstSchema(meta *Metadata) error {
	return validateMetadataManually(meta)
}

// validateMetadataManually performs manual validation of metadata fields.
func validateMetadataManually(meta *Metadata) error {
	// ID pattern: lowercase alphanumeric and hyphens
	idPattern := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !idPattern.MatchString(meta.ID) {
		return fmt.Errorf("id must match pattern ^[a-z0-9-]+$")
	}
	if len(meta.ID) < 1 || len(meta.ID) > 64 {
		return fmt.Errorf("id length must be between 1 and 64 characters")
	}

	// Validate repo_types enum
	validRepoTypes := map[string]bool{
		"go-service": true,
		"go-library": true,
		"terraform":  true,
		"generic":    true,
	}
	for _, rt := range meta.RepoTypes {
		if !validRepoTypes[rt] {
			return fmt.Errorf("invalid repo_type: %s", rt)
		}
	}

	// Validate type enum
	validTypes := map[string]bool{
		"concrete": true,
		"library":  true,
		"generic":  true,
	}
	if meta.Type == "" {
		return fmt.Errorf("type is required")
	}
	if !validTypes[meta.Type] {
		return fmt.Errorf("invalid type: %s", meta.Type)
	}

	// Validate tags pattern
	tagPattern := regexp.MustCompile(`^[a-z0-9-]+$`)
	for _, tag := range meta.Tags {
		if !tagPattern.MatchString(tag) {
			return fmt.Errorf("tag %q must match pattern ^[a-z0-9-]+$", tag)
		}
	}

	// Validate dependency IDs
	depPattern := regexp.MustCompile(`^[a-z0-9-]+$`)
	for _, dep := range meta.DependsOn {
		if !depPattern.MatchString(dep) {
			return fmt.Errorf("dependency ID %q must match pattern ^[a-z0-9-]+$", dep)
		}
	}
	for _, conflict := range meta.ConflictsWith {
		if !depPattern.MatchString(conflict) {
			return fmt.Errorf("conflict ID %q must match pattern ^[a-z0-9-]+$", conflict)
		}
	}

	return nil
}

// detectCycles detects circular dependencies using DFS.
// visited tracks nodes in current path, finished tracks completed nodes.
func detectCycles(reg *Registry, skillID string, visited, finished map[string]bool) error {
	if finished[skillID] {
		return nil
	}
	if visited[skillID] {
		return fmt.Errorf("circular dependency detected involving skill: %s", skillID)
	}

	visited[skillID] = true
	skill, exists := reg.skills[skillID]
	if !exists {
		return fmt.Errorf("skill not found: %s", skillID)
	}

	for _, depID := range skill.DependsOn {
		if err := detectCycles(reg, depID, visited, finished); err != nil {
			// Add context to error
			if strings.Contains(err.Error(), "circular dependency") {
				return fmt.Errorf("%w -> %s", err, skillID)
			}
			return err
		}
	}

	finished[skillID] = true
	delete(visited, skillID)
	return nil
}
