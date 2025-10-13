package project

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Kind string

const (
	KindUnknown   Kind = "unknown"
	KindGoService Kind = "go-service"
	KindGoLibrary Kind = "go-library"
	KindTerraform Kind = "terraform"
	KindShellLib  Kind = "shell"
)

type Summary struct {
	Name        string   `json:"name"`
	Root        string   `json:"root"`
	Kind        Kind     `json:"kind"`
	Entrypoints []string `json:"entrypoints"`
	Signals     []string `json:"signals"`
	Fingerprint string   `json:"fingerprint"`
}

func Detect(root string) (Summary, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Summary{}, err
	}
	name := filepath.Base(abs)
	s := Summary{Name: name, Root: abs, Kind: KindUnknown}

	if exists(filepath.Join(abs, "main.tf")) || exists(filepath.Join(abs, "terraform")) {
		s.Kind = KindTerraform
	}
	if exists(filepath.Join(abs, "go.mod")) {
		// heuristic: cmd/ => service
		if isDir(filepath.Join(abs, "cmd")) {
			s.Kind = KindGoService
		} else {
			s.Kind = KindGoLibrary
		}
	}
	if exists(filepath.Join(abs, "Makefile")) {
		s.Signals = append(s.Signals, "Makefile")
	}
	if exists(filepath.Join(abs, "composer.json")) {
		s.Signals = append(s.Signals, "composer.json")
	}

	// entrypoints (bounded)
	var eps []string
	walkLimit := 2000
	_ = filepath.WalkDir(abs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		walkLimit--
		if walkLimit <= 0 {
			return fs.SkipAll
		}
		if d.IsDir() {
			bn := d.Name()
			if bn == ".git" || bn == "vendor" || bn == ".terraform" {
				return fs.SkipDir
			}
			return nil
		}
		base := filepath.Base(path)
		if base == "main.go" && (strings.Contains(path, string(filepath.Separator)+"cmd"+string(filepath.Separator)) || strings.Contains(path, string(filepath.Separator)+"cmd/")) {
			rel, _ := filepath.Rel(abs, path)
			eps = append(eps, filepath.ToSlash(rel))
		}
		if len(eps) >= 10 {
			return fs.SkipAll
		}
		return nil
	})
	s.Entrypoints = eps

	// fingerprint: cheap dir signature (names of key files)
	h := sha256.New()
	for _, f := range []string{"go.mod", "composer.json", "main.tf", ".claude/CLAUDE.md"} {
		if exists(filepath.Join(abs, f)) {
			_, _ = h.Write([]byte(f))
		}
	}
	s.Fingerprint = hex.EncodeToString(h.Sum(nil))

	return s, nil
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func (s Summary) Markdown() string {
	return fmt.Sprintf("# SCG Project Summary\n\n- Name: %s\n- Root: %s\n- Kind: %s\n- Entrypoints: %s\n- Signals: %s\n- Fingerprint: %s\n", s.Name, s.Root, s.Kind, strings.Join(s.Entrypoints, ", "), strings.Join(s.Signals, ", "), s.Fingerprint)
}
