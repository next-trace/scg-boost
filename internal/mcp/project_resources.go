package mcp

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

type ProjectResourceOptions struct {
	Root           string
	MaxBytes       int64
	MaxTreeEntries int
}

func RegisterProjectResources(s ToolAdder, opt ProjectResourceOptions, projectSummaryMarkdown string) error {
	if opt.MaxBytes == 0 {
		opt.MaxBytes = 64 * 1024
	}
	if opt.MaxTreeEntries == 0 {
		opt.MaxTreeEntries = 500
	}
	abs, err := filepath.Abs(opt.Root)
	if err != nil {
		return err
	}
	opt.Root = abs

	// scg://project/summary
	if err := s.AddResource(mcp.Resource{Name: "scg://project/summary"}, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{mcp.TextResourceContents{URI: req.Params.URI, MIMEType: "text/markdown", Text: projectSummaryMarkdown}}, nil
	}); err != nil {
		return err
	}

	// scg://project/claude
	if err := registerMarkdownResource(s, "scg://project/claude", opt.Root, filepath.Join(".claude", "CLAUDE.md")); err != nil {
		return err
	}
	// scg://project/codex
	if err := registerMarkdownResource(s, "scg://project/codex", opt.Root, filepath.Join(".codex", "CODEX.md")); err != nil {
		return err
	}
	// scg://project/gemini
	if err := registerMarkdownResource(s, "scg://project/gemini", opt.Root, filepath.Join(".gemini", "GEMINI.md")); err != nil {
		return err
	}

	// scg://project/tree
	if err := s.AddResource(mcp.Resource{Name: "scg://project/tree"}, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		lines := make([]string, 0, 256)
		count := 0
		_ = filepath.WalkDir(opt.Root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if count >= opt.MaxTreeEntries {
				return fs.SkipAll
			}
			name := d.Name()
			if d.IsDir() && (name == ".git" || name == "vendor" || name == ".terraform" || name == "node_modules") {
				return fs.SkipDir
			}
			rel, _ := filepath.Rel(opt.Root, path)
			rel = filepath.ToSlash(rel)
			if rel == "." {
				return nil
			}
			lines = append(lines, rel)
			count++
			return nil
		})
		text := strings.Join(lines, "\n")
		return []mcp.ResourceContents{mcp.TextResourceContents{URI: req.Params.URI, MIMEType: "text/plain", Text: text}}, nil
	}); err != nil {
		return err
	}

	return nil
}

func registerMarkdownResource(s ToolAdder, uri string, rootPath string, relPath string) error {
	return s.AddResource(mcp.Resource{Name: uri}, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		root, err := os.OpenRoot(rootPath)
		if err != nil {
			return nil, fmt.Errorf("open root %s: %w", rootPath, err)
		}
		defer func() {
			_ = root.Close()
		}()

		b, err := root.ReadFile(relPath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", filepath.Join(rootPath, relPath), err)
		}
		return []mcp.ResourceContents{mcp.TextResourceContents{URI: req.Params.URI, MIMEType: "text/markdown", Text: string(b)}}, nil
	})
}
