package bootstrap

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Conflict describes a file overlap between skills.
type Conflict struct {
	Path          string
	ExistingSkill string
	NewSkill      string
	Severity      string
}

type overrideRule struct {
	full     bool
	sections map[string]struct{}
}

// ApplyOverrides applies overrides from .scg/overrides/<skill-id>/ onto .claude/.
// Returns warnings for ignored override files.
func ApplyOverrides(root, skillID string, overridePaths []string) ([]string, error) {
	rules, err := parseOverridePaths(overridePaths)
	if err != nil {
		return nil, err
	}

	overrides, err := loadOverrideFiles(root, skillID)
	if err != nil {
		return nil, err
	}
	if len(overrides) == 0 {
		return nil, nil
	}

	var warnings []string
	for targetPath, content := range overrides {
		rule := rules[targetPath]
		if rule == nil {
			warnings = append(warnings, fmt.Sprintf("override ignored for %s (not in override_paths)", targetPath))
			continue
		}

		targetAbs := filepath.Join(root, filepath.FromSlash(targetPath))
		if rule.full {
			if err := writeFile(targetAbs, content); err != nil {
				return warnings, err
			}
			continue
		}

		if err := applySectionOverrides(root, targetPath, content, rule.sections); err != nil {
			return warnings, err
		}
	}

	return warnings, nil
}

// DetectConflicts checks for overlapping files between installed skills and a new skill.
func DetectConflicts(templates fs.FS, installedSkills []string, newSkill string) ([]Conflict, error) {
	if newSkill == "" {
		return nil, fmt.Errorf("new skill is required")
	}

	existing := make(map[string]string)
	for _, skill := range installedSkills {
		files, err := listSkillFiles(templates, skill)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if _, ok := existing[file]; !ok {
				existing[file] = skill
			}
		}
	}

	newFiles, err := listSkillFiles(templates, newSkill)
	if err != nil {
		return nil, err
	}

	var conflicts []Conflict
	for _, file := range newFiles {
		if owner, ok := existing[file]; ok {
			conflicts = append(conflicts, Conflict{
				Path:          file,
				ExistingSkill: owner,
				NewSkill:      newSkill,
				Severity:      "warn",
			})
		}
	}

	sort.Slice(conflicts, func(i, j int) bool {
		return conflicts[i].Path < conflicts[j].Path
	})
	return conflicts, nil
}

// ExtractSectionBody returns the body of a markdown section by anchor.
// The returned content excludes the heading line.
func ExtractSectionBody(content []byte, section string) ([]byte, error) {
	lines := strings.Split(string(content), "\n")
	sections := parseMarkdownSections(lines)
	base, ok := sections[section]
	if !ok {
		return nil, fmt.Errorf("section %q not found", section)
	}
	if base.start+1 >= base.end {
		return []byte{}, nil
	}
	body := strings.Join(lines[base.start+1:base.end], "\n")
	return []byte(body), nil
}

func listSkillFiles(templates fs.FS, skillName string) ([]string, error) {
	root := path.Join(skillName, ".claude")
	if _, err := fs.Stat(templates, root); err != nil {
		return nil, fmt.Errorf("skill %q not found: %w", skillName, err)
	}

	var files []string
	if err := fs.WalkDir(templates, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(p, skillName+"/")
		files = append(files, path.Join(".", rel))
		return nil
	}); err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}

func parseOverridePaths(paths []string) (map[string]*overrideRule, error) {
	rules := make(map[string]*overrideRule)
	for _, raw := range paths {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		parts := strings.SplitN(raw, "#", 2)
		filePath := path.Clean(parts[0])
		if filePath == "." || strings.HasPrefix(filePath, "../") {
			return nil, fmt.Errorf("invalid override path: %s", raw)
		}
		if !strings.HasPrefix(filePath, ".claude/") && filePath != ".claude" {
			return nil, fmt.Errorf("override path must start with .claude/: %s", raw)
		}

		rule := rules[filePath]
		if rule == nil {
			rule = &overrideRule{
				sections: make(map[string]struct{}),
			}
			rules[filePath] = rule
		}

		if len(parts) == 1 || parts[1] == "" {
			rule.full = true
			continue
		}
		rule.sections[parts[1]] = struct{}{}
	}

	return rules, nil
}

func loadOverrideFiles(root, skillID string) (map[string][]byte, error) {
	overrideRoot := filepath.Join(root, ".scg", "overrides", skillID)
	if _, err := os.Stat(overrideRoot); os.IsNotExist(err) {
		return nil, nil
	}

	overrides := make(map[string][]byte)
	overrideFS := os.DirFS(overrideRoot)
	err := filepath.WalkDir(overrideRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(overrideRoot, p)
		if err != nil {
			return err
		}
		rel = path.Clean(filepath.ToSlash(rel))
		if !fs.ValidPath(rel) {
			return fmt.Errorf("invalid override path: %s", rel)
		}
		target := path.Join(".claude", rel)
		data, err := fs.ReadFile(overrideFS, rel)
		if err != nil {
			return err
		}
		overrides[target] = data
		return nil
	})
	if err != nil {
		return nil, err
	}

	return overrides, nil
}

