package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/next-trace/scg-boost/internal/bootstrap"
	"github.com/next-trace/scg-boost/internal/skills"
	"github.com/next-trace/scg-boost/resources"
)

func cmdSkillsOverride(args []string) int {
	fs := flag.NewFlagSet("skills:override", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	skillName := fs.String("skill", "", "skill name (required)")
	root := fs.String("root", ".", "repo root")
	overridePath := fs.String("path", "", "override path (from skill override_paths)")
	force := fs.Bool("force", false, "overwrite existing override file")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *skillName == "" {
		fmt.Fprintln(os.Stderr, "error: --skill is required")
		return 2
	}

	abs, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	tpl, err := resources.BootstrapTemplates()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	reg, err := skills.Load(tpl)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	meta := reg.Get(*skillName)
	if meta == nil {
		fmt.Fprintf(os.Stderr, "error: skill %q not found\n", *skillName)
		return 1
	}

	if len(meta.OverridePaths) == 0 {
		fmt.Fprintf(os.Stderr, "Skill %q does not declare override paths\n", meta.ID)
		return 0
	}

	if *overridePath == "" {
		fmt.Printf("Overrideable paths for %q:\n", meta.ID)
		for _, p := range meta.OverridePaths {
			fmt.Printf("  - %s\n", p)
		}
		return 0
	}

	if !containsString(meta.OverridePaths, *overridePath) {
		fmt.Fprintf(os.Stderr, "error: path %q is not in override_paths for %s\n", *overridePath, meta.ID)
		return 1
	}

	filePath, section := splitOverridePath(*overridePath)
	if !strings.HasPrefix(filePath, ".claude/") {
		fmt.Fprintf(os.Stderr, "error: override path must start with .claude/: %s\n", filePath)
		return 1
	}

	baseContent, err := readBaseFile(abs, tpl, meta.ID, filePath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	overrideContent := baseContent
	if section != "" {
		if len(baseContent) == 0 {
			fmt.Fprintln(os.Stderr, "error: base file not found for section override")
			return 1
		}
		body, err := bootstrap.ExtractSectionBody(baseContent, section)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		overrideContent = body
	}

	rel := strings.TrimPrefix(filePath, ".claude/")
	overrideFile := filepath.Join(abs, ".scg", "overrides", meta.ID, filepath.FromSlash(rel))
	if _, err := os.Stat(overrideFile); err == nil && !*force {
		fmt.Fprintf(os.Stderr, "error: override file exists (use --force): %s\n", overrideFile)
		return 1
	}
	if err := os.MkdirAll(filepath.Dir(overrideFile), 0o750); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if err := os.WriteFile(overrideFile, overrideContent, 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	state, err := loadInstalledState(abs)
	if err == nil {
		upsertInstalledSkill(state, meta, hasOverrides(abs, meta.ID))
		if err := saveInstalledState(abs, state); err != nil {
			fmt.Fprintln(os.Stderr, "warning: failed to write installed skills:", err)
		}
	}

	fmt.Printf("Created override at %s\n", overrideFile)
	return 0
}

func splitOverridePath(path string) (string, string) {
	parts := strings.SplitN(path, "#", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func readBaseFile(root string, templates fs.FS, skillID, filePath string) ([]byte, error) {
	if data, err := readRepoFile(root, filePath); err == nil {
		return data, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	templatePath := path.Join(skillID, filepath.ToSlash(filePath))
	return fs.ReadFile(templates, templatePath)
}

func containsString(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
