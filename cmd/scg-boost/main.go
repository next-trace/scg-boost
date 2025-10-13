package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	iofs "io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/next-trace/scg-boost/boost"
	"github.com/next-trace/scg-boost/internal/bootstrap"
	"github.com/next-trace/scg-boost/internal/project"
	"github.com/next-trace/scg-boost/internal/skills"
	"github.com/next-trace/scg-boost/resources"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}

	switch args[0] {
	case "install":
		return cmdInstall(args[1:])
	case "config":
		return cmdConfig(args[1:])
	case "mcp", "serve":
		return cmdMCP(args[1:])
	case "scan":
		return cmdScan(args[1:])
	case "tools":
		return cmdTools(args[1:])
	case "version":
		return cmdVersion()
	case "validate":
		return cmdValidate(args[1:])
	case "skills:list":
		return cmdSkillsList(args[1:])
	case "skills:install":
		return cmdSkillsInstall(args[1:])
	case "skills:sync":
		return cmdSkillsSync(args[1:])
	case "skills:override":
		return cmdSkillsOverride(args[1:])
	case "help", "-h", "--help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		usage()
		return 2
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `scg-boost

Laravel Boost-style MCP + context bootstrap for SupplyChainGuard repos.

Usage:
  scg-boost install [--root .] [--repo <name>] [--force]
  scg-boost config --client claude|cursor|junie [--root .] [--name <server>]
  scg-boost scan [--root .]
  scg-boost mcp [--root .] [--name <app>] [--version <v>]
  scg-boost tools [--json]
  scg-boost version
  scg-boost validate [--root .]
  scg-boost skills:list [--format json|table]
  scg-boost skills:install --skill <name> [--root .] [--force]
  scg-boost skills:sync [--root .]
  scg-boost skills:override --skill <name> [--root .] [--path <path>] [--force]`)
}

func cmdInstall(args []string) int {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repo root")
	repoName := fs.String("repo", "", "repo name (defaults to folder name)")
	force := fs.Bool("force", false, "overwrite existing .claude")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	abs, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	name := *repoName
	if name == "" {
		name = filepath.Base(abs)
	}

	tpl, err := resources.BootstrapTemplates()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	// Auto-detect repo type and suggest skill if --repo not specified
	if *repoName == "" {
		repoType := skills.DetectRepoType(abs)
		reg, err := skills.Load(tpl)
		if err == nil && reg.Count() > 0 {
			matches := reg.MatchRepoType(repoType)
			if len(matches) > 0 {
				fmt.Fprintf(os.Stderr, "Detected repo type: %s\n", repoType)
				fmt.Fprintf(os.Stderr, "Suggested skill: %s\n", matches[0].Name)
				name = matches[0].Name
			}
		}
	}

	if err := bootstrap.Install(tpl, bootstrap.InstallOptions{RepoName: name, TargetDir: abs, Force: *force}); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if _, err := fmt.Fprintf(os.Stdout, "Installed .claude for %s in %s\n", name, abs); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}

