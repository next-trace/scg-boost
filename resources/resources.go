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

// NOTE: go:embed does not include dot-prefixed paths unless explicitly matched.
// Our templates live under bootstrap_templates/**/.claude/**
//
//go:embed bootstrap_templates/_PACK_README.md
//go:embed bootstrap_templates/_PACK_PLAN.md
//go:embed bootstrap_templates/**/.claude/**
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
