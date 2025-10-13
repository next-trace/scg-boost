package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

func readRepoFile(root, rel string) ([]byte, error) {
	clean := path.Clean(filepath.ToSlash(rel))
	if !fs.ValidPath(clean) {
		return nil, fmt.Errorf("invalid path: %s", rel)
	}
	return fs.ReadFile(os.DirFS(root), clean)
}
