package resources

import (
	"embed"
	"io/fs"
)

//go:embed guidelines/SCG_CODING_GUIDELINES.md
var guidelinesFS embed.FS

//go:embed schemas/dbquery.run.input.json
//go:embed schemas/dbquery.run.output.json
var schemasFS embed.FS

// Assistant-specific dot-directories are intentionally not embedded from
// bootstrap_templates to keep them untracked in Git. The installer provides
// minimal defaults when those templates are absent.
//
//go:embed bootstrap_templates/_PACK_README.md
//go:embed bootstrap_templates/_PACK_PLAN.md
//go:embed bootstrap_templates/*/skill.json
var bootstrapFS embed.FS

func Guidelines() ([]byte, error) {
	return guidelinesFS.ReadFile("guidelines/SCG_CODING_GUIDELINES.md")
}

func GetDbQueryRunInputSchema() ([]byte, error) {
	return schemasFS.ReadFile("schemas/dbquery.run.input.json")
}

func GetDbQueryRunOutputSchema() ([]byte, error) {
	return schemasFS.ReadFile("schemas/dbquery.run.output.json")
}

// BootstrapTemplates returns a filesystem containing the embedded per-repo
// bootstrap templates (CLAUDE.md, agents, commands).
func BootstrapTemplates() (fs.FS, error) {
	return fs.Sub(bootstrapFS, "bootstrap_templates")
}
