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

type InstallOptions struct {
	RepoName  string
	TargetDir string
	Force     bool
}

// Install copies the embedded .claude template set into the target repository.
//
// Rules:
// - Writes only under <TargetDir>/.claude
// - Refuses to overwrite without Force
// - Uses repo-specific templates if present, otherwise falls back to _generic
func Install(templates fs.FS, opt InstallOptions) error {
	if opt.TargetDir == "" {
		return errors.New("target dir is required")
	}
	absTarget, err := filepath.Abs(opt.TargetDir)
	if err != nil {
		return fmt.Errorf("abs target dir: %w", err)
	}
	opt.TargetDir = absTarget
	if opt.RepoName == "" {
		opt.RepoName = filepath.Base(opt.TargetDir)
	}

	// Ensure repo template exists; fallback to _generic
	srcRoot := filepath.ToSlash(filepath.Join(opt.RepoName, ".claude"))
	if _, err := fs.Stat(templates, srcRoot); err != nil {
		srcRoot = "_generic/.claude"
		if _, err2 := fs.Stat(templates, srcRoot); err2 != nil {
			return fmt.Errorf("missing embedded templates: %w", err2)
		}
	}

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

// installFromPath performs the actual installation from srcRoot to opt.TargetDir/.claude.
func installFromPath(templates fs.FS, srcRoot string, opt InstallOptions) error {
	root, err := os.OpenRoot(opt.TargetDir)
	if err != nil {
		return fmt.Errorf("open target root: %w", err)
	}
	defer func() {
		_ = root.Close()
	}()

	dstRootRel := ".claude"
	dstRootAbs := filepath.Join(opt.TargetDir, dstRootRel)
	if st, err := root.Stat(dstRootRel); err == nil && st.IsDir() && !opt.Force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", dstRootAbs)
	}
	if err := root.MkdirAll(dstRootRel, 0o750); err != nil {
		return fmt.Errorf("mkdir %s: %w", dstRootAbs, err)
	}

	return fs.WalkDir(templates, srcRoot, func(path string, d fs.DirEntry, err error) (retErr error) {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(path, srcRoot)
		rel = strings.TrimPrefix(rel, "/")
		dst := filepath.Join(dstRootRel, filepath.FromSlash(rel))
		if d.IsDir() {
			return root.MkdirAll(dst, 0o750)
		}
		if err := root.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
			return err
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
		if _, err := io.Copy(out, srcFile); err != nil {
			return err
		}
		return nil
	})
}
