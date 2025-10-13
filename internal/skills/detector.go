package skills

import (
	"os"
	"path/filepath"
)

// DetectRepoType examines the repository at root and returns the detected type.
// Returns one of: "go-service", "go-library", "terraform", "generic"
func DetectRepoType(root string) string {
	// Check for Go module
	if hasFile(root, "go.mod") {
		// Service if has cmd/ directory, otherwise library
		if hasDirectory(root, "cmd") {
			return "go-service"
		}
		return "go-library"
	}

	// Check for Terraform
	if hasFilesWithExt(root, ".tf") {
		return "terraform"
	}

	return "generic"
}

// hasFile returns true if the file exists at root/name.
func hasFile(root, name string) bool {
	path := filepath.Join(root, name)
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// hasDirectory returns true if the directory exists at root/name.
func hasDirectory(root, name string) bool {
	path := filepath.Join(root, name)
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// hasFilesWithExt returns true if any files with the given extension exist in root.
func hasFilesWithExt(root, ext string) bool {
	entries, err := os.ReadDir(root)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ext {
			return true
		}
	}
	return false
}