func applySectionOverrides(root, targetPath string, overrideContent []byte, allowedSections map[string]struct{}) error {
	baseContent, err := readFileFromRoot(root, targetPath)
	if err != nil {
		return fmt.Errorf("read base file: %w", err)
	}

	baseLines := strings.Split(string(baseContent), "\n")
	baseSections := parseMarkdownSections(baseLines)

	overrideLines := strings.Split(string(overrideContent), "\n")
	overrideSections := parseMarkdownSections(overrideLines)

	var replacements []sectionReplacement
	if len(overrideSections) == 0 {
		if len(allowedSections) != 1 {
			return fmt.Errorf("override requires markdown headings for multiple sections")
		}
		var anchor string
		for section := range allowedSections {
			anchor = section
		}
		base, ok := baseSections[anchor]
		if !ok {
			return fmt.Errorf("section %q not found in %s", anchor, filepath.Join(root, filepath.FromSlash(targetPath)))
		}

		heading := baseLines[base.start]
		body := strings.Join(overrideLines, "\n")
		replLines := []string{heading}
		if body != "" {
			replLines = append(replLines, strings.Split(body, "\n")...)
		}
		replacements = append(replacements, sectionReplacement{
			start: base.start,
			end:   base.end,
			lines: replLines,
		})
	} else {
		for anchor, overrideSection := range overrideSections {
			if _, allowed := allowedSections[anchor]; !allowed {
				continue
			}
			base, ok := baseSections[anchor]
			if !ok {
				return fmt.Errorf("section %q not found in %s", anchor, filepath.Join(root, filepath.FromSlash(targetPath)))
			}
			replLines := overrideLines[overrideSection.start:overrideSection.end]
			replacements = append(replacements, sectionReplacement{
				start: base.start,
				end:   base.end,
				lines: replLines,
			})
		}
	}

	if len(replacements) == 0 {
		return nil
	}

	sort.Slice(replacements, func(i, j int) bool {
		return replacements[i].start > replacements[j].start
	})

	for _, repl := range replacements {
		baseLines = append(baseLines[:repl.start], append(repl.lines, baseLines[repl.end:]...)...)
	}

	targetAbs := filepath.Join(root, filepath.FromSlash(targetPath))
	return writeFile(targetAbs, []byte(strings.Join(baseLines, "\n")))
}

func readFileFromRoot(root, rel string) ([]byte, error) {
	clean := path.Clean(filepath.ToSlash(rel))
	if !fs.ValidPath(clean) {
		return nil, fmt.Errorf("invalid path: %s", rel)
	}
	return fs.ReadFile(os.DirFS(root), clean)
}

type sectionReplacement struct {
	start int
	end   int
	lines []string
}

type mdSection struct {
	anchor string
	level  int
	start  int
	end    int
}

func parseMarkdownSections(lines []string) map[string]mdSection {
	type heading struct {
		index  int
		level  int
		anchor string
	}

	var headings []heading
	for i, line := range lines {
		level, text, ok := parseHeading(line)
		if !ok {
			continue
		}
		anchor := anchorFromHeading(text)
		if anchor == "" {
			continue
		}
		headings = append(headings, heading{index: i, level: level, anchor: anchor})
	}

	sections := make(map[string]mdSection)
	for i, h := range headings {
		end := len(lines)
		for j := i + 1; j < len(headings); j++ {
			if headings[j].level <= h.level {
				end = headings[j].index
				break
			}
		}
		if _, exists := sections[h.anchor]; !exists {
			sections[h.anchor] = mdSection{
				anchor: h.anchor,
				level:  h.level,
				start:  h.index,
				end:    end,
			}
		}
	}

	return sections
}

func parseHeading(line string) (int, string, bool) {
	line = strings.TrimLeft(line, " \t")
	if !strings.HasPrefix(line, "#") {
		return 0, "", false
	}
	count := 0
	for count < len(line) && line[count] == '#' {
		count++
	}
	if count == 0 || count > 6 {
		return 0, "", false
	}
	text := strings.TrimSpace(line[count:])
	if text == "" {
		return 0, "", false
	}
	return count, text, true
}

func anchorFromHeading(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	var b strings.Builder
	b.Grow(len(text))
	lastDash := false
	for i := 0; i < len(text); i++ {
		ch := text[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			b.WriteByte(ch)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