func cmdConfig(args []string) int {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	client := fs.String("client", "claude", "client: claude|cursor|junie")
	root := fs.String("root", ".", "repo root")
	name := fs.String("name", "", "server name (defaults to folder name)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	abs, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	serverName := *name
	if serverName == "" {
		serverName = filepath.Base(abs)
	}

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	exe, _ = filepath.Abs(exe)
	exe = filepath.Clean(exe)

	switch *client {
	case "claude":
		if _, err := fmt.Fprintf(os.Stdout, `{"mcpServers": {"scg-boost": {"command": %q, "args": ["mcp", "--root", %q, "--name", %q]}}}\n`, exe, abs, serverName); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		return 0
	case "cursor":
		if _, err := fmt.Fprintf(os.Stdout, `{"mcp.servers": {"scg-boost": {"command": %q, "args": ["mcp", "--root", %q, "--name", %q]}}}\n`, exe, abs, serverName); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		return 0
	case "junie":
		if _, err := fmt.Fprintf(os.Stdout, "[tool_provider.scg_boost]\nname = %q\ncommand = %q\nargs = [\"mcp\", \"--root\", %q, \"--name\", %q]\n", "scg-boost", exe, abs, serverName); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown client: %s\n", *client)
		return 2
	}
}

func cmdScan(args []string) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repo root")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	sum, err := project.Detect(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if _, err := fmt.Fprint(os.Stdout, sum.Markdown()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}

func cmdMCP(args []string) int {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repo root")
	name := fs.String("name", "", "server name (defaults to folder name)")
	version := fs.String("version", "0.1.0", "server version")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	abs, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	serverName := *name
	if serverName == "" {
		serverName = filepath.Base(abs)
	}

	sum, err := project.Detect(abs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	srv, err := boost.New(
		boost.WithName(serverName),
		boost.WithVersion(*version),
		boost.WithProjectResources(abs, sum.Markdown()),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if _, err := srv.Start(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	<-ctx.Done()
	return 0
}

func cmdTools(args []string) int {
	fs := flag.NewFlagSet("tools", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	jsonOut := fs.Bool("json", false, "output as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	tools := []map[string]string{
		{"name": "appinfo.get", "description": "Get application info (name, version, Go runtime, uptime)"},
		{"name": "config.get", "description": "Get a configuration value"},
		{"name": "config.list", "description": "List configuration keys with prefix"},
		{"name": "dbquery.run", "description": "Execute a read-only SQL query"},
		{"name": "dbschema.list", "description": "List database schema, tables, and columns"},
		{"name": "logs.lastError", "description": "Get the last error log entry"},
		{"name": "health.status", "description": "Get liveness and readiness status"},
		{"name": "events.outbox.peek", "description": "Peek into the event outbox"},
		{"name": "trace.lookup", "description": "Lookup recent traces"},
		{"name": "service.topology", "description": "Get service topology snapshot"},
		{"name": "routes.list", "description": "List registered HTTP/gRPC routes"},
		{"name": "migrations.status", "description": "Get database migration status"},
		{"name": "cache.stats", "description": "Get cache statistics"},
		{"name": "docs.search", "description": "Search project documentation"},
		{"name": "metrics.summary", "description": "Get metrics summary"},
		{"name": "env.check", "description": "Validate environment configuration"},
	}

	if *jsonOut {
		fmt.Println("[")
		for i, t := range tools {
			comma := ","
			if i == len(tools)-1 {
				comma = ""
			}
			fmt.Printf(`  {"name": %q, "description": %q}%s`+"\n", t["name"], t["description"], comma)
		}
		fmt.Println("]")
	} else {
		fmt.Println("Available tools:")
		for _, t := range tools {
			fmt.Printf("  %-20s %s\n", t["name"], t["description"])
		}
	}
	return 0
}

func cmdVersion() int {
	fmt.Printf("scg-boost version %s\n", boost.Version)
	return 0
}

func cmdValidate(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repo root")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	abs, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	issues := []string{}

	// Check for .claude directory
	claudeDir := filepath.Join(abs, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		issues = append(issues, "missing .claude directory (run 'scg-boost install')")
	} else {
		// Check for CLAUDE.md
		if _, err := os.Stat(filepath.Join(claudeDir, "CLAUDE.md")); os.IsNotExist(err) {
			issues = append(issues, "missing .claude/CLAUDE.md")
		}
		// Check for agents directory
		if _, err := os.Stat(filepath.Join(claudeDir, "agents")); os.IsNotExist(err) {
			issues = append(issues, "missing .claude/agents directory")
		}
		// Check for commands directory
		if _, err := os.Stat(filepath.Join(claudeDir, "commands")); os.IsNotExist(err) {
			issues = append(issues, "missing .claude/commands directory")
		}
	}

	// Check for go.mod (Go project)
	if _, err := os.Stat(filepath.Join(abs, "go.mod")); os.IsNotExist(err) {
		issues = append(issues, "no go.mod found (not a Go project?)")
	}

	if len(issues) == 0 {
		fmt.Println("Validation passed: project setup is valid")
		return 0
	}

	fmt.Println("Validation issues found:")
	for _, issue := range issues {
		fmt.Printf("  - %s\n", issue)
	}
	return 1
}

func cmdSkillsList(args []string) int {
	fs := flag.NewFlagSet("skills:list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	format := fs.String("format", "table", "output format: table|json")
	if err := fs.Parse(args); err != nil {
		return 2
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

	list := reg.List()

	if *format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(list); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		return 0
	}

	// Table format
	fmt.Printf("%-25s %-10s %s\n", "NAME", "VERSION", "DESCRIPTION")
	fmt.Println("---------------------------------------------------------------")
	for _, skill := range list {
		desc := skill.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		fmt.Printf("%-25s %-10s %s\n", skill.Name, skill.Version, desc)
	}
	fmt.Printf("\nTotal: %d skills\n", len(list))
	return 0
}

func cmdSkillsInstall(args []string) int {
	fs := flag.NewFlagSet("skills:install", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	skillName := fs.String("skill", "", "skill name (required)")
	root := fs.String("root", ".", "repo root")
	force := fs.Bool("force", false, "overwrite existing .claude")
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

	// Verify skill exists
	reg, err := skills.Load(tpl)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if !reg.HasSkill(*skillName) {
		fmt.Fprintf(os.Stderr, "error: skill %q not found\n", *skillName)
		fmt.Fprintln(os.Stderr, "Run 'scg-boost skills:list' to see available skills")
		return 1
	}

	meta := reg.Get(*skillName)
	if meta == nil {
		fmt.Fprintf(os.Stderr, "error: skill %q metadata not found\n", *skillName)
		return 1
	}

	state, err := loadInstalledState(abs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: failed to read installed skills:", err)
	}
	if state != nil && len(state.Skills) > 0 {
		conflicts, err := bootstrap.DetectConflicts(tpl, installedSkillIDs(state), meta.ID)
		if err != nil {
			fmt.Fprintln(os.Stderr, "warning: conflict detection failed:", err)
		} else if len(conflicts) > 0 {
			fmt.Fprintln(os.Stderr, "warning: potential skill conflicts detected:")
			for _, conflict := range conflicts {
				fmt.Fprintf(os.Stderr, "  - %s (existing: %s, new: %s)\n", conflict.Path, conflict.ExistingSkill, conflict.NewSkill)
			}
		}
	}

	if err := bootstrap.InstallSkill(tpl, *skillName, bootstrap.InstallOptions{TargetDir: abs, Force: *force}); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	// Write skill.json marker to .claude/ for sync tracking
	skillJSON := filepath.Join(abs, ".claude", "skill.json")
	data, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(skillJSON, data, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to write skill marker: %v\n", err)
	}

	warnings, err := bootstrap.ApplyOverrides(abs, meta.ID, meta.OverridePaths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, "warning:", warning)
	}

	if state != nil {
		upsertInstalledSkill(state, meta, hasOverrides(abs, meta.ID))
		if err := saveInstalledState(abs, state); err != nil {
			fmt.Fprintln(os.Stderr, "warning: failed to write installed skills:", err)
		}
	}

	fmt.Printf("Installed skill %q in %s\n", *skillName, abs)
	return 0
}

func cmdSkillsSync(args []string) int {
	fs := flag.NewFlagSet("skills:sync", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repo root")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	abs, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	var skillID string
	var meta skills.Metadata
	if data, err := readRepoFile(abs, ".claude/skill.json"); err == nil {
		if err := json.Unmarshal(data, &meta); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		skillID = meta.ID
		if skillID == "" {
			skillID = meta.Name
		}
	} else if !errors.Is(err, iofs.ErrNotExist) {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	} else {
		state, err := loadInstalledState(abs)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: no skill.json found in .claude/")
			fmt.Fprintln(os.Stderr, "This directory may not have been installed with skills:install")
			return 1
		}
		if len(state.Skills) != 1 {
			fmt.Fprintln(os.Stderr, "error: unable to determine installed skill")
			fmt.Fprintln(os.Stderr, "This directory may not have been installed with skills:install")
			return 1
		}
		skillID = state.Skills[0].ID
	}

	tpl, err := resources.BootstrapTemplates()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	// Re-install with force=true
	reg, err := skills.Load(tpl)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	current := reg.Get(skillID)
	if current == nil {
		fmt.Fprintf(os.Stderr, "error: skill %q not found\n", skillID)
		return 1
	}

	if err := bootstrap.InstallSkill(tpl, skillID, bootstrap.InstallOptions{TargetDir: abs, Force: true}); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	warnings, err := bootstrap.ApplyOverrides(abs, current.ID, current.OverridePaths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, "warning:", warning)
	}

	// Update skill.json marker
	skillJSON := filepath.Join(abs, ".claude", "skill.json")
	data, _ := json.MarshalIndent(current, "", "  ")
	_ = os.WriteFile(skillJSON, data, 0o600)

	state, err := loadInstalledState(abs)
	if err == nil {
		upsertInstalledSkill(state, current, hasOverrides(abs, current.ID))
		if err := saveInstalledState(abs, state); err != nil {
			fmt.Fprintln(os.Stderr, "warning: failed to write installed skills:", err)
		}
	}

	fmt.Printf("Synced skill %q in %s\n", current.Name, abs)
	return 0
}
