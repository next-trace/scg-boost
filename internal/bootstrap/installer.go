package bootstrap

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var defaultAssistantFiles = map[string]string{
	".claude/CLAUDE.md": "# SCG Context\n\nLocal bootstrap context for this repository.\n",
	".codex/CODEX.md":   "# Codex Context\n\nLocal Codex context for this repository.\n",
	".gemini/GEMINI.md": "# Gemini Context\n\nLocal Gemini context for this repository.\n",
}

var defaultAssistantDirs = []string{
	".claude/agents",
	".claude/commands",
	".codex/agents",
	".codex/commands",
	".codex/skills",
	".gemini/agents",
	".gemini/commands",
	".gemini/skills",
}

type InstallOptions struct {
	RepoName  string
	TargetDir string
	Force     bool
}

// Install copies the generic embedded assistant template set into the target repository.
//
// Rules:
// - Writes only under <TargetDir>/.claude
// - Refuses to overwrite without Force
// - Uses repo-agnostic templates from _generic
func Install(templates fs.FS, opt InstallOptions) error {
	if opt.TargetDir == "" {
		return errors.New("target dir is required")
	}
	absTarget, err := filepath.Abs(opt.TargetDir)
	if err != nil {
		return fmt.Errorf("abs target dir: %w", err)
	}
	opt.TargetDir = absTarget
	if _, err := fs.Stat(templates, "_generic"); err != nil {
		return fmt.Errorf("missing embedded templates: %w", err)
	}

	// We pass the .claude path because the installer expects it as a starting point,
	// but it will also look for sibling .gemini and .codex.
	srcRoot := filepath.ToSlash(filepath.Join("_generic", ".claude"))
	return installFromPath(templates, srcRoot, opt)
}

// InstallSkill installs a specific named skill into the target repository.
// Skill name should match a directory in the templates FS (e.g., "gateway-service").
func InstallSkill(templates fs.FS, skillName string, opt InstallOptions) error {
	if skillName == "" {
		return errors.New("skill name is required")
	}
	if opt.TargetDir == "" {
		return errors.New("target dir is required")
	}

	absTarget, err := filepath.Abs(opt.TargetDir)
	if err != nil {
		return fmt.Errorf("abs target dir: %w", err)
	}
	opt.TargetDir = absTarget

	// Verify skill exists
	srcRoot := filepath.ToSlash(filepath.Join(skillName, ".claude"))
	if _, err := fs.Stat(templates, srcRoot); err != nil {
		return fmt.Errorf("skill %q not found: %w", skillName, err)
	}

	return installFromPath(templates, srcRoot, opt)
}

// installFromPath performs the actual installation from srcRoot to opt.TargetDir.
// It applies repo-specific templates first, then fills any missing files from
// _generic templates (without overwriting existing files unless Force=true).
func installFromPath(templates fs.FS, srcRoot string, opt InstallOptions) error {
	root, err := os.OpenRoot(opt.TargetDir)
	if err != nil {
		return fmt.Errorf("open target root: %w", err)
	}
	defer func() {
		_ = root.Close()
	}()

	// Install .claude, .gemini, and .codex if they exist in templates
	clientDirs := []string{".claude", ".gemini", ".codex"}
	baseSrcRoot := filepath.Dir(srcRoot)
	sourceRoots := []string{baseSrcRoot}
	if baseSrcRoot != "_generic" {
		if _, err := fs.Stat(templates, "_generic"); err == nil {
			sourceRoots = append(sourceRoots, "_generic")
		}
	}

	for _, srcRootName := range sourceRoots {
		for _, dir := range clientDirs {
			srcDir := filepath.ToSlash(filepath.Join(srcRootName, dir))
			if _, err := fs.Stat(templates, srcDir); err != nil {
				continue
			}

			dstRootRel := dir
			dstRootAbs := filepath.Join(opt.TargetDir, dstRootRel)
			if err := root.MkdirAll(dstRootRel, 0o750); err != nil {
				return fmt.Errorf("mkdir %s: %w", dstRootAbs, err)
			}

			err = fs.WalkDir(templates, srcDir, func(path string, d fs.DirEntry, err error) (retErr error) {
				if err != nil {
					return err
				}
				rel := strings.TrimPrefix(path, srcDir)
				rel = strings.TrimPrefix(rel, "/")
				dst := filepath.Join(dstRootRel, filepath.FromSlash(rel))
				if d.IsDir() {
					return root.MkdirAll(dst, 0o750)
				}
				if err := root.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
					return err
				}
				if !opt.Force {
					if _, err := root.Stat(dst); err == nil {
						return nil
					}
				}
				srcFile, err := templates.Open(path)
				if err != nil {
					return err
				}
				defer func() {
					if cerr := srcFile.Close(); cerr != nil && retErr == nil {
						retErr = cerr
					}
				}()
				out, err := root.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
				if err != nil {
					return err
				}
				defer func() {
					if cerr := out.Close(); cerr != nil && retErr == nil {
						retErr = cerr
					}
				}()
				content, err := io.ReadAll(srcFile)
				if err != nil {
					return err
				}
				content = renderTemplate(path, content, opt)
				if _, err := out.Write(content); err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
	}

	// Keep bootstrap usable in CI builds where assistant template files are
	// intentionally not embedded from Git-tracked assets.
	for relPath, content := range defaultAssistantFiles {
		if err := ensureDefaultFile(root, relPath, content); err != nil {
			return err
		}
	}
	for _, dir := range defaultAssistantDirs {
		if err := root.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	return nil
}

func renderTemplate(srcPath string, content []byte, opt InstallOptions) []byte {
	lower := strings.ToLower(srcPath)
	switch {
	case strings.HasSuffix(lower, ".md"),
		strings.HasSuffix(lower, ".json"),
		strings.HasSuffix(lower, ".txt"),
		strings.HasSuffix(lower, ".yaml"),
		strings.HasSuffix(lower, ".yml"):
	default:
		return content
	}

	repoName := strings.TrimSpace(opt.RepoName)
	if repoName == "" {
		repoName = filepath.Base(opt.TargetDir)
	}
	replacements := map[string]string{
		"{{REPO_NAME}}": repoName,
		"{{REPO_ROOT}}": opt.TargetDir,
	}
	out := string(content)
	for placeholder, value := range replacements {
		out = strings.ReplaceAll(out, placeholder, value)
	}
	return []byte(out)
}

func ensureDefaultFile(root *os.Root, relPath, content string) error {
	if _, err := root.Stat(relPath); err == nil {
		return nil
	}
	if err := root.MkdirAll(filepath.Dir(relPath), 0o750); err != nil {
		return fmt.Errorf("mkdir %s: %w", relPath, err)
	}
	f, err := root.OpenFile(relPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open %s: %w", relPath, err)
	}
	defer func() {
		_ = f.Close()
	}()
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("write %s: %w", relPath, err)
	}
	return nil
}
