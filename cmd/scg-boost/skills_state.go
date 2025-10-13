package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/next-trace/scg-boost/internal/skills"
)

type InstalledState struct {
	Skills []InstalledSkill `json:"skills"`
}

type InstalledSkill struct {
	ID           string    `json:"id"`
	Version      string    `json:"version"`
	InstalledAt  time.Time `json:"installed_at"`
	HasOverrides bool      `json:"has_overrides"`
}

func loadInstalledState(root string) (*InstalledState, error) {
	data, err := readRepoFile(root, ".scg/installed.json")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &InstalledState{}, nil
		}
		return nil, err
	}

	var state InstalledState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func saveInstalledState(root string, state *InstalledState) error {
	if err := os.MkdirAll(filepath.Join(root, ".scg"), 0o750); err != nil {
		return fmt.Errorf("mkdir .scg: %w", err)
	}
	statePath := filepath.Join(root, ".scg", "installed.json")
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statePath, data, 0o600)
}

func upsertInstalledSkill(state *InstalledState, meta *skills.Metadata, hasOverrides bool) {
	if state == nil || meta == nil {
		return
	}
	for i := range state.Skills {
		if state.Skills[i].ID == meta.ID {
			state.Skills[i].Version = meta.Version
			state.Skills[i].InstalledAt = time.Now().UTC()
			state.Skills[i].HasOverrides = hasOverrides
			return
		}
	}
	state.Skills = append(state.Skills, InstalledSkill{
		ID:           meta.ID,
		Version:      meta.Version,
		InstalledAt:  time.Now().UTC(),
		HasOverrides: hasOverrides,
	})
}

func installedSkillIDs(state *InstalledState) []string {
	if state == nil {
		return nil
	}
	ids := make([]string, 0, len(state.Skills))
	for _, skill := range state.Skills {
		if skill.ID != "" {
			ids = append(ids, skill.ID)
		}
	}
	return ids
}

func hasOverrides(root, skillID string) bool {
	overrideRoot := filepath.Join(root, ".scg", "overrides", skillID)
	_, err := os.Stat(overrideRoot)
	if err != nil {
		return false
	}
	found := false
	_ = filepath.WalkDir(overrideRoot, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		found = true
		return filepath.SkipDir
	})
	return found
}
